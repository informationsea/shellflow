// Code generated by "stringer -type=JobState"; DO NOT EDIT.

package main

import "strconv"

const _JobState_name = "JobDoneJobRunningJobFailedJobPendingJobUnknown"

var _JobState_index = [...]uint8{0, 7, 17, 26, 36, 46}

func (i JobState) String() string {
	if i < 0 || i >= JobState(len(_JobState_index)-1) {
		return "JobState(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _JobState_name[_JobState_index[i]:_JobState_index[i+1]]
}
