package runner

import (
	"os/exec"
	"strconv"
)

type ExitError struct {
	Code    uint
	Wrapped error
	Stderr  []byte
}

func (e ExitError) Error() string {
	return "exit code: " + strconv.Itoa(int(e.Code)) + "; stderr: " + string(e.Stderr)
}

func (e ExitError) Unwrap() error {
	return e.Wrapped
}

func ExitErrorFromExec(e *exec.ExitError) *ExitError {
	return &ExitError{
		Code:    uint(e.ExitCode()),
		Wrapped: e,
		Stderr:  e.Stderr,
	}
}
