package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"syscall"
)

var localRunPidFile = "local-run-pid.txt"

type executationError struct {
	message   string
	exitCode  int
	jobRoot   string
	shellTask *ShellTask
}

func (s *executationError) Error() string {
	return s.message
}

func IsExecutionError(err error) bool {
	_, ok := err.(*executationError)
	return ok
}

func FollowUpLocalSingle(jobLogRoot string) (bool, error) {
	rc, err := os.Open(path.Join(jobLogRoot, "rc"))
	if err == nil {
		defer rc.Close()
		return false, nil
	} else if os.IsNotExist(err) {
		pidFile, err := os.Open(path.Join(jobLogRoot, localRunPidFile))
		if err == nil {
			var pid int
			i, err := fmt.Fscanf(pidFile, "%d", &pid)
			if err != nil {
				return false, err
			}
			if i != 1 {
				return false, fmt.Errorf("Cannot read pid: %d", i)
			}
			process, err := os.FindProcess(pid)
			if err != nil {
				return false, err
			}
			err = process.Signal(syscall.Signal(0))
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

// ExecuteLocalSingle runs tasks in local machine and single thread
func ExecuteLocalSingle(ge *TaskScripts) error {
	originalWorkDir, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(ge.env.workDir)
	if err != nil {
		return err
	}
	defer os.Chdir(originalWorkDir)

	var finalErr error

	for _, v := range ge.builder.Tasks {
		if finalErr != nil {
			scriptInfo := ge.scripts[v.ID]
			rc, err := os.OpenFile(path.Join(scriptInfo.JobRoot, "rc"), os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer rc.Close()
			fmt.Fprintf(rc, "2000")
			continue
		}

		if v.ShouldSkip {
			fmt.Printf("skipping: %s\n", v.ShellScript)
			continue
		}

		err := ExecuteLocalSingleOneTask(ge, v)
		if err != nil {
			finalErr = err
		}
	}

	return finalErr
}

func ExecuteLocalSingleOneTask(ge *TaskScripts, v *ShellTask) error {
	scriptInfo := ge.scripts[v.ID]
	args := []string{scriptInfo.RunScriptPath}

	fmt.Printf("%s\n", v.ShellScript)

	cmd := exec.Command("/bin/bash", args...)

	stdout, err := os.OpenFile(path.Join(scriptInfo.JobRoot, "run.stdout"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer stdout.Close()

	stderr, err := os.OpenFile(path.Join(scriptInfo.JobRoot, "run.stderr"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer stderr.Close()

	pid, err := os.OpenFile(path.Join(scriptInfo.JobRoot, localRunPidFile), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer pid.Close()

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	if _, err = fmt.Fprintf(pid, "%d", cmd.Process.Pid); err != nil {
		return err
	}

	if _, err = io.Copy(stdout, stdoutPipe); err != nil {
		return err
	}

	if _, err = io.Copy(stderr, stderrPipe); err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		exitCode := 1000
		status, ok := cmd.ProcessState.Sys().(syscall.WaitStatus)
		if ok {
			exitCode = status.ExitStatus()
		}

		return &executationError{
			message:   fmt.Sprintf("exit status %d: %s", exitCode, scriptInfo.RunScriptPath),
			exitCode:  exitCode,
			jobRoot:   scriptInfo.JobRoot,
			shellTask: v,
		}
	}

	return nil
}
