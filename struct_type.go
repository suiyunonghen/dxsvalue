package dxsvalue

import (
	"reflect"
	"sync"
	"time"
)

var(
	InterfaceType = reflect.TypeOf((*interface{})(nil)).Elem()
	StringType = reflect.TypeOf((*string)(nil)).Elem()
	TimePtrType = reflect.TypeOf((*time.Time)(nil))
	TimeType = TimePtrType.Elem()
	ErrorType = reflect.TypeOf((*error)(nil)).Elem()
	structTypePool	sync.Map			//reflect.Type
)

type StdValueFromDxValue	func(fvalue reflect.Value,value *DxValue)

func RegisterTypeMapFunc(tp reflect.Type,handler StdValueFromDxValue)  {
	structTypePool.Store(tp,handler)
}

