package errors

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
)

type errorTestSuite struct {
	suite.Suite
}

func (s *errorTestSuite) TestError() {
	err := New(1, 200, "I am an error")
	s.Equal(200, err.StatusCode())
	s.Equal("I am an error", err.Message())
	s.Equal(1, err.InternalCode())

	s.Equal("I am an error", err.Error())

	// Base the error on another error
	newErr := err.New(2, "I am an error too")
	s.Equal(200, newErr.StatusCode())
	s.Equal("I am an error too", newErr.Message())
	s.Equal(2, newErr.InternalCode())
}

func (s *errorTestSuite) TestError_Error() {
	err := New(1, 200, "I am an error")

	err2 := errors.Wrap(err, "something went wrong")
	s.Equal("something went wrong: I am an error", err2.Error())

	err3 := errors.Wrap(err2, "something went wrong again")
	s.Equal("something went wrong again: something went wrong: I am an error", err3.Error())
}

func (s *errorTestSuite) TestError_IsError() {
	err1 := New(1, 200, "I am an error")
	err2 := New(1, 200, "I am an error")

	s.True(err1.Is(err2))

	err3 := New(2, 200, "I am not an error")
	s.False(err1.Is(err3))

	err3 = New(2, 300, "I am an error")
	s.False(err1.Is(err3))

	err3 = New(123, 200, "I am an error")
	s.False(err1.Is(err3))

	err3 = err1.New(2, "I am an error too")
	s.False(err1.Is(err3))
}

func (s *errorTestSuite) TestError_Unwrap() {
	err := New(1, 200, "I am an error")
	s.Nil(err.Unwrap())
}

func (s *errorTestSuite) TestError_StatusCode() {
	err1 := New(1, 200, "I am an error")
	err2 := New(1, 300, "I am an error")

	s.Equal(200, err1.StatusCode())
	s.Equal(300, err2.StatusCode())

	err3 := err1.New(2, "I am an error too")
	s.Equal(200, err3.StatusCode())
}

func TestErrors(t *testing.T) {
	suite.Run(t, new(errorTestSuite))
}
