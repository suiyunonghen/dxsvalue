package dxsvalue

import (
	"encoding/json"
	"github.com/suiyunonghen/DxCommonLib"
	"reflect"
	"strings"
	"time"
)

func decode2reflectFromdxValue(fvalue reflect.Value,value *DxValue,ignoreCase bool,valueType reflect.Type) {
	value.Visit(func(Key string, childvalue *DxValue) bool {
		for i := 0;i<valueType.NumField();i++{
			fld := valueType.Field(i)
			if ignoreCase && strings.EqualFold(Key,fld.Name) || Key == fld.Name{
				childreflectv := fvalue.Field(i)
				tp := fld.Type
				vhandler,ok := structTypePool.Load(tp)
				if ok && vhandler != nil{
					convertHandler := vhandler.(StdValueFromDxValue)
					if convertHandler != nil{
						convertHandler(childreflectv,childvalue)
						break
					}
				}
				switch tp.Kind() {
				case reflect.String:
					childreflectv.SetString(childvalue.String())
				case reflect.Int,reflect.Int64,reflect.Int8,reflect.Int16,reflect.Int32:
					childreflectv.SetInt(childvalue.Int())
				case reflect.Uint,reflect.Uint64,reflect.Uint8,reflect.Uint16,reflect.Uint32:
					childreflectv.SetUint(uint64(childvalue.Int()))
				case reflect.Float32,reflect.Float64:
					childreflectv.SetFloat(childvalue.Double())
				case reflect.Bool:
					childreflectv.SetBool(childvalue.Bool())
				case reflect.Slice:
					decodeArray2reflect(childreflectv,childvalue,ignoreCase)
				case reflect.Struct:
					if tp == TimeType {
						childreflectv.Set(reflect.ValueOf(childvalue.GoTime()))
					}else{
						decode2reflectFromdxValue(childreflectv,childvalue,ignoreCase,tp)
					}
				}
				break
			}
		}
		return true
	})
}

func growSliceValue(v reflect.Value, n int) reflect.Value {
	diff := n - v.Len()
	if diff > 256 {
		diff = 256
	}
	v = reflect.AppendSlice(v, reflect.MakeSlice(v.Type(), diff, diff))
	return v
}

func decodeArray2reflect(sliceValue reflect.Value,arrvalue *DxValue,ignoreCase bool)bool  {
	if arrvalue.DataType != VT_Array{
		return false
	}
	n := arrvalue.Count()
	if n == 0{
		return false
	}
	var vtype reflect.Type
	if sliceValue.Cap() == 0{
		vslice := reflect.MakeSlice(sliceValue.Type(), n, n)
		sv := vslice.Index(0)
		vtype = sv.Type()
		sliceValue.Set(vslice.Slice(0, n))
	}else if sliceValue.Cap() >= n {
		vtype = sliceValue.Slice(1,1).Field(0).Type()
		sliceValue.Set(sliceValue.Slice(0, n))
	} else if sliceValue.Len() < sliceValue.Cap() {
		vtype = sliceValue.Slice(1,1).Field(0).Type()
		sliceValue.Set(sliceValue.Slice(0, sliceValue.Cap()))
	}

	vhandler,ok := structTypePool.Load(vtype)
	if ok && vhandler != nil{
		convertHandler := vhandler.(StdValueFromDxValue)
		if convertHandler != nil{
			for i := 0;i<n;i++{
				arrNode := arrvalue.ValueByIndex(i)
				if i >= sliceValue.Len() {
					sliceValue.Set(growSliceValue(sliceValue, n))
				}
				convertHandler(sliceValue.Field(i),arrNode)
			}
			return true
		}
	}

	if vtype == TimeType{
		for i := 0;i<n;i++{
			arrNode := arrvalue.ValueByIndex(i)
			if i >= sliceValue.Len() {
				sliceValue.Set(growSliceValue(sliceValue, n))
			}
			sliceValue.Field(i).Set(reflect.ValueOf(arrNode.GoTime()))
		}
		return true
	}
	switch vtype.Kind() {
	case reflect.String:
		for i := 0;i<n;i++{
			arrNode := arrvalue.ValueByIndex(i)
			if i >= sliceValue.Len() {
				sliceValue.Set(growSliceValue(sliceValue, n))
			}
			sliceValue.Index(i).SetString(arrNode.String())
		}
	case reflect.Int,reflect.Int64,reflect.Int8,reflect.Int16,reflect.Int32:
		for i := 0;i<n;i++{
			arrNode := arrvalue.ValueByIndex(i)
			if i >= sliceValue.Len() {
				sliceValue.Set(growSliceValue(sliceValue, n))
			}
			sliceValue.Index(i).SetInt(arrNode.Int())
		}
	case reflect.Uint,reflect.Uint64,reflect.Uint8,reflect.Uint16,reflect.Uint32:
		for i := 0;i<n;i++{
			arrNode := arrvalue.ValueByIndex(i)
			if i >= sliceValue.Len() {
				sliceValue.Set(growSliceValue(sliceValue, n))
			}
			sliceValue.Index(i).SetUint(uint64(arrNode.Int()))
		}
	case reflect.Float32,reflect.Float64:
		for i := 0;i<n;i++{
			arrNode := arrvalue.ValueByIndex(i)
			if i >= sliceValue.Len() {
				sliceValue.Set(growSliceValue(sliceValue, n))
			}
			sliceValue.Index(i).SetFloat(arrNode.Double())
		}
	case reflect.Bool:
		for i := 0;i<n;i++{
			arrNode := arrvalue.ValueByIndex(i)
			if i >= sliceValue.Len() {
				sliceValue.Set(growSliceValue(sliceValue, n))
			}
			sliceValue.Index(i).SetBool(arrNode.Bool())
		}
	case reflect.Slice:
		for i := 0;i<n;i++{
			arrNode := arrvalue.ValueByIndex(i)
			if i >= sliceValue.Len() {
				sliceValue.Set(growSliceValue(sliceValue, n))
			}
			decodeArray2reflect(sliceValue.Index(i),arrNode,ignoreCase)
		}
	case reflect.Struct:
		for i := 0;i<n;i++{
			arrNode := arrvalue.ValueByIndex(i)
			if i >= sliceValue.Len() {
				sliceValue.Set(growSliceValue(sliceValue, n))
			}

			decode2reflectFromdxValue(sliceValue.Index(i),arrNode,ignoreCase,vtype)
		}

	default:
		return false
	}
	return true
}

func (v *DxValue)SetValue(value interface{})  {
	if v == valueNAN || v == valueINF || v == valueTrue || v == valueFalse || v == valueNull{
		return
	}
	switch realv := value.(type) {
	case int:
		v.SetInt(int64(realv))
	case *int:
		v.SetInt(int64(*realv))
	case uint:
		v.SetInt(int64(realv))
	case *uint:
		v.SetInt(int64(*realv))
	case int32:
		v.SetInt(int64(realv))
	case *int32:
		v.SetInt(int64(*realv))
	case uint32:
		v.SetInt(int64(realv))
	case *uint32:
		v.SetInt(int64(*realv))
	case int16:
		v.SetInt(int64(realv))
	case *int16:
		v.SetInt(int64(*realv))
	case uint16:
		v.SetInt(int64(realv))
	case *uint16:
		v.SetInt(int64(*realv))
	case int8:
		v.SetInt(int64(realv))
	case *int8:
		v.SetInt(int64(*realv))
	case uint8:
		v.SetInt(int64(realv))
	case *uint8:
		v.SetInt(int64(*realv))
	case int64:
		v.SetInt(realv)
	case *int64:
		v.SetInt(*realv)
	case uint64:
		v.SetInt(int64(realv))
	case *uint64:
		v.SetInt(int64(*realv))
	case string:
		v.SetString(realv)
	case *DxValue:
		v.CopyFrom(realv,nil)
	case DxValue:
		v.CopyFrom(&realv,nil)
	case bool:
		v.SetBool(realv)
	case *bool:
		v.SetBool(*realv)
	case time.Time:
		v.SetDouble(float64(DxCommonLib.Time2DelphiTime(realv)))
	case *time.Time:
		v.SetDouble(float64(DxCommonLib.Time2DelphiTime(*realv)))
	case float32:
		v.SetFloat(realv)
	case *float32:
		v.SetFloat(*realv)
	case float64:
		v.SetDouble(realv)
	case *float64:
		v.SetDouble(*realv)
	case []byte:
		v.SetBinary(realv,true)
	case *[]byte:
		v.SetBinary(*realv,true)
	case map[string]interface{}:
		v.Reset(VT_Object)
		cache := v.ValueCache()
		for key,objv := range realv{
			v.SetKeyvalue(key,objv,cache)
		}
	case *map[string]interface{}:
		v.Reset(VT_Object)
		cache := v.ValueCache()
		for key,objv := range *realv{
			v.SetKeyvalue(key,objv,cache)
		}
	case map[string]string:
		v.Reset(VT_Object)
		cache := v.ValueCache()
		for key,objv := range realv{
			v.SetKeyCached(key,VT_String,cache).SetString(objv)
		}
	case *map[string]string:
		v.Reset(VT_Object)
		cache := v.ValueCache()
		for key,objv := range *realv{
			v.SetKeyCached(key,VT_String,cache).SetString(objv)
		}
	case map[string]int:
		v.Reset(VT_Object)
		cache := v.ValueCache()
		for key,objv := range realv{
			v.SetKeyCached(key,VT_Int,cache).SetInt(int64(objv))
		}
	case *map[string]int:
		v.Reset(VT_Object)
		cache := v.ValueCache()
		for key,objv := range *realv{
			v.SetKeyCached(key,VT_Int,cache).SetInt(int64(objv))
		}
	case []interface{}:
		v.Reset(VT_Array)
		cache := v.ValueCache()
		for i := 0;i<len(realv);i++{
			v.SetIndexvalue(i,realv[i],cache)
		}
	case *[]interface{}:
		v.Reset(VT_Array)
		cache := v.ValueCache()
		for i := 0;i<len(*realv);i++{
			v.SetIndexvalue(i,(*realv)[i],cache)
		}
	case []string:
		v.Reset(VT_Array)
		cache := v.ValueCache()
		for i := 0;i<len(realv);i++{
			v.SetIndexCached(i,VT_String,cache).SetString(realv[i])
		}
	case *[]string:
		v.Reset(VT_Array)
		cache := v.ValueCache()
		for i := 0;i<len(*realv);i++{
			v.SetIndexCached(i,VT_String,cache).SetString((*realv)[i])
		}
	default:
		//判断一下是否是结构体
		reflectv := reflect.ValueOf(value)
		if !reflectv.IsValid(){
			return
		}
		if reflectv.Type().Implements(ValueMarshalerType){
			value.(DxValueMarshaler).EncodeToDxValue(v)
			return
		}
		if reflectv.Kind() == reflect.Ptr{
			reflectv = reflectv.Elem()
		}/*else if reflect.PtrTo(reflectv.Type()).Implements(ValueMarshalerType){
			value.(DxValueMarshaler).EncodeToDxValue(v)
			return
		}*/
		switch reflectv.Kind(){
		case reflect.Struct:
			tp := reflectv.Type()
			if tp == TimeType{
				vi := reflectv.Interface()
				v.SetTime(vi.(time.Time))
				return
			}else if tp == TimePtrType{
				vi := reflectv.Interface()
				v.SetTime(*(vi.(*time.Time)))
				return
			}
			v.Reset(VT_Object)
			cache := v.ValueCache()
			rtype := reflectv.Type()
			for i := 0;i < rtype.NumField();i++{
				sfield := rtype.Field(i)
				fv := reflectv.Field(i)
				if fv.Kind() == reflect.Ptr{
					fv = fv.Elem()
				}
				switch fv.Kind() {
				case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64,
					reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
					v.SetKeyCached(sfield.Name,VT_Int,cache).SetInt(fv.Int())
				case reflect.Bool:
					if fv.Bool(){
						v.SetKeyValue(sfield.Name,valueTrue)
					}else{
						v.SetKeyValue(sfield.Name,valueFalse)
					}
				case reflect.Float32:
					v.SetKeyCached(sfield.Name,VT_Double,cache).SetFloat(float32(fv.Float()))
				case reflect.Float64:
					v.SetKeyCached(sfield.Name,VT_Double,cache).SetDouble(fv.Float())
				case reflect.String:
					v.SetKeyCached(sfield.Name,VT_String,cache).SetString(fv.String())
				default:
					if fv.CanInterface(){
						if fv.Type() == TimeType{
							v.SetKeyCached(sfield.Name,VT_DateTime,cache).SetTime(fv.Interface().(time.Time))
						}else if fv.Type() == TimePtrType{
							v.SetKeyCached(sfield.Name,VT_DateTime,cache).SetTime(*(fv.Interface().(*time.Time)))
						}else{
							v.SetKeyvalue(sfield.Name,fv.Interface(),cache)
						}
					}
				}
			}
		case reflect.Map:
			mapkeys := reflectv.MapKeys()
			if len(mapkeys) == 0{
				return
			}
			kv := mapkeys[0]
			if kv.Type().Kind() == reflect.Ptr{
				kv = kv.Elem()
			}
			if kv.Kind() != reflect.String{
				return
			}
			v.Reset(VT_Object)
			cache := v.ValueCache()
			for _,kv = range mapkeys{
				rvalue := reflectv.MapIndex(kv)
				if rvalue.Kind() == reflect.Ptr{
					rvalue = rvalue.Elem()
				}
				switch rvalue.Kind() {
				case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
					v.SetKeyCached(kv.String(),VT_Int,cache).SetInt(rvalue.Int())
				case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
					v.SetKeyCached(kv.String(),VT_Int,cache).SetInt(int64(rvalue.Uint()))
				case reflect.Bool:
					if rvalue.Bool(){
						v.SetKeyValue(kv.String(),valueTrue)
					}else{
						v.SetKeyValue(kv.String(),valueFalse)
					}
				case reflect.Float32:
					v.SetKeyCached(kv.String(),VT_Double,cache).SetFloat(float32(rvalue.Float()))
				case reflect.Float64:
					v.SetKeyCached(kv.String(),VT_Double,cache).SetDouble(rvalue.Float())
				case reflect.String:
					v.SetKeyCached(kv.String(),VT_String,cache).SetString(rvalue.String())
				default:
					if rvalue.CanInterface(){
						if rvalue.Type() == TimeType{
							v.SetKeyCached(kv.String(),VT_DateTime,cache).SetTime(rvalue.Interface().(time.Time))
						}else if rvalue.Type() == TimePtrType{
							v.SetKeyCached(kv.String(),VT_DateTime,cache).SetTime(*(rvalue.Interface().(*time.Time)))
						}else{
							v.SetKeyvalue(kv.String(),rvalue.Interface(),cache)
						}
					}
				}
			}
		case reflect.Slice:
			v.Reset(VT_Array)
			cache := v.ValueCache()
			vlen := reflectv.Len()
			for i := 0;i<vlen;i++{
				av := reflectv.Index(i)
				realvalue := getRealValue(&av)
				if realvalue == nil{
					v.SetIndexCached(i,VT_NULL,cache)
				}else if realvalue.CanInterface(){
					v.SetIndexvalue(i,realvalue.Interface(),cache)
				}
			}
		}

	}
}

func (v *DxValue)SetKeyvalue(Name string,value interface{},cache *ValueCache)  {
	switch realv := value.(type) {
	case int:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(realv))
	case *int:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(*realv))
	case uint:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(realv))
	case *uint:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(*realv))
	case int32:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(realv))
	case *int32:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(*realv))
	case uint32:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(realv))
	case *uint32:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(*realv))
	case int16:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(realv))
	case *int16:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(*realv))
	case uint16:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(realv))
	case *uint16:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(*realv))
	case int8:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(realv))
	case *int8:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(*realv))
	case uint8:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(realv))
	case *uint8:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(*realv))
	case int64:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(realv)
	case *int64:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(*realv))
	case uint64:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(realv))
	case *uint64:
		v.SetKeyCached(Name,VT_Int,cache).SetInt(int64(*realv))
	case string:
		v.SetKeyCached(Name,VT_String,cache).SetString(realv)
	case *DxValue:
		v.SetKeyValue(Name,realv)
	case DxValue:
		v.SetKeyValue(Name,&realv)
	case bool:
		if realv{
			v.SetKeyValue(Name,valueTrue)
		}else{
			v.SetKeyValue(Name,valueFalse)
		}
	case *bool:
		if *realv{
			v.SetKeyValue(Name,valueTrue)
		}else{
			v.SetKeyValue(Name,valueFalse)
		}
	case time.Time:
		v.SetKeyCached(Name,VT_DateTime,cache).SetDouble(float64(DxCommonLib.Time2DelphiTime(realv)))
	case *time.Time:
		v.SetKeyCached(Name,VT_DateTime,cache).SetDouble(float64(DxCommonLib.Time2DelphiTime(*realv)))
	case float32:
		v.SetKeyCached(Name,VT_Float,cache).SetFloat(realv)
	case *float32:
		v.SetKeyCached(Name,VT_Float,cache).SetFloat(*realv)
	case float64:
		v.SetKeyCached(Name,VT_Double,cache).SetDouble(realv)
	case *float64:
		v.SetKeyCached(Name,VT_Double,cache).SetDouble(*realv)
	case []byte:
		v.SetKeyCached(Name,VT_Binary,cache).SetBinary(realv,true)
	case *[]byte:
		v.SetKeyCached(Name,VT_Binary,cache).SetBinary(*realv,true)
	case map[string]interface{}:
		newv := v.SetKeyCached(Name,VT_Object,cache)
		for key,objv := range realv{
			newv.SetKeyvalue(key,objv,cache)
		}
	case *map[string]interface{}:
		newv := v.SetKeyCached(Name,VT_Object,cache)
		for key,objv := range *realv{
			newv.SetKeyvalue(key,objv,cache)
		}
	case map[string]string:
		newv := v.SetKeyCached(Name,VT_Object,cache)
		for key,objv := range realv{
			newv.SetKeyCached(key,VT_String,cache).SetString(objv)
		}
	case *map[string]string:
		newv := v.SetKeyCached(Name,VT_Object,v.ownercache)
		for key,objv := range *realv{
			newv.SetKeyCached(key,VT_String,cache).SetString(objv)
		}
	case map[string]int:
		newv := v.SetKeyCached(Name,VT_Object,cache)
		for key,objv := range realv{
			newv.SetKeyCached(key,VT_Int,cache).SetInt(int64(objv))
		}
	case *map[string]int:
		newv := v.SetKeyCached(Name,VT_Object,cache)
		for key,objv := range *realv{
			newv.SetKeyCached(key,VT_Int,cache).SetInt(int64(objv))
		}
	case []interface{}:
		newv := v.SetKeyCached(Name,VT_Array,cache)
		for i := 0;i<len(realv);i++{
			newv.SetIndexvalue(i,realv[i],cache)
		}
	case *[]interface{}:
		newv := v.SetKeyCached(Name,VT_Array,cache)
		for i := 0;i<len(*realv);i++{
			newv.SetIndexvalue(i,(*realv)[i],cache)
		}
	case []string:
		newv := v.SetKeyCached(Name,VT_Array,cache)
		for i := 0;i<len(realv);i++{
			newv.SetIndexCached(i,VT_String,cache).SetString(realv[i])
		}
	case *[]string:
		newv := v.SetKeyCached(Name,VT_Array,cache)
		for i := 0;i<len(*realv);i++{
			newv.SetIndexCached(i,VT_String,cache).SetString((*realv)[i])
		}
	default:
		//判断一下是否是结构体
		reflectv := reflect.ValueOf(value)
		if !reflectv.IsValid(){
			return
		}
		if reflectv.Type().Implements(ValueMarshalerType){
			value.(DxValueMarshaler).EncodeToDxValue(v.SetKeyCached(Name,VT_Int,cache))
			return
		}
		if reflectv.Kind() == reflect.Ptr{
			reflectv = reflectv.Elem()
		}
		switch reflectv.Kind(){
		case reflect.Struct:
			rtype := reflectv.Type()
			if rtype == TimeType{
				v.SetKeyCached(Name,VT_DateTime,cache).SetTime(reflectv.Interface().(time.Time))
				return
			}else if rtype == TimePtrType{
				v.SetKeyCached(Name,VT_DateTime,cache).SetTime(*(reflectv.Interface().(*time.Time)))
				return
			}
			newv := v.SetKeyCached(Name,VT_Object,cache)
			for i := 0;i < rtype.NumField();i++{
				sfield := rtype.Field(i)
				fv := reflectv.Field(i)
				if fv.Kind() == reflect.Ptr{
					fv = fv.Elem()
				}
				switch fv.Kind() {
				case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64,
					reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
					newv.SetKeyCached(sfield.Name,VT_Int,cache).SetInt(fv.Int())
				case reflect.Bool:
					if fv.Bool(){
						newv.SetKeyValue(sfield.Name,valueTrue)
					}else{
						newv.SetKeyValue(sfield.Name,valueFalse)
					}
				case reflect.Float32:
					newv.SetKeyCached(sfield.Name,VT_Double,cache).SetFloat(float32(fv.Float()))
				case reflect.Float64:
					newv.SetKeyCached(sfield.Name,VT_Double,cache).SetDouble(fv.Float())
				case reflect.String:
					newv.SetKeyCached(sfield.Name,VT_String,cache).SetString(fv.String())
				default:
					if fv.CanInterface(){
						newv.SetKeyvalue(sfield.Name,fv.Interface(),cache)
					}
				}
			}
		case reflect.Map:
			mapkeys := reflectv.MapKeys()
			if len(mapkeys) == 0{
				return
			}
			kv := mapkeys[0]
			if kv.Type().Kind() == reflect.Ptr{
				kv = kv.Elem()
			}
			if kv.Kind() != reflect.String{
				return
			}
			newv := v.SetKeyCached(Name,VT_Object,cache)
			for _,kv = range mapkeys{
				rvalue := reflectv.MapIndex(kv)
				if rvalue.Kind() == reflect.Ptr{
					rvalue = rvalue.Elem()
				}
				switch rvalue.Kind() {
				case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
					newv.SetKeyCached(kv.String(),VT_Int,cache).SetInt(rvalue.Int())
				case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
					newv.SetKeyCached(kv.String(),VT_Int,cache).SetInt(int64(rvalue.Uint()))
				case reflect.Bool:
					if rvalue.Bool(){
						newv.SetKeyValue(kv.String(),valueTrue)
					}else{
						newv.SetKeyValue(kv.String(),valueFalse)
					}
				case reflect.Float32:
					newv.SetKeyCached(kv.String(),VT_Double,cache).SetFloat(float32(rvalue.Float()))
				case reflect.Float64:
					newv.SetKeyCached(kv.String(),VT_Double,cache).SetDouble(rvalue.Float())
				case reflect.String:
					newv.SetKeyCached(kv.String(),VT_String,cache).SetString(rvalue.String())
				default:
					if rvalue.CanInterface(){
						newv.SetKeyvalue(kv.String(),rvalue.Interface(),cache)
					}
				}
			}
		case reflect.Slice:
			newv := v.SetKeyCached(Name,VT_Array,cache)
			vlen := reflectv.Len()
			for i := 0;i<vlen;i++{
				av := reflectv.Index(i)
				realvalue := getRealValue(&av)
				if realvalue == nil{
					newv.SetIndexCached(i,VT_NULL,cache)
				}else if realvalue.CanInterface(){
					newv.SetIndexvalue(i,realvalue.Interface(),cache)
				}
			}
		}

	}
}

func (v *DxValue)SetIndexvalue(idx int,value interface{},cache *ValueCache)  {
	switch realv := value.(type) {
	case int:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(realv))
	case *int:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(*realv))
	case uint:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(realv))
	case *uint:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(*realv))
	case int32:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(realv))
	case *int32:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(*realv))
	case uint32:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(realv))
	case *uint32:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(*realv))
	case int16:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(realv))
	case *int16:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(*realv))
	case uint16:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(realv))
	case *uint16:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(*realv))
	case int8:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(realv))
	case *int8:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(*realv))
	case uint8:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(realv))
	case *uint8:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(*realv))
	case int64:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(realv)
	case *int64:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(*realv)
	case uint64:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(realv))
	case *uint64:
		v.SetIndexCached(idx,VT_Int,cache).SetInt(int64(*realv))
	case string:
		v.SetIndexCached(idx,VT_String,cache).SetString(realv)
	case *DxValue:
		v.SetIndexValue(idx,realv)
	case DxValue:
		v.SetIndexValue(idx,&realv)
	case bool:
		if realv{
			v.SetIndexValue(idx,valueTrue)
		}else{
			v.SetIndexValue(idx,valueFalse)
		}
	case *bool:
		if *realv{
			v.SetIndexValue(idx,valueTrue)
		}else{
			v.SetIndexValue(idx,valueFalse)
		}
	case time.Time:
		v.SetIndexCached(idx,VT_DateTime,cache).SetDouble(float64(DxCommonLib.Time2DelphiTime(realv)))
	case *time.Time:
		v.SetIndexCached(idx,VT_DateTime,cache).SetDouble(float64(DxCommonLib.Time2DelphiTime(*realv)))
	case float32:
		v.SetIndexCached(idx,VT_Float,cache).SetFloat(realv)
	case *float32:
		v.SetIndexCached(idx,VT_Float,cache).SetFloat(*realv)
	case float64:
		v.SetIndexCached(idx,VT_Double,cache).SetDouble(realv)
	case *float64:
		v.SetIndexCached(idx,VT_Double,cache).SetDouble(*realv)
	case []byte:
		v.SetIndexCached(idx,VT_Binary,cache).SetBinary(realv,true)
	case *[]byte:
		v.SetIndexCached(idx,VT_Binary,cache).SetBinary(*realv,true)
	case map[string]interface{}:
		newv := v.SetIndexCached(idx,VT_Object,cache)
		for key,objv := range realv{
			newv.SetKeyvalue(key,objv,cache)
		}
	case *map[string]interface{}:
		newv := v.SetIndexCached(idx,VT_Object,cache)
		for key,objv := range *realv{
			newv.SetKeyvalue(key,objv,cache)
		}
	case map[string]string:
		newv := v.SetIndexCached(idx,VT_Object,cache)
		for key,objv := range realv{
			newv.SetKeyCached(key,VT_String,cache).SetString(objv)
		}
	case *map[string]string:
		newv := v.SetIndexCached(idx,VT_Object,v.ownercache)
		for key,objv := range *realv{
			newv.SetKeyCached(key,VT_String,cache).SetString(objv)
		}
	case map[string]int:
		newv := v.SetIndexCached(idx,VT_Object,cache)
		for key,objv := range realv{
			newv.SetKeyCached(key,VT_Int,cache).SetInt(int64(objv))
		}
	case *map[string]int:
		newv := v.SetIndexCached(idx,VT_Object,cache)
		for key,objv := range *realv{
			newv.SetKeyCached(key,VT_Int,cache).SetInt(int64(objv))
		}
	case []interface{}:
		newv := v.SetIndexCached(idx,VT_Array,cache)
		for i := 0;i<len(realv);i++{
			newv.SetIndexvalue(i,realv[i],cache)
		}
	case *[]interface{}:
		newv := v.SetIndexCached(idx,VT_Array,cache)
		for i := 0;i<len(*realv);i++{
			newv.SetIndexvalue(i,(*realv)[i],cache)
		}
	case []string:
		newv := v.SetIndexCached(idx,VT_Array,cache)
		for i := 0;i<len(realv);i++{
			newv.SetIndexCached(i,VT_String,cache).SetString(realv[i])
		}
	case *[]string:
		newv := v.SetIndexCached(idx,VT_Array,cache)
		for i := 0;i<len(*realv);i++{
			newv.SetIndexCached(i,VT_String,cache).SetString((*realv)[i])
		}
	default:
		//判断一下是否是结构体
		reflectv := reflect.ValueOf(value)
		if !reflectv.IsValid(){
			return
		}
		rtype := reflectv.Type()
		if rtype.Implements(ValueMarshalerType){
			value.(DxValueMarshaler).EncodeToDxValue(v.SetIndexCached(idx,VT_Int,cache))
			return
		}
		if reflectv.Kind() == reflect.Ptr{
			reflectv = reflectv.Elem()
			rtype = reflectv.Type()
		}
		switch reflectv.Kind(){
		case reflect.Struct:
			newv := v.SetIndexCached(idx,VT_Object,cache)
			for i := 0;i < rtype.NumField();i++{
				sfield := rtype.Field(i)
				fv := reflectv.Field(i)
				if fv.Kind() == reflect.Ptr{
					fv = fv.Elem()
				}
				switch fv.Kind() {
				case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64,
					reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
					newv.SetKeyCached(sfield.Name,VT_Int,cache).SetInt(fv.Int())
				case reflect.Bool:
					if fv.Bool(){
						newv.SetKeyValue(sfield.Name,valueTrue)
					}else{
						newv.SetKeyValue(sfield.Name,valueFalse)
					}
				case reflect.Float32:
					newv.SetKeyCached(sfield.Name,VT_Double,cache).SetFloat(float32(fv.Float()))
				case reflect.Float64:
					newv.SetKeyCached(sfield.Name,VT_Double,cache).SetDouble(fv.Float())
				case reflect.String:
					newv.SetKeyCached(sfield.Name,VT_String,cache).SetString(fv.String())
				default:
					if fv.CanInterface(){
						newv.SetKeyvalue(sfield.Name,fv.Interface(),cache)
					}
				}
			}
		case reflect.Map:
			mapkeys := reflectv.MapKeys()
			if len(mapkeys) == 0{
				return
			}
			kv := mapkeys[0]
			if kv.Type().Kind() == reflect.Ptr{
				kv = kv.Elem()
			}
			if kv.Kind() != reflect.String{
				return
			}
			newv := v.SetIndexCached(idx,VT_Object,cache)
			for _,kv = range mapkeys{
				rvalue := reflectv.MapIndex(kv)
				if rvalue.Kind() == reflect.Ptr{
					rvalue = rvalue.Elem()
				}
				switch rvalue.Kind() {
				case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
					newv.SetKeyCached(kv.String(),VT_Int,cache).SetInt(rvalue.Int())
				case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
					newv.SetKeyCached(kv.String(),VT_Int,cache).SetInt(int64(rvalue.Uint()))
				case reflect.Bool:
					if rvalue.Bool(){
						newv.SetKeyValue(kv.String(),valueTrue)
					}else{
						newv.SetKeyValue(kv.String(),valueFalse)
					}
				case reflect.Float32:
					newv.SetKeyCached(kv.String(),VT_Double,cache).SetFloat(float32(rvalue.Float()))
				case reflect.Float64:
					newv.SetKeyCached(kv.String(),VT_Double,cache).SetDouble(rvalue.Float())
				case reflect.String:
					newv.SetKeyCached(kv.String(),VT_String,cache).SetString(rvalue.String())
				default:
					if rvalue.CanInterface(){
						newv.SetKeyvalue(kv.String(),rvalue.Interface(),cache)
					}
				}
			}
		case reflect.Slice:
			//数组的话
			newarr := v.SetIndexCached(idx,VT_Array,cache)
			vlen := reflectv.Len()
			for i := 0;i<vlen;i++{
				av := reflectv.Index(i)
				realvalue := getRealValue(&av)
				if realvalue == nil{
					newarr.SetIndexCached(i,VT_NULL,cache)
				}else if realvalue.CanInterface(){
					newarr.SetIndexvalue(i,realvalue.Interface(),cache)
				}
			}
		}

	}
}

//转换到标砖的数据
func (v *DxValue)ToStdValue(destv interface{},ignoreCase bool)bool  {
	switch value := destv.(type) {
	case int,int8,int16,int32,int64,uint,uint8,uint16,uint32,uint64,[]byte,string:
		return false
	case *int8:
		*value = int8(v.Int())
	case *int:
		*value = int(v.Int())
	case *int16:
		*value = int16(v.Int())
	case *int32:
		*value = int32(v.Int())
	case *int64:
		*value = v.Int()
	case *uint8:
		*value = uint8(v.Int())
	case *uint:
		*value = uint(v.Int())
	case *uint16:
		*value = uint16(v.Int())
	case *uint32:
		*value = uint32(v.Int())
	case *uint64:
		*value = uint64(v.Int())
	case *string:
		*value = v.String()
	case *[]byte:
		*value = v.Binary()
	case *bool:
		*value = v.Bool()
	case *time.Time:
		*value = v.GoTime()
	case *float64:
		*value = v.Double()
	case *float32:
		*value = v.Float()
	case map[string]string:
		if v.DataType != VT_Object{
			return false
		}
		v.Visit(func(Key string, mapvalue *DxValue) bool {
			value[Key] = mapvalue.String()
			return true
		})
	case map[string]interface{}:
		if v.DataType != VT_Object{
			return false
		}
		v.Visit(func(Key string, mapvalue *DxValue) bool {
			switch mapvalue.DataType {
			case VT_Object:
				newvalue := make(map[string]interface{},mapvalue.Count())
				mapvalue.ToStdValue(newvalue,ignoreCase)
				value[Key] = newvalue
			case VT_Array:
				newvalue := make([]interface{},mapvalue.Count())
				mapvalue.ToStdValue(newvalue,ignoreCase)
				value[Key] = newvalue
			case VT_String,VT_RawString:
				value[Key] = mapvalue.String()
			case VT_Int:
				value[Key] = mapvalue.Int()
			case VT_DateTime:
				value[Key] = mapvalue.GoTime()
			case VT_Double:
				value[Key] = mapvalue.Double()
			case VT_Float:
				value[Key] = mapvalue.Float()
			case VT_Binary,VT_ExBinary:
				value[Key] = mapvalue.Binary()
			case VT_True:
				value[Key] = true
			case VT_False:
				value[Key] = false
			case VT_NULL:
				value[Key] = nil
			}
			return true
		})
	case []interface{}:
		if v.DataType != VT_Array{
			return false
		}
		v.Visit(func(Key string, arrvalue *DxValue) bool {
			switch arrvalue.DataType {
			case VT_Object:
				newvalue := make(map[string]interface{},arrvalue.Count())
				arrvalue.ToStdValue(newvalue,ignoreCase)
				value = append(value,newvalue)
			case VT_Array:
				newvalue := make([]interface{},arrvalue.Count())
				arrvalue.ToStdValue(newvalue,ignoreCase)
				value = append(value,newvalue)
			case VT_String,VT_RawString:
				value = append(value,arrvalue.String())
			case VT_Int:
				value = append(value,arrvalue.Int())
			case VT_DateTime:
				value = append(value,arrvalue.GoTime())
			case VT_Double:
				value = append(value,arrvalue.Double())
			case VT_Float:
				value = append(value,arrvalue.Float())
			case VT_Binary,VT_ExBinary:
				value = append(value,arrvalue.Binary())
			case VT_True:
				value = append(value,true)
			case VT_False:
				value = append(value,false)
			case VT_NULL:
				value = append(value,nil)
			}
			return true
		})
	case []string:
		if v.DataType != VT_Array{
			return false
		}
		v.Visit(func(Key string, arrvalue *DxValue) bool {
			value = append(value,arrvalue.String())
			return true
		})
	default:
		reflectv := reflect.ValueOf(destv)
		if !reflectv.IsValid(){
			return false
		}
		if reflectv.Kind() != reflect.Ptr{
			return false
		}
		if reflectv.Type().Implements(ValueUnMarshalerType){
			destv.(DxValueUnMarshaler).DecodeFromDxValue(v)
			return true
		}
		reflectv = reflectv.Elem()
		valueType := reflectv.Type()

		if v.DataType != VT_Object{
			return false
		}
		//反射处理,只处理结构体
		vhandler,ok := structTypePool.Load(valueType)
		if ok && vhandler != nil{
			convertHandler := vhandler.(StdValueFromDxValue)
			if convertHandler != nil{
				convertHandler(reflectv,v)
				return true
			}
		}

		switch valueType.Kind() {
		case reflect.Struct:
			decode2reflectFromdxValue(reflectv,v,ignoreCase,valueType)
		case reflect.Slice:
			decodeArray2reflect(reflectv,v,ignoreCase)
		default:
			return false
		}
	}
	return true
}

func Marshal(v interface{}) ([]byte, error) {
	switch value := v.(type) {
	case DxValue:
		return Value2Json(&value,JSE_AllEscape,true,make([]byte,0,128)),nil
	case *DxValue:
		return Value2Json(value,JSE_AllEscape,true,make([]byte,0,128)),nil
	default:
		return json.Marshal(v)
	}
}

func Unmarshal(data []byte, v interface{}) error {
	switch value := v.(type) {
	case DxValue:
		return value.LoadFromJson(data,false)
	case *DxValue:
		return value.LoadFromJson(data,false)
	default:
		return json.Unmarshal(data,v)
	}
}