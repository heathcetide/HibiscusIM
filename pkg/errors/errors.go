package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// Error represents a custom error with stack trace
type Error struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Err     error      `json:"-"` // 原始错误，不序列化
	Stack   string     `json:"stack,omitempty"`
	Context []KeyValue `json:"context,omitempty"`
}

// KeyValue represents a key-value pair for context
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

// Unwrap implements the errors.Wrapper interface
func (e *Error) Unwrap() error {
	return e.Err
}

// WithCode creates a new error with code
func WithCode(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Stack:   captureStack(),
	}
}

// WithCodef creates a new error with code and formatted message
func WithCodef(code int, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Stack:   captureStack(),
	}
}

// Wrap wraps an error with message
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}

	return &Error{
		Message: message,
		Err:     err,
		Stack:   captureStack(),
	}
}

// Wrapf wraps an error with formatted message
func Wrapf(err error, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}

	return &Error{
		Message: fmt.Sprintf(format, args...),
		Err:     err,
		Stack:   captureStack(),
	}
}

// New creates a new error
func New(message string) *Error {
	return &Error{
		Message: message,
		Stack:   captureStack(),
	}
}

// Errorf creates a new formatted error
func Errorf(format string, args ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Stack:   captureStack(),
	}
}

// WithContext adds context to an error
func (e *Error) WithContext(key, value string) *Error {
	if e == nil {
		return nil
	}

	// 创建新的错误实例以避免修改原始错误
	newErr := &Error{
		Code:    e.Code,
		Message: e.Message,
		Err:     e.Err,
		Stack:   e.Stack,
		Context: make([]KeyValue, len(e.Context)),
	}

	// 复制现有上下文
	copy(newErr.Context, e.Context)

	// 添加新上下文
	newErr.Context = append(newErr.Context, KeyValue{Key: key, Value: value})

	return newErr
}

// WithContexts adds multiple contexts to an error
func (e *Error) WithContexts(kv map[string]string) *Error {
	if e == nil || len(kv) == 0 {
		return e
	}

	// 创建新的错误实例
	newErr := &Error{
		Code:    e.Code,
		Message: e.Message,
		Err:     e.Err,
		Stack:   e.Stack,
		Context: make([]KeyValue, len(e.Context)),
	}

	// 复制现有上下文
	copy(newErr.Context, e.Context)

	// 添加新上下文
	for k, v := range kv {
		newErr.Context = append(newErr.Context, KeyValue{Key: k, Value: v})
	}

	return newErr
}

// captureStack captures the current stack trace
func captureStack() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])

	// 移除顶部几行（通常是 captureStack 和 Error 相关的调用）
	lines := strings.Split(stack, "\n")
	if len(lines) > 6 {
		stack = strings.Join(lines[6:], "\n")
	}

	return strings.TrimSpace(stack)
}

// GetCode returns the error code
func GetCode(err error) int {
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return 0
}

// GetMessage returns the error message
func GetMessage(err error) string {
	if e, ok := err.(*Error); ok {
		return e.Message
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

// GetStack returns the error stack trace
func GetStack(err error) string {
	if e, ok := err.(*Error); ok {
		return e.Stack
	}
	return ""
}

// Is checks if the error chain contains the target error
func Is(err, target error) bool {
	if e, ok := err.(*Error); ok {
		return e.Message == target.Error() || (e.Err != nil && e.Err.Error() == target.Error())
	}
	return err == target
}

// Cause returns the underlying error
func Cause(err error) error {
	for err != nil {
		if e, ok := err.(*Error); ok && e.Err != nil {
			err = e.Err
		} else {
			return err
		}
	}
	return err
}

// Format implements fmt.Formatter
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%s", e.Error())
			if e.Stack != "" {
				fmt.Fprintf(s, "\n%s", e.Stack)
			}
			return
		}
		fallthrough
	case 's':
		fmt.Fprintf(s, "%s", e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}
