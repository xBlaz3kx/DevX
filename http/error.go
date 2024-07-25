package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	apiErrors "github.com/xBlaz3kx/DevX/errors"
)

type (
	// swagger:model errorResponse
	ErrorPayload struct {
		Error       string `json:"error" mapstructure:"error"`
		Code        int    `json:"code" mapstructure:"code"`
		Description string `json:"description" mapstructure:"description"`
	}

	// swagger:model emptyResponse
	EmptyResponse struct{}

	// swagger:model authError
	AuthError struct {
		Error Error `json:"error"`
	}

	Error struct {
		Code    int    `json:"code"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
)

func errorHandler(c *gin.Context) {
	c.Next()

	if len(c.Errors) == 0 {
		return
	}

	err := c.Errors[0].Err

	var apiErr apiErrors.ApiError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode() {

		case http.StatusUnauthorized:
			fallthrough
		case http.StatusForbidden:
			c.JSON(apiErr.StatusCode(), AuthError{
				Error: Error{
					Code:    apiErr.InternalCode(),
					Message: apiErr.Message(),
				},
			})

		default:
			c.JSON(apiErr.StatusCode(), ErrorPayload{
				Code:  apiErr.InternalCode(),
				Error: apiErr.Message(),
			})
		}
		return
	}

	c.JSON(http.StatusInternalServerError, ErrorPayload{
		Error:       "An unknown error occurred",
		Description: "Please try again later or contact support",
	})
}
