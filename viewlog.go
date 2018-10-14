package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/informationsea/shellflow/flowscript"
)

//go:generate stringer -type=WorkflowState
type WorkflowState int

const (
	WorkflowDone WorkflowState = iota
	WorkflowRunning
	WorkflowFailed
	WorkflowUnknown
)

//go:generate stringer -type=JobState
type JobState int

const (
	JobDone JobState = iota
	JobRunning
	JobFailed
	JobPending
	JobUnknown
)

type WorkflowLog struct {
	WorkflowLogRoot string
	WorkflowScript  string
	ParameterFile   string
	StartDate       time.Time
	ChangedInput    []string
	JobLogs         []*JobLog
}

func (v *WorkflowLog) String() string {
	return fmt.Sprintf(`WorkflowLog{
  WorkflowScript: %s,
  WorkflowLogRoot: %s,
  ChangedInput: %s,
  JobLogs: %s,
}`, strconv.Quote(v.WorkflowScript), strconv.Quote(v.WorkflowLogRoot), v.ChangedInput, v.JobLogs)
}

func (v *WorkflowLog) IsDone() bool {
	for _, x := range v.JobLogs {
		if !x.IsDone {
			return false
		}
	}
	return true
}

func (v *WorkflowLog) State() WorkflowState {
	for _, x := range v.JobLogs {
		if x.ExitCode > 0 {
			return WorkflowFailed
		}
	}

	for _, x := range v.JobLogs {
		if x.ExitCode < 0 {
			return WorkflowRunning
		}
	}

	return WorkflowDone
}

func (v *WorkflowLog) IsChanged() bool {
	for _, x := range v.JobLogs {
		if x.IsAnyInputChanged || x.IsAnyOutputChanged {
			return true
		}
	}
	return false
}

func BoolToYesNo(x bool) string {
	if x {
		return "Yes"
	}
	return "No"
}

func (v *WorkflowLog) Summary(failedOnly bool) string {
	var buf = bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "Workflow Script Path: %s\n", v.WorkflowScript)
	fmt.Fprintf(buf, "   Workflow Log Path: %s\n", v.WorkflowLogRoot)
	fmt.Fprintf(buf, "           Job Start: %s\n", v.StartDate.Format("2006/01/02 15:04:05"))
	fmt.Fprintf(buf, " Changed Input Files:")
	for _, x := range v.ChangedInput {
		fmt.Fprintf(buf, " %s", x)
	}
	fmt.Fprint(buf, "\n")

	for _, j := range v.JobLogs {
		if j.State() != JobFailed && failedOnly {
			continue
		}

		fmt.Fprintf(buf, "---- Job: %d ------------\n", j.ShellTask.ID)
		fmt.Fprintf(buf, "             State: %s\n", j.State().String())
		if j.ExitCode >= 0 {
			fmt.Fprintf(buf, "         Exit code: %d\n", j.ExitCode)
		}
		fmt.Fprintf(buf, "          Reusable: %s\n", BoolToYesNo(j.IsReusable()))
		fmt.Fprintf(buf, "            Script: %s\n", j.ShellTask.ShellScript)
		fmt.Fprintf(buf, "             Input:")
		for _, x := range j.ShellTask.DependentFiles.Array() {
			fmt.Fprintf(buf, " %s", x)
		}
		fmt.Fprint(buf, "\n")
		fmt.Fprintf(buf, "            Output:")
		for _, x := range j.ShellTask.CreatingFiles.Array() {
			fmt.Fprintf(buf, " %s", x)
		}
		fmt.Fprint(buf, "\n")

		fmt.Fprintf(buf, " Dependent Job IDs:")
		for _, x := range j.ShellTask.DependentTaskID {
			fmt.Fprintf(buf, " %d", x)
		}
		fmt.Fprint(buf, "\n")

		if j.SgeTaskID != "" {
			fmt.Fprintf(buf, "       SGE Task ID: %s\n", strings.TrimSpace(j.SgeTaskID))
		}

		fmt.Fprintf(buf, "     Log directory: %s\n", j.JobLogRoot)

		if j.State() == JobFailed {
			logfile, err := os.Open(path.Join(j.JobLogRoot, "script.stderr"))
			if err == nil {
				fmt.Fprintf(buf, "  - - - - - - Stderr - - - - - -\n")
				scanner := bufio.NewScanner(logfile)
				scanner.Split(bufio.ScanLines)
				for i := 0; i < 3 && scanner.Scan(); i++ {
					fmt.Fprintf(buf, "  %s\n", scanner.Text())
				}
			}
		}
	}

	return string(buf.Bytes())
}

type JobLog struct {
	JobLogRoot         string
	InputFiles         []FileLog
	OutputFiles        []FileLog
	IsStarted          bool
	IsAnyInputChanged  bool
	IsDone             bool
	IsAnyOutputChanged bool
	ExitCode           int
	ScriptExitCode     int
	ShellTask          *ShellTask
	SgeTaskID          string
}

func (v *JobLog) String() string {
	return fmt.Sprintf(`  JobLog{
    JobLogRoot:         %s,
    InputFiles:         %s,
    OutputFiles:        %s,
    IsReusable:         %v,
    IsStarted:          %v,
    IsAnyInputChanged:  %v,
    IsDone:             %v,
    IsAnyOutputChanged: %v,
    ExitCode:           %d,
    ScriptExitCode:     %d,
    ShellTask:          %s
  }
`, strconv.Quote(v.JobLogRoot), v.InputFiles, v.OutputFiles, v.IsReusable(), v.IsStarted, v.IsAnyInputChanged, v.IsDone, v.IsAnyOutputChanged, v.ExitCode, v.ScriptExitCode, v.ShellTask.String())
}

func (v *JobLog) State() JobState {
	if v.IsDone && v.ExitCode == 0 {
		return JobDone
	} else if v.IsDone {
		return JobFailed
	} else if v.IsStarted {
		return JobRunning
	} else {
		return JobPending
	}
}

func (v *JobLog) IsReusable() bool {
	return v.IsDone && v.IsStarted && !v.IsAnyInputChanged && !v.IsAnyOutputChanged && v.ExitCode == 0
}

func LoadJsonFromFile(path string, obj interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	return decoder.Decode(obj)
}

func CollectLogsForOneJob(jobRoot string, oneTask *ShellTask) (*JobLog, error) {
	// check return code file
	var jobDone = true
	var exitCode = -1
	var scriptExitCode = -1
	var exitCodeFile io.Reader
	var err error
	exitCodeFile, err = os.Open(path.Join(jobRoot, "rc"))
	if os.IsNotExist(err) {
		followUpDone, err := FollowUpLocalSingle(jobRoot)
		if err != nil {
			return nil, fmt.Errorf("Failed to follow up %s : %s", jobRoot, err.Error())
		}
		if !followUpDone {
			followUpDone, err = FollowUpSge(jobRoot)
			if err != nil {
				return nil, fmt.Errorf("Failed to follow up %s : %s", jobRoot, err.Error())
			}
		}
		if !followUpDone {
			exitCodeFile = bytes.NewReader([]byte("1000"))
		} else {
			exitCodeFile, err = os.Open(path.Join(jobRoot, "rc"))
		}
	}

	if err == nil {
		data, err := ioutil.ReadAll(exitCodeFile)
		if err != nil {
			return nil, err
		}
		exitCode64, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 32)
		if err != nil {
			return nil, err
		}
		exitCode = int(exitCode64)
		scriptExitCode = exitCode
	} else if os.IsNotExist(err) {
		jobDone = false
	} else {
		jobDone = false
		return nil, err
	}

	// check input files
	var inputFiles []FileLog
	var jobStarted bool
	var anyInputChanged = false
	err = LoadJsonFromFile(path.Join(jobRoot, "input.json"), &inputFiles)
	if err == nil {
		jobStarted = true
		for _, v := range inputFiles {
			//inputFiles[i].Content = nil
			chg, err := v.IsChanged()
			if err != nil {
				return nil, err
			}
			if chg {
				anyInputChanged = true
			}
		}
	} else if os.IsNotExist(err) {
		jobStarted = false
	} else {
		fmt.Fprintf(os.Stderr, "Failed to load %s\n", err.Error())
		jobStarted = false
	}

	// check output files
	var outputFiles []FileLog
	var anyOutputChanged = false
	var outputFound = false
	err = LoadJsonFromFile(path.Join(jobRoot, "output.json"), &outputFiles)
	if err == nil {
		for _, v := range outputFiles {
			chg, err := v.IsChanged()
			//outputFiles[i].Content = nil
			if err != nil {
				return nil, err
			}
			if chg {
				anyOutputChanged = true
			}
		}
		outputFound = true
	} else if os.IsNotExist(err) {
		outputFound = false
	} else {
		fmt.Fprintf(os.Stderr, "Failed to load %s\n", err.Error())
		outputFound = false
	}

	if jobDone && exitCode == 0 && !outputFound {
		exitCode = 1000
	}

	// check SGE task id
	var sgeTaskID string
	sgeTaskIDFile, err := os.Open(path.Join(jobRoot, "sge-taskid.txt"))
	if err == nil {
		defer sgeTaskIDFile.Close()
		data, err := ioutil.ReadAll(sgeTaskIDFile)
		if err != nil {
			return nil, err
		}
		sgeTaskID = string(data)
	} else if os.IsNotExist(err) {
		// ignore
	} else {
		return nil, err
	}

	return &JobLog{
		JobLogRoot:         jobRoot,
		InputFiles:         inputFiles,
		OutputFiles:        outputFiles,
		IsStarted:          jobStarted,
		IsAnyInputChanged:  anyInputChanged,
		IsDone:             jobDone,
		IsAnyOutputChanged: anyOutputChanged,
		ExitCode:           exitCode,
		ScriptExitCode:     scriptExitCode,
		ShellTask:          oneTask,
		SgeTaskID:          sgeTaskID,
	}, nil
}

const workflowLogCacheFileName = "workflowLogCache.json.gz"

func CollectLogsForOneWorkFromCache(logdirPath string) (*WorkflowLog, error) {
	cacheFile, err := os.OpenFile(path.Join(logdirPath, workflowLogCacheFileName), os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer cacheFile.Close()
	cacheFileUngzip, err := gzip.NewReader(cacheFile)
	if err != nil {
		return nil, err
	}
	jsonDecoder := json.NewDecoder(cacheFileUngzip)

	var workflowLog WorkflowLog
	err = jsonDecoder.Decode(&workflowLog)
	if err != nil {
		return nil, err
	}

	var dependentFiles []FileLog
	err = LoadJsonFromFile(path.Join(logdirPath, "input.json"), &dependentFiles)
	if err != nil {
		return nil, err
	}

	changedInput := make([]string, 0)
	for _, oneDependent := range dependentFiles {
		changed, err := oneDependent.IsChanged()
		if err != nil {
			return nil, err
		}
		if changed {
			changedInput = append(changedInput, oneDependent.Relpath)
		}
	}
	workflowLog.ChangedInput = changedInput

	// re-check files
	for i, job := range workflowLog.JobLogs {
		if job.IsDone {
			if !job.IsAnyInputChanged {
				anyInputChanged := false
				for _, v := range job.InputFiles {
					changed, err := v.IsChanged()
					if err != nil {
						return nil, err
					}

					if changed {
						anyInputChanged = true
						break
					}
				}
				job.IsAnyInputChanged = anyInputChanged
			}

			if !job.IsAnyOutputChanged {
				anyOutputChanged := false
				for _, v := range job.OutputFiles {
					changed, err := v.IsChanged()
					if err != nil {
						return nil, err
					}

					if changed {
						anyOutputChanged = true
						break
					}
				}
				job.IsAnyOutputChanged = anyOutputChanged
			}

		} else {
			fmt.Fprintf(os.Stderr, "rescanning %s\n", job.JobLogRoot)
			newJob, err := CollectLogsForOneJob(job.JobLogRoot, job.ShellTask)
			if err != nil {
				return nil, err
			}
			workflowLog.JobLogs[i] = newJob
		}
	}

	cacheFile.Seek(0, os.SEEK_SET)
	cacheFileGzip := gzip.NewWriter(cacheFile)
	defer cacheFileGzip.Close()

	cacheEncoder := json.NewEncoder(cacheFileGzip)
	cacheEncoder.SetIndent("", "  ")
	err = cacheEncoder.Encode(workflowLog)
	if err != nil {
		return nil, fmt.Errorf("Failed to encode workflow log cache: %s", err.Error())
	}

	return &workflowLog, nil
}

func CollectLogsForOneWork(logdirPath string) (*WorkflowLog, error) {
	workflowLog, err := CollectLogsForOneWorkFromCache(logdirPath)
	if os.IsNotExist(err) {
		workflowLog, err = CollectLogsForOneWorkWithScan(logdirPath)
	}
	if err != nil {
		return nil, fmt.Errorf("Error while scanning workflow log: %s", err.Error())
	}
	return workflowLog, nil
}

func CollectLogsForOneWorkWithScan(logdirPath string) (*WorkflowLog, error) {
	jobs := make([]*JobLog, 0)

	var metadata WorkflowMetaData
	err := LoadJsonFromFile(path.Join(logdirPath, "runtime.json"), &metadata)
	if err != nil {
		return nil, nil
	}

	var dependentFiles []FileLog
	err = LoadJsonFromFile(path.Join(logdirPath, "input.json"), &dependentFiles)
	if err != nil {
		return nil, err
	}

	changedInput := make([]string, 0)
	for _, oneDependent := range dependentFiles {
		changed, err := oneDependent.IsChanged()
		if err != nil {
			return nil, err
		}
		if changed {
			changedInput = append(changedInput, oneDependent.Relpath)
		}
	}

	for _, oneTask := range metadata.Tasks {
		jobRoot := path.Join(logdirPath, fmt.Sprintf("job%03d", oneTask.ID))
		oneJob, err := CollectLogsForOneJob(jobRoot, oneTask)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, oneJob)
	}

	result := &WorkflowLog{
		WorkflowScript:  metadata.WorkflowPath,
		WorkflowLogRoot: logdirPath,
		ParameterFile:   metadata.ParameterFile,
		StartDate:       metadata.Date,
		JobLogs:         jobs,
		ChangedInput:    changedInput,
	}

	cacheFile, err := os.OpenFile(path.Join(logdirPath, workflowLogCacheFileName), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("Cannot open cache file: %s", err.Error())
	}
	defer cacheFile.Close()
	cacheFileGzip := gzip.NewWriter(cacheFile)
	defer cacheFileGzip.Close()

	cacheEncoder := json.NewEncoder(cacheFileGzip)
	cacheEncoder.SetIndent("", "  ")
	err = cacheEncoder.Encode(result)
	if err != nil {
		return nil, fmt.Errorf("Failed to encode workflow log cache: %s", err.Error())
	}

	return result, nil
}

func CollectLogs(logdir string) (WorkflowLogArray, error) {
	logdirFile, err := os.Open(logdir)
	if os.IsNotExist(err) {
		return WorkflowLogArray([]*WorkflowLog{}), nil
	}
	if err != nil {
		return nil, fmt.Errorf("Cannot open log directory in collectlogs: %s", err.Error())
	}

	workflowLogDirs, err := logdirFile.Readdir(0)
	if err != nil {
		return nil, fmt.Errorf("Cannot read log directory in collectlogs: %s", err.Error())
	}

	workflowLogs := WorkflowLogArray(make([]*WorkflowLog, 0))

	for _, oneWorkflowLogDir := range workflowLogDirs {
		if !oneWorkflowLogDir.IsDir() {
			continue
		}
		logdirPath := path.Join(logdir, oneWorkflowLogDir.Name())

		one, err := CollectLogsForOneWork(logdirPath)
		if err != nil {
			return nil, err
		}
		if one != nil {
			workflowLogs = append(workflowLogs, one)
		}
	}

	sort.Sort(workflowLogs)

	return workflowLogs, nil
}

type WorkflowLogArray []*WorkflowLog

func (v WorkflowLogArray) Len() int           { return len(v) }
func (v WorkflowLogArray) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v WorkflowLogArray) Less(i, j int) bool { return v[i].StartDate.Before(v[j].StartDate) }

func (v WorkflowLogArray) SearchReusableJob(shellscript string, workdir string, dependentFiles flowscript.StringSet, creatingFiles flowscript.StringSet) *JobLog {

	for _, x := range v {
		for _, y := range x.JobLogs {
			if (!y.IsReusable()) || (y.ShellTask.ShellScript != shellscript) || (!reflect.DeepEqual(y.ShellTask.DependentFiles, dependentFiles)) {
				continue
			}
			//fmt.Printf("found %s\n", y.JobLogRoot)
			return y
		}
	}

	return nil
}

const viewLogShowMax = 10

func ViewLog(showAll bool, failedOnly bool) error {
	logs, err := CollectLogs(WorkflowLogDir)
	if err != nil {
		return err
	}

	fmt.Printf("%3s|%7s|Success|Failed|Running|Pending|File Changed|%-19s|Name\n", "#", "State", "Start Date")

	showLogs := make([]int, 0)

	count := 0
	for i := len(logs) - 1; i >= 0; i-- {
		if !showAll && count >= viewLogShowMax {
			break
		}
		if !failedOnly || logs[i].State() == WorkflowFailed {
			count++
			showLogs = append(showLogs, i)
		}
	}

	sort.Sort(sort.IntSlice(showLogs))

	for _, i := range showLogs {
		v := logs[i]

		files := "No"
		if v.IsChanged() {
			files = "Yes"
		}

		successJobs := 0
		failedJobs := 0
		runningJobs := 0
		notStartedJobs := 0

		for _, x := range v.JobLogs {
			switch x.State() {
			case JobDone:
				successJobs++
			case JobFailed:
				failedJobs++
			case JobRunning:
				runningJobs++
			case JobPending:
				notStartedJobs++
			}
		}

		name := path.Base(v.WorkflowScript)
		if v.ParameterFile != "" {
			name += " " + path.Base(v.ParameterFile)
		}

		fmt.Printf("%3d|%7s|%7d|%6d|%7d|%7d|%12s|%10s|%s\n", i+1, v.State().String()[8:], successJobs, failedJobs, runningJobs, notStartedJobs, files, v.StartDate.Format("2006/01/02 15:04:05"), name)
	}
	return nil
}

func ViewLogDetail(args []string, failedOnly bool) error {
	logs, err := CollectLogs(WorkflowLogDir)
	if err != nil {
		return err
	}

	fmt.Printf("len: %d\n", len(logs))
	first := true
	for _, v := range args {
		val, err := strconv.ParseInt(v, 10, 32)
		if err == nil && val > 0 && val <= int64(len(logs)) {
			if first {
				first = false
			} else {
				fmt.Printf("===========================")
			}
			fmt.Printf("%s", logs[val-1].Summary(failedOnly))
		} else {
			fmt.Fprintf(os.Stderr, "Bad Workflow Record Number: %s\n", v)
		}
	}
	return nil
}
