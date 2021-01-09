package dxsvalue

import (
	"reflect"
	"sync"
	"time"
)

//转换接口
type DxValueMarshaler interface {
	EncodeToDxValue(dest *DxValue)
}

type DxValueUnMarshaler interface {
	DecodeFromDxValue(from *DxValue)
}


var(
	InterfaceType = reflect.TypeOf((*interface{})(nil)).Elem()
	StringType = reflect.TypeOf((*string)(nil)).Elem()
	TimePtrType = reflect.TypeOf((*time.Time)(nil))
	TimeType = TimePtrType.Elem()
	ErrorType = reflect.TypeOf((*error)(nil)).Elem()
	structTypePool	sync.Map			//reflect.Type
	ValueMarshalerType = reflect.TypeOf((*DxValueMarshaler)(nil)).Elem()
	ValueUnMarshalerType = reflect.TypeOf((*DxValueUnMarshaler)(nil)).Elem()
)

type StdValueFromDxValue	func(fvalue reflect.Value,value *DxValue)

func RegisterTypeMapFunc(tp reflect.Type,handler StdValueFromDxValue)  {
	structTypePool.Store(tp,handler)
}

