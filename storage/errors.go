package storage

import "fmt"

type Multierror []error

func (me Multierror) Error() string {
	if len(me) == 0 {
		panic("empty multierror")
	}

	if len(me) == 1 {
		return me[0].Error()
	}

	return fmt.Sprintf("multiple errors: %v", me)
}

func (me Multierror) Result() error {
	if len(me) > 0 {
		return me
	}

	return nil
}

type ErrorRemovingFile struct {
	File string
	Err  error
}

func (e ErrorRemovingFile) Error() string {
	return fmt.Sprintf("error removing %s: %v", e.File, e.Err)
}

func (e ErrorRemovingFile) Unwrap() error {
	return e.Err
}
