package apperror

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Error struct {
	err        error
	isError    bool
	statusCode int
}

type Interface interface {
	Exists() bool
	Error() string
	Unwrap() error
	AbortWithError(ctx *gin.Context)
}

func New(err error, errCode int) Error {
	return Error{err: err, isError: true, statusCode: errCode}
}

func NewWithMessage(errMsg string, errCode int) Error {
	return Error{err: errors.New(errMsg), isError: true, statusCode: errCode}
}

func (e Error) Exists() bool {
	return e.isError
}

func (e Error) Error() string {
	if e.err != nil {
		return e.err.Error()
	}

	return ""
}

func (e Error) Unwrap() error {
	return e.err
}

func (e Error) AbortWithError(ctx *gin.Context) {
	status := e.statusCode
	if status < 100 || status >= 600 {
		status = http.StatusInternalServerError
	}

	ctx.AbortWithStatusJSON(status, gin.H{
		"error": e.Error(),
	})
}
