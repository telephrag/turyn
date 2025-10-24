package turyn

import "fmt"

type TurynError struct {
	path string
	err  error
}

func Err(path string, err error) TurynError {
	return TurynError{path, err}
}

func (e TurynError) Error() string {
	return fmt.Sprintf("%s: %s\n", e.path, e.err)
}
