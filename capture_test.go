package capture

import (
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestCapture(t *testing.T) {
	// ARRANGE
	fnerr := errors.New("function error")

	writeOutput := func() {
		fmt.Println("to stdout (1)")
		fmt.Println("to stdout (2)")
		os.Stderr.WriteString("to stderr (1)\n")
		os.Stderr.WriteString("to stderr (2)")
	}

	// ACT
	stdout, stderr, err := Output(func() error {
		writeOutput()
		return fnerr
	})

	// ASSERT
	t.Run("returns error", func(t *testing.T) {
		wanted := fnerr
		got := err
		if !errors.Is(got, wanted) {
			t.Errorf("\nwanted: %#v\ngot   : %#v", wanted, got)
		}
	})

	t.Run("stdout captured", func(t *testing.T) {
		wanted := []string{"to stdout (1)", "to stdout (2)"}
		got := stdout
		if len(wanted) == 0 || len(got) == 0 || len(wanted) != len(got) || wanted[0] != got[0] || wanted[1] != got[1] {
			t.Errorf("\nwanted: %v\ngot   : %v", wanted, got)
		}
	})

	t.Run("stderr captured", func(t *testing.T) {
		wanted := []string{"to stderr (1)", "to stderr (2)"}
		got := stderr
		if len(wanted) == 0 || len(got) == 0 || len(wanted) != len(got) || wanted[0] != got[0] || wanted[1] != got[1] {
			t.Errorf("\nwanted: %v\ngot   : %v", wanted, got)
		}
	})

	t.Run("when no output is produced", func(t *testing.T) {
		// ACT
		stdout, stderr, _ := Output(func() error { return nil })

		// ASSERT
		t.Run("stdout is nil", func(t *testing.T) {
			got := stdout
			if got != nil {
				t.Errorf("\nwanted: nil\ngot   : %v", got)
			}
		})

		t.Run("stderr is nil", func(t *testing.T) {
			got := stderr
			if got != nil {
				t.Errorf("\nwanted: nil\ngot   : %v", got)
			}
		})
	})

	t.Run("when error copying captured buffers", func(t *testing.T) {
		// ARRANGE
		cpyerr := fmt.Errorf("copy error")
		og := copyFn
		defer func() { copyFn = og }()
		copyFn = func(dst io.Writer, src io.Reader) (int64, error) { _, _ = io.Copy(dst, src); return 0, cpyerr }

		// ACT
		stdout, stderr, err := Output(func() error { fmt.Println("some output"); return nil })

		// ASSERT
		t.Run("errors", func(t *testing.T) {
			got := err

			wanted := ErrStdoutCapture
			if !errors.Is(got, wanted) {
				t.Errorf("\nwanted: %#v\ngot   : %#v", wanted, got)
			}

			wanted = ErrStderrCapture
			if !errors.Is(got, wanted) {
				t.Errorf("\nwanted: %#v\ngot   : %#v", wanted, got)
			}
		})

		t.Run("stdout is nil", func(t *testing.T) {
			got := stdout
			if got != nil {
				t.Errorf("\nwanted: nil\ngot   : %v", got)
			}
		})

		t.Run("stderr is nil", func(t *testing.T) {
			got := stderr
			if got != nil {
				t.Errorf("\nwanted: nil\ngot   : %v", got)
			}
		})
	})
}
