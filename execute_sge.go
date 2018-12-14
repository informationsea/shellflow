package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const sgeTaskIDFileName = "sge-taskid.txt"

func FollowUpSge(jobLogRoot string) (bool, error) {
	rc, err := os.Open(path.Join(jobLogRoot, "rc"))
	if err == nil {
		defer rc.Close()
		return false, nil
	} else if os.IsNotExist(err) {
		sgeTaskIDFile, err := os.Open(path.Join(jobLogRoot, sgeTaskIDFileName))
		if err == nil {
			var sgeTaskID int
			i, err := fmt.Fscanf(sgeTaskIDFile, "%d", &sgeTaskID)
			if err != nil {
				return false, err
			}
			if i != 1 {
				return false, fmt.Errorf("Cannot read pid: %d", i)
			}

			cmd := exec.Command("qstat", "-j", fmt.Sprintf("%d", sgeTaskID))
			err = cmd.Run()

			if err != nil {
				rc, err = os.OpenFile(path.Join(jobLogRoot, "rc"), os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return false, err
				}
				defer rc.Close()
				_, err = fmt.Fprintf(rc, "1000")
				if err != nil {
					return false, err
				}
			}
			return true, nil
		} else if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return false, err
	}
}

var jobNameReplace = regexp.MustCompile("[^\\w_\\-.]")

func ExecuteInSge(ge *TaskScripts) error {
	sgeTaskID := make(map[int]string)

	jobNameBase := jobNameReplace.ReplaceAllString(ge.jobName, "_")

	for _, v := range ge.builder.Tasks {
		if v.ShouldSkip {
			fmt.Printf("skipping: %s\n", v.ShellScript)
			continue
		}

		if v.CommandConfiguration.RunImmediate {
			err := ExecuteLocalSingleOneTask(ge, v)
			if err != nil {
				return err
			}
			continue
		}

		scriptInfo := ge.scripts[v.ID]
		qsub := []string{"-wd", scriptInfo.JobRoot, "-terse", "-o", path.Join(scriptInfo.JobRoot, "run.stdout"), "-e", path.Join(scriptInfo.JobRoot, "run.stderr")}

		var holdID strings.Builder
		first := true
		for _, d := range v.DependentTaskID {
			if u, ok := sgeTaskID[d]; ok && strings.TrimSpace(u) != "" {
				if first {
					first = false
				} else {
					holdID.WriteString(",")
				}
				holdID.WriteString(strings.TrimSpace(u))
			}
		}

		if holdID.Len() > 0 {
			qsub = append(qsub, "-hold_jid", holdID.String())
		}

		qsub = append(qsub, "-N", "sf-"+jobNameBase+"__ID-"+strconv.Itoa(v.ID))

		if len(v.CommandConfiguration.SGEOption) > 0 {
			qsub = append(qsub, v.CommandConfiguration.SGEOption...)
		}

		qsub = append(qsub, scriptInfo.RunScriptPath)

		// write SGE options
		submitArgs, err := os.OpenFile(path.Join(scriptInfo.JobRoot, "sge-submit-args.txt"), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("cannot write SGE option log: %s", err.Error())
		}
		defer submitArgs.Close()
		for _, v := range qsub {
			submitArgs.WriteString(v)
			submitArgs.WriteString("\n")
		}

		// run SGE
		cmd := exec.Command("qsub", qsub...)

		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("cannot run qsub successfully: %s", err.Error())
		}
		currentTaskID := strings.TrimSpace(string(out))
		sgeTaskID[v.ID] = currentTaskID

		taskid, err := os.OpenFile(path.Join(scriptInfo.JobRoot, "sge-taskid.txt"), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("Cannot open SGE task ID log file: %s", err.Error())
		}
		defer taskid.Close()
		fmt.Fprintf(taskid, "%s\n", currentTaskID)
		fmt.Printf("Submit ID:%s  : %s\n", currentTaskID, v.ShellScript)
	}

	return nil
}
