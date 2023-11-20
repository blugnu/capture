package capture

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var copyFn = io.Copy

// capture is used to setup the capture of stdout or stderr.
// The function returns a function that must be called to restore
// the original stdout or stderr, a function that must be called
// to close the pipe (completing the capture) and a channel that
// will receive the captured output.
//
// Example:
//
//	  func DoSomething() {
//		rs, cl := capture(&os.Stdout)
//		defer rs()
//
//		fmt.Println("some output")
//		s, err := cl()
//
//		fmt.Println(s) // "some output"
//	  }
func capture(t **os.File) (func(), func() (string, error)) {
	og := *t
	r, w, _ := os.Pipe()
	*t = w

	c := make(chan string)
	e := make(chan error)
	go func() {
		var buf bytes.Buffer
		_, err := copyFn(&buf, r)
		c <- buf.String()
		e <- err
	}()

	return func() { *t = og }, func() (string, error) { w.Close(); return <-c, <-e }
}

// Output captures the stdout and stderr output produced during
// execution of a supplied function.
//
// If the supplied function returns an error, the error is returned
// together with any captured output from stdout and stderr.
//
// If an error occurs while capturing the output ErrStdoutCapture
// and/or ErrStderrCapture error are also returned.
//
//   - if ErrStdoutCapture is returned, any captured stdout output
//     is discarded.
//   - If ErrStderrCapture is returned, any captured stderr
//     output is discarded.
//   - If both ErrStdoutCapture and ErrStderrCapture are returned,
//     both captured outputs are discarded.
//
// These errors are returned wrapped with any error returned from
// the supplied function itself.
//
// Example:
//
//	  func DoSomething() {
//		stdout, stderr, err := capture.Output(func () error {
//		   return doSomething()
//		})
//
//		fmt.Printf("stdout: %v", stdout)
//		fmt.Printf("stderr: %v", stderr)
//		fmt.Printf("error: %v", err)
//	  }
func Output(fn func() error) ([]string, []string, error) {
	strings := func(s string) []string {
		if l := strings.Split(s, "\n"); len(l) > 1 || (len(l) == 1 && l[0] != "") {
			if l[len(l)-1:][0] == "" {
				l = l[:len(l)-1]
			}
			return l
		}
		return nil
	}

	restoreStdout, closeout := capture(&os.Stdout)
	defer restoreStdout()

	restoreStderr, closeerr := capture(&os.Stderr)
	defer restoreStderr()

	var (
		stdout string
		stderr string
		err    error
	)
	errs := []error{fn()}

	if stdout, err = closeout(); err != nil {
		errs = append(errs, fmt.Errorf("%w: %w", ErrStdoutCapture, err))
		stdout = "" // discard captured output
	}
	if stderr, err = closeerr(); err != nil {
		errs = append(errs, fmt.Errorf("%w: %w", ErrStderrCapture, err))
		stderr = "" // discard captured output
	}

	return strings(stdout), strings(stderr), errors.Join(errs...)
}
