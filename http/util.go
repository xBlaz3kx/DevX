package http

import (
	"context"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

var ErrInvalidFieldsQuery = errors.New("Invalid fields query")

// AddPaginationToContext adds the pagination parameters to the context
func AddPaginationToContext(ctx *gin.Context) context.Context {
	reqCtx := ctx.Request.Context()
	// nolint:all
	reqCtx = context.WithValue(reqCtx, "offset", ctx.Query("offset"))
	// nolint:all
	reqCtx = context.WithValue(reqCtx, "limit", ctx.Query("limit"))
	return reqCtx
}

// ParseQueryFields parses the query fields and returns a fields and an alias map
// Use fieldsMap to check if a field is present in the query
func ParseQueryFields(q string, allowed []string, locked []string) ([]string, map[string]bool, error) {
	fieldsMap := make(map[string]bool)
	aliasMap := make(map[string]string)

	if allowed == nil {
		return []string{}, fieldsMap, nil
	}

	if locked != nil {
		fieldsMap = lo.SliceToMap(locked, func(item string) (string, bool) {
			return item, true
		})

		aliasMap = lo.SliceToMap(locked, func(item string) (string, string) {
			return item, item
		})
	}

	// No query case, return all
	if q == "" {
		fieldsMap = lo.SliceToMap(allowed, func(item string) (string, bool) {
			return item, true
		})
		return allowed, fieldsMap, nil
	}

	addField := regexp.MustCompile("^[[:alnum:]]+$")
	aliasField := regexp.MustCompile("^[[:alnum:]]+:[[:alnum:]]+$")

	tokens := strings.Split(q, ",")
	for _, token := range tokens {
		switch {
		case addField.MatchString(token):
			fieldsMap[token] = true
			aliasMap[token] = token
		case aliasField.MatchString(token):
			field := strings.Split(token, ":")
			fieldsMap[field[0]] = true
			aliasMap[field[0]] = field[1]
		default:
			return nil, nil, ErrInvalidFieldsQuery
		}
	}

	fields := allowed

	// Filter included fields and map to aliases
	fields = lo.Map(lo.Filter(fields, func(item string, index int) bool {
		return fieldsMap[item]
	}), func(field string, index int) string {
		return aliasMap[field]
	})

	return fields, fieldsMap, nil
}
