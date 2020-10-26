package dxsvalue

import (
	"reflect"
	"time"
)

var(
	InterfaceType = reflect.TypeOf((*interface{})(nil)).Elem()
	StringType = reflect.TypeOf((*string)(nil)).Elem()
	TimePtrType = reflect.TypeOf((*time.Time)(nil))
	TimeType = TimePtrType.Elem()
	ErrorType = reflect.TypeOf((*error)(nil)).Elem()
)