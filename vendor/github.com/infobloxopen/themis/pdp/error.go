package pdp

//go:generate bash -c "(egen -i $GOPATH/src/github.com/infobloxopen/themis/pdp/errors.yaml > $GOPATH/src/github.com/infobloxopen/themis/pdp/errors.go) && gofmt -l -s -w $GOPATH/src/github.com/infobloxopen/themis/pdp/errors.go"

import (
	"fmt"
	"strings"
)

const errorSourcePathSeparator = ">"

type boundError interface {
	error
	bind(src string)
}

func bindError(err error, src string) boundError {
	b, ok := err.(boundError)
	if ok {
		b.bind(src)
		return b
	}

	return bindError(newExternalError(err), src)
}

func bindErrorf(err error, format string, args ...interface{}) boundError {
	return bindError(err, fmt.Sprintf(format, args...))
}

type errorLink struct {
	id   int
	path []string
}

func (e *errorLink) errorf(format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)

	if len(e.path) > 0 {
		return fmt.Sprintf("#%02x (%s): %s", e.id, strings.Join(e.path, errorSourcePathSeparator), msg)
	}

	return fmt.Sprintf("#%02x: %s", e.id, msg)
}

func (e *errorLink) bind(src string) {
	e.path = append([]string{src}, e.path...)
}
