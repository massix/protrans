package tools_test

import (
	"testing"

	"github.com/massix/protrans/pkg/tools"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type DivideByZeroError struct{}

func (e *DivideByZeroError) Error() string {
	return "Cannot divide by zero"
}

// Divides 42 by the given dividend
func divideFortyTwo(dividend int) (int, error) {
	if dividend == 0 {
		return 0, &DivideByZeroError{}
	}

	return (42 / dividend), nil
}

var lastError error

func warning(args ...interface{}) {
	lastError = args[0].(error)
}

func Test_AssertPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("Did not panic")
		} else {
			entry := r.(*logrus.Entry)
			assert.Equal(t, "Cannot divide by zero", entry.Message)
		}
	}()

	tools.Assert(divideFortyTwo(0))(logrus.Panic)
}

func Test_AssertNotPanic(t *testing.T) {
	res := tools.Assert(divideFortyTwo(1))(logrus.Panic)
	assert.Equal(t, 42, res)
}

func Test_ShouldPrintWarning(t *testing.T) {
	res := tools.Should(divideFortyTwo(0))(warning)
	assert.Equal(t, 0, res)
	assert.Equal(t, "Cannot divide by zero", lastError.Error())

	lastError = nil
}

func Test_ShouldNotPrintWarning(t *testing.T) {
	res := tools.Should(divideFortyTwo(42))(warning)
	assert.Equal(t, 1, res)
	assert.Nil(t, lastError)
}

func Test_OrElseActivated(t *testing.T) {
	res, err := divideFortyTwo(0)
	r := tools.OrElse(res, err, func(e error) int {
		return 128
	})

	assert.Equal(t, 128, r)
}

func Test_OrElseNotActivated(t *testing.T) {
	res, err := divideFortyTwo(42)
	r := tools.OrElse(res, err, func(e error) int {
		t.Error("This should not happen")
		return 0
	})

	assert.Equal(t, 1, r)
}
