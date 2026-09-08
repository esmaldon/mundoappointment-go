package httpmessage

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorInfo `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Success(c *gin.Context, status int, data any) {
	c.JSON(http.StatusOK, data)
}

func Fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, &ErrorInfo{Code: code, Message: message})
}
