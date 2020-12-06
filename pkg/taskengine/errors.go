package taskengine

import (
	"fmt"
)

// var TaskSkipped = errors.New("task skipped")

// engineError is a taskengine error.
// type engineError string

// func (e engineError) Error() string {
// 	return string(e)
// }

func errorf(format string, a ...interface{}) error {
	// return engineError(fmt.Sprintf(format, a...))
	return fmt.Errorf(format, a...)
}

// taskError is ...
// type taskError struct {
// 	err error
// }

// type StatusType int

// const (
// 	Success StatusType = iota
// 	Error
// 	Canceled
// 	Skipped
// )

// func (e *taskError) Status() StatusType {
// 	if e == nil || e.err == nil {
// 		return Success
// 	}
// 	if errors.Is(e.err, context.Canceled) {
// 		return Canceled
// 	}
// 	if errors.Is(e.err, TaskSkipped) {
// 		return Skipped
// 	}
// 	return Error
// }
