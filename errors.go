package capture

import "errors"

var (
	ErrStderrCapture = errors.New("stderr capture error")
	ErrStdoutCapture = errors.New("stdout capture error")
)
