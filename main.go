package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"bufio"

	"github.com/chzyer/readline"
	"github.com/informationsea/shellflow/flowscript"
)

func main() {
	if len(os.Args) <= 1 {
		helpMode([]string{})
		return
	}

	var err error

	switch os.Args[1] {
	case "flowscript":
		runFlowscriptIntepreter()
	case "run":
		err = runMode()
	case "dot":
		err = dotMode()
	case "filelog":
		err = fileLogMode()
	case "viewlog":
		err = viewLogMode()
	case "-h", "-?", "help":
		helpMode(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func helpMode(args []string) error {
	_, err := fmt.Print(`shellflow @DEV@: shell-script like workflow management system

Commands:
  run         Run workflow
  dot         Export workflow as dot language for visualization
  flowscript  Launch flowscript interpreter
  viewlog     Show execution log
  filelog     Create a file log file, which contains SHA256 hash, modification date and so on
  help        Show this help
`)
	return err
}

func viewLogMode() error {
	f := flag.NewFlagSet("shellflow filelog", flag.ExitOnError)
	var showAll bool
	var failedOnly bool
	f.BoolVar(&showAll, "all", false, "Show All")
	f.BoolVar(&failedOnly, "failed", false, "Show Failed Job Only")
	f.Parse(os.Args[2:])

	var err error
	if len(f.Args()) > 0 {
		err = ViewLogDetail(f.Args(), failedOnly)
	} else {
		err = ViewLog(showAll, failedOnly)
	}
	return err
}

func fileLogMode() error {
	f := flag.NewFlagSet("shellflow filelog", flag.ExitOnError)
	var output string
	f.StringVar(&output, "output", "", "output json file")
	f.Parse(os.Args[2:])
	var writer io.WriteCloser
	var err error

	files, err := CreateFileLog(f.Args(), false, MaximumContentLogSize)
	if err != nil {
		return err
	}

	if output == "" || output == "-" {
		writer = os.Stdout
	} else {
		writer, err = os.OpenFile(output, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}
	defer writer.Close()

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(files)
	if err != nil {
		return err
	}
	return nil
}

func loadParameter(paramFile string) (map[string]interface{}, error) {
	parameters := make(map[string]interface{})

	if paramFile != "" {
		if strings.HasSuffix(paramFile, ".json") {

			paramReader, err := os.Open(paramFile)
			if err != nil {
				return nil, err
			}
			decoder := json.NewDecoder(paramReader)
			err = decoder.Decode(&parameters)
			if err != nil {
				return nil, err
			}
			//fmt.Println(parameters)

		} else {
			return nil, fmt.Errorf("Unknown file type %s", paramFile)
		}
	}
	return parameters, nil
}

func dotMode() error {
	paramFile := ""

	f := flag.NewFlagSet("shellflow dot", flag.ExitOnError)
	f.StringVar(&paramFile, "param", "", "Parameter File")
	f.Parse(os.Args[2:])

	if len(f.Args()) != 1 {
		helpMode([]string{"dot"})
		return fmt.Errorf("No workflow file")
	}

	env := NewEnvironment()
	parameters, err := loadParameter(paramFile)
	if err != nil {
		return err
	}

	builder, err := parse(env, f.Args()[0], parameters)
	if err != nil {
		return err
	}
	dag := builder.CreateDag()
	fmt.Println(dag)
	return nil
}

func runMode() error {
	f := flag.NewFlagSet("shellflow run", flag.ExitOnError)

	useSge := false
	paramFile := ""

	env := NewEnvironment()
	f.BoolVar(&env.skipSha, "skip-sha", false, "Skip SHA256 calculation")
	f.BoolVar(&env.dryRun, "dry-run", false, "Print jobs to run without execute")
	f.BoolVar(&env.scriptsOnly, "scripts-only", false, "Generate scripts only")
	f.BoolVar(&env.rerunAll, "rerun", false, "Rerun all commands even if contents are not changed")
	f.BoolVar(&useSge, "sge", false, "Use SGE/UGE instead of local executer")
	f.StringVar(&paramFile, "param", "", "Parameter File")
	f.Parse(os.Args[2:])

	if len(f.Args()) == 0 {
		helpMode([]string{"run"})
		return fmt.Errorf("No workflow file")
	}

	parameters, err := loadParameter(paramFile)
	if err != nil {
		return err
	}

	builder, err := parse(env, f.Args()[0], parameters)
	if err != nil {
		return err
	}
	//fmt.Printf("%s\n", f.Args())

	if env.dryRun {
		for _, v := range builder.Tasks {
			if !v.ShouldSkip || env.rerunAll {
				fmt.Printf("%s\n", v.ShellScript)
			}
		}
		return nil
	}

	if env.rerunAll {
		for _, v := range builder.Tasks {
			v.ShouldSkip = false
		}
	}

	gen, err := GenerateTaskScripts(f.Args()[0], paramFile, env, builder)
	if err != nil {
		return err
	}

	if !env.scriptsOnly {
		if useSge {
			err = ExecuteInSge(gen)
		} else {
			err = ExecuteLocalSingle(gen)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			exeErr, ok := err.(*executationError)
			if ok {
				jobLog, err := CollectLogsForOneJob(exeErr.jobRoot, exeErr.shellTask)
				var errReader io.ReadCloser
				if jobLog.ScriptExitCode == 0 {
					errReader, err = os.Open(path.Join(exeErr.jobRoot, "run.stderr"))
				} else {
					errReader, err = os.Open(path.Join(exeErr.jobRoot, "script.stderr"))
				}
				if err != nil {
					return err
				}
				defer errReader.Close()
				io.Copy(os.Stderr, errReader)

				os.Exit(1)
			} else {
				return err
			}
		}
	}
	return nil
}

//func setupRoot() {
//	rootInfo, err := os.Stat(*workflowRoot)
//	if err != nil && os.IsNotExist(err) {
//		os.MkdirAll(*workflowRoot, 0750)
//	} else if err != nil || !rootInfo.IsDir() {
//		panic(fmt.Sprintf("workflow root directory path is not directory: %s", *workflowRoot))
//	}
//}

func parse(env *Environment, file string, param map[string]interface{}) (*ShellTaskBuilder, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	builder, err := ParseShellflow(bufio.NewReader(reader), env, param)

	return builder, err
}

func runFlowscriptIntepreter() {
	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}

	defer rl.Close()

	ge := flowscript.NewGlobalEnvironment()

	for {
		line, err := rl.Readline()
		if err != nil || line == "exit" {
			break
		}

		ev, err := flowscript.EvaluateScript(line, ge)

		if err == nil {
			fmt.Printf("%s\n", ev.String())
		} else {
			fmt.Printf("Error: %s\n", err.Error())
		}

	}
}
