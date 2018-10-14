package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type GeneratedScript struct {
	JobRoot       string
	StdoutPath    string
	StderrPath    string
	RunScriptPath string
	ScriptPath    string
	Skip          bool
}

type TaskScripts struct {
	workflowRoot string
	jobName      string
	scripts      map[int]*GeneratedScript
	env          *Environment
	builder      *ShellTaskBuilder
}

type Execute func(ge *TaskScripts) error
type FollowUp func(log *JobLog) error

type WorkflowMetaData struct {
	Env           map[string]string
	Shellflow     string
	Args          []string
	WorkDir       string
	Date          time.Time
	User          *user.User
	Workflow      string
	WorkflowPath  string
	Tasks         []*ShellTask
	Parameters    map[string]interface{}
	ParameterFile string
}

var jobNameRegexp = regexp.MustCompile("(\\w+).*")

func GenerateTaskScripts(scriptPath string, paramPath string, env *Environment, builder *ShellTaskBuilder) (*TaskScripts, error) {
	originalWorkDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	err = os.Chdir(env.workDir)
	if err != nil {
		return nil, err
	}
	defer os.Chdir(originalWorkDir)

	basename := path.Base(scriptPath)
	uuidObj := uuid.New()
	workflowDir := path.Join(env.workflowRoot, fmt.Sprintf("%s-%s-%s", time.Now().Format("20060102-150405.000"), basename, uuidObj))
	os.MkdirAll(workflowDir, 0755)

	var shellflowPath string
	_, err = Stat(os.Args[0])
	if os.IsNotExist(err) {
		shellflowPath, err = exec.LookPath(os.Args[0])
		if err != nil {
			return nil, fmt.Errorf("Cannot find shellflow binary absolute path: %s", err.Error())
		}
	} else if err == nil {
		shellflowPath = Abs(os.Args[0])
	} else {
		return nil, fmt.Errorf("Cannot find shellflow binary absolute path: %s", err.Error())
	}

	jobName := path.Base(scriptPath)
	if paramPath != "" {
		jobName += " " + path.Base(paramPath)
	}

	ret := TaskScripts{
		workflowRoot: workflowDir,
		jobName:      jobName,
		scripts:      make(map[int]*GeneratedScript),
		env:          env,
		builder:      builder,
	}

	{
		// create input file list
		fileList, err := os.OpenFile(path.Join(workflowDir, "input.json"), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		defer fileList.Close()

		encoder := json.NewEncoder(fileList)
		encoder.SetIndent("", "  ")
		files, err := CreateFileLog(builder.MissingCreatorFiles.Array(), false, MaximumContentLogSize)
		if err != nil {
			return nil, err
		}
		err = encoder.Encode(files)
		if err != nil {
			return nil, err
		}

		//		for _, x := range files {
		//			tmp, _ := json.MarshalIndent(x, "", "  ")
		//			changed, _ := x.IsChanged()
		//			fmt.Printf("input files: %s %v\n", tmp, changed)
		//		}
	}

	pathEnv := os.ExpandEnv("${PATH}")
	ldLibraryPathEnv := os.ExpandEnv("${LD_LIBRARY_PATH}")

	{
		// create runtime information
		runtimeFile, err := os.OpenFile(path.Join(workflowDir, "runtime.json"), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		defer runtimeFile.Close()

		envMap := make(map[string]string)
		envMap["PATH"] = pathEnv
		envMap["LD_LIBRARY_PATH"] = ldLibraryPathEnv

		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		user, err := user.Current()
		if err != nil {
			return nil, err
		}

		var absParamPath string
		if paramPath != "" {
			absParamPath = Abs(paramPath)
		}

		runtime := WorkflowMetaData{
			Env:           envMap,
			Shellflow:     os.Args[0],
			Args:          os.Args,
			WorkDir:       wd,
			Date:          time.Now(),
			User:          user,
			Workflow:      builder.WorkflowContent,
			WorkflowPath:  Abs(scriptPath),
			Tasks:         builder.Tasks,
			Parameters:    env.parameters,
			ParameterFile: absParamPath,
		}

		encoder := json.NewEncoder(runtimeFile)
		encoder.SetIndent("", "  ")
		encoder.Encode(runtime)
	}

	{
		// create job scripts
		for _, v := range builder.Tasks {
			jobDir := path.Join(workflowDir, fmt.Sprintf("job%03d", v.ID))
			err := os.MkdirAll(jobDir, 0755)
			if err != nil {
				return nil, err
			}

			absScriptPath := Abs(path.Join(jobDir, "script.sh"))
			absRunScriptPath := Abs(path.Join(jobDir, "run.sh"))
			absStdoutPath := Abs(path.Join(jobDir, "script.stdout"))
			absStderrPath := Abs(path.Join(jobDir, "script.stderr"))
			absResultPath := Abs(path.Join(jobDir, "rc"))
			absInputPath := Abs(path.Join(jobDir, "input.json"))
			absOutputPath := Abs(path.Join(jobDir, "output.json"))

			if v.ShouldSkip {
				copyFiles := []string{"script.sh", "run.sh", "script.stdout", "script.stderr", "rc", "input.json", "output.json"}
				for _, x := range copyFiles {
					srcFile, err := os.Open(path.Join(v.ReuseLog.JobLogRoot, x))
					if err != nil {
						return nil, err
					}
					defer srcFile.Close()
					dstFile, err := os.OpenFile(path.Join(jobDir, x), os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						return nil, err
					}
					defer dstFile.Close()
					_, err = io.Copy(dstFile, srcFile)
					if err != nil {
						return nil, err
					}
				}
				rel, err := filepath.Rel(Abs(jobDir), Abs(v.ReuseLog.JobLogRoot))
				if err != nil {
					return nil, err
				}
				err = os.Symlink(rel, path.Join(jobDir, "original"))
				if err != nil {
					return nil, err
				}
			} else {

				scriptFile, err := os.OpenFile(path.Join(jobDir, "script.sh"), os.O_CREATE|os.O_WRONLY, 0755)
				if err != nil {
					return nil, err
				}
				defer scriptFile.Close()
				_, err = scriptFile.WriteString(v.ShellScript)
				if err != nil {
					return nil, err
				}

				var absDependentFilesBuilder strings.Builder
				for _, v := range v.DependentFiles.Array() {
					//str := Abs(v)
					absDependentFilesBuilder.WriteString(strconv.Quote(v))
					absDependentFilesBuilder.WriteString(" ")
				}
				absDependentFiles := absDependentFilesBuilder.String()

				var absCreatingFilesBuilder strings.Builder
				for _, v := range v.CreatingFiles.Array() {
					//str := Abs(v)
					absCreatingFilesBuilder.WriteString(strconv.Quote(v))
					absCreatingFilesBuilder.WriteString(" ")
				}
				absCreatingFiles := absCreatingFilesBuilder.String()

				runFile, err := os.OpenFile(path.Join(jobDir, "run.sh"), os.O_CREATE|os.O_WRONLY, 0755)
				if err != nil {
					return nil, err
				}
				defer runFile.Close()
				fmt.Fprintf(runFile, `#/bin/bash
#set -x
cd %s
`, env.workDir)

				if !v.CommandConfiguration.DontInheirtPath {
					fmt.Fprintf(runFile, `#/bin/bash
export PATH="%s"
export LD_LIBRARY_PATH="%s"
`, pathEnv, ldLibraryPathEnv)
				}

				skipSha := ""
				if env.skipSha {
					skipSha = " -skipSha "
				}

				fmt.Fprintf(runFile, "%s filelog %s -output %s %s || exit 1\n", shellflowPath, skipSha, absInputPath, absDependentFiles)

				fmt.Fprintf(runFile, `/bin/bash -o pipefail -e "%s" > %s 2> %s
EXIT_CODE=$?
`, absScriptPath, absStdoutPath, absStderrPath)

				skipSha = ""
				if env.skipSha {
					skipSha = " -skipSha "
				}
				fmt.Fprintf(runFile, "%s filelog %s -output %s %s || exit 1\n", shellflowPath, skipSha, absOutputPath, absCreatingFiles)
				fmt.Fprintf(runFile, "echo $EXIT_CODE > \"%s\"\n", absResultPath)
				fmt.Fprintf(runFile, "exit $EXIT_CODE\n")

			}

			ret.scripts[v.ID] = &GeneratedScript{
				JobRoot:       jobDir,
				StdoutPath:    absStdoutPath,
				StderrPath:    absStderrPath,
				ScriptPath:    absScriptPath,
				RunScriptPath: absRunScriptPath,
				Skip:          v.ShouldSkip,
			}
		}
	}

	// print workflow directory
	{
		cwd, err := os.Getwd()
		if err != nil {
			return &ret, nil
		}
		r, err := filepath.Rel(cwd, workflowDir)
		if err != nil {
			return &ret, nil
		}
		fmt.Printf("Workflow Log: %s\n", r)
	}

	return &ret, nil
}

func Abs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return abs
}
