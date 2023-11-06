package response

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type responseError struct {
	Code    ErrCode `json:"code"`
	Message string  `json:"message"`
	Err     error   `json:"-"`
}

func (re *responseError) Error() string {
	if re == nil {
		return ""
	}
	s := `{
    "code": ` + strconv.Itoa(int(re.Code)) + `,
    "message": ` + re.Message + `
}`
	return s
}

func (re *responseError) GetCode() ErrCode {
	if re == nil {
		return 0
	}
	return re.Code
}

func (re *responseError) Unwrap() error {
	return re.Err
}

func IsResponseError(err error) bool {
	_, ok := err.(*responseError)
	return ok
}

// MultiError contains multiple errors and implements the error interface. Its
// zero value is ready to use. All its methods are goroutine safe.
type MultiError struct {
	mtx    sync.Mutex
	errors []error
}

func NewMultiError(err ...error) *MultiError {
	return &MultiError{
		errors: err,
	}
}

// Add adds an error to the MultiError.
func (e *MultiError) Add(err ...error) {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	e.errors = append(e.errors, err...)
}

// Len returns the number of errors added to the MultiError.
func (e *MultiError) Len() int {
	if e == nil {
		return 0
	}
	e.mtx.Lock()
	defer e.mtx.Unlock()

	return len(e.errors)
}

// MultiError returns the errors added to the MuliError. The returned slice is a
// copy of the internal slice of errors.
func (e *MultiError) Errors() []error {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	return append(make([]error, 0, len(e.errors)), e.errors...)
}

func (e *MultiError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Errors []error `json:"errors"`
	}{
		Errors: e.errors,
	})
}

func (e *MultiError) UnmarshalJSON(bytes []byte) error {
	errs := struct {
		Errors []*responseError `json:"errors"`
	}{}
	if err := json.Unmarshal(bytes, &errs); err != nil {
		return err
	}
	for _, err := range errs.Errors {
		e.Add(err)
	}
	return nil
}

func (e *MultiError) Error() string {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	es := make([]string, 0, len(e.errors))
	for _, err := range e.errors {
		es = append(es, err.Error())
	}
	return strings.Join(es, "; ")
}

func generateError(code ErrCode, s ...interface{}) *responseError {
	return &responseError{
		Code:    code,
		Message: fmt.Sprintf(errors[code], s...),
	}
}

func generateErrorWrapper(code ErrCode, err error, s ...interface{}) *responseError {
	return &responseError{
		Code:    code,
		Message: fmt.Sprintf(errors[code], s...),
		Err:     err,
	}
}

// https://golang.org/doc/faq#convert_slice_of_interface
func convert(infos []string) []interface{} {
	s := make([]interface{}, len(infos))
	for i, v := range infos {
		s[i] = v
	}
	return s
}

func ErrResourceExists(resource string) *responseError {
	return generateError(ErrCodeResourceExists, resource)
}
func ErrResourceNotFound(resource string) *responseError {
	return generateError(ErrCodeResourceNotFound, resource)
}
