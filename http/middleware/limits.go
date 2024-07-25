package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func UploadLimit(maxBytes int64) func(context *gin.Context) {
	return func(context *gin.Context) {
		var w http.ResponseWriter = context.Writer
		context.Request.Body = http.MaxBytesReader(w, context.Request.Body, maxBytes)
		context.Next()
	}
}
