package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type parseQueryFieldsTestIn struct {
	query   string
	allowed []string
	locked  []string
}

type parseQueryFieldsTestOut struct {
	fields    []string
	fieldsMap map[string]bool
	errorOut  error
}

type parseQueryFieldsTest struct {
	in  parseQueryFieldsTestIn
	out parseQueryFieldsTestOut
}

var parseQueryFieldsTestCases = []parseQueryFieldsTest{
	{
		in: parseQueryFieldsTestIn{
			query:   "field1:alias1,field2,field3",
			allowed: []string{"field1", "field2", "field3"},
			locked:  []string{},
		},
		out: parseQueryFieldsTestOut{
			fields: []string{"alias1", "field2", "field3"},
			fieldsMap: map[string]bool{
				"field1": true,
				"field2": true,
				"field3": true,
			},
			errorOut: nil,
		},
	},
	{
		in: parseQueryFieldsTestIn{
			query:   "field2,field3",
			allowed: []string{"field1", "field2", "field3"},
			locked:  []string{"field1"},
		},
		out: parseQueryFieldsTestOut{
			fields: []string{"field1", "field2", "field3"},
			fieldsMap: map[string]bool{
				"field1": true,
				"field2": true,
				"field3": true,
			},
			errorOut: nil,
		},
	},
	{
		in: parseQueryFieldsTestIn{
			query:   "-field1",
			allowed: []string{"field1", "field2", "field3"},
			locked:  []string{},
		},
		out: parseQueryFieldsTestOut{
			fields:    nil,
			fieldsMap: nil,
			errorOut:  ErrInvalidFieldsQuery,
		},
	},
	{
		in: parseQueryFieldsTestIn{
			query:   "",
			allowed: []string{"field1", "field2", "field3"},
			locked:  []string{},
		},
		out: parseQueryFieldsTestOut{
			fields: []string{"field1", "field2", "field3"},
			fieldsMap: map[string]bool{
				"field1": true,
				"field2": true,
				"field3": true,
			},
			errorOut: nil,
		},
	},
	{
		in: parseQueryFieldsTestIn{
			query:   "field3",
			allowed: []string{"field1", "field2", "field3"},
			locked:  []string{"field1"},
		},
		out: parseQueryFieldsTestOut{
			fields: []string{"field1", "field3"},
			fieldsMap: map[string]bool{
				"field1": true,
				"field3": true,
			},
			errorOut: nil,
		},
	},
}

func TestParseQueryFields(t *testing.T) {
	for _, tst := range parseQueryFieldsTestCases {
		fields, fieldsMap, err := ParseQueryFields(tst.in.query, tst.in.allowed, tst.in.locked)
		assert.Equal(t, tst.out.fields, fields)
		assert.Equal(t, tst.out.fieldsMap, fieldsMap)
		assert.Equal(t, tst.out.errorOut, err)
	}
}
