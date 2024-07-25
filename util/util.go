package util

import (
	"reflect"
)

// IsNilInterfaceOrPointer checks if the given interface is nil or if it is a pointer and the pointer is nil.
func IsNilInterfaceOrPointer(sth interface{}) bool {
	return sth == nil || (reflect.ValueOf(sth).Kind() == reflect.Ptr && reflect.ValueOf(sth).IsNil())
}
