package dxsvalue

import (
	"github.com/suiyunonghen/DxCommonLib"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type  ValueType uint8
const(
	VT_Object ValueType=iota
	VT_Array
	VT_String
	VT_RawString //未正常转义的字符串,Json解析之后是这样的
	VT_Int
	VT_Float
	VT_DateTime
	VT_Binary		//二进制
	VT_ExBinary		//MsgPack的扩展二进制
	VT_NULL
	VT_True
	VT_False
	VT_NAN
	VT_INF
)

type  strkv struct {
	K			string
	V			*DxValue
}


type VObject struct {
	strkvs			[]strkv
	keysUnescaped	bool
}

func (obj *VObject)getKv()*strkv  {
	kvlen := len(obj.strkvs)
	if cap(obj.strkvs) > kvlen {
		obj.strkvs = obj.strkvs[:kvlen+1]
	} else {
		obj.strkvs = append(obj.strkvs, strkv{})
	}
	return &obj.strkvs[len(obj.strkvs)-1]
}

func (obj *VObject)Remove(Name string)  {
	if !obj.keysUnescaped && strings.IndexByte(Name, '\\') < 0 {
		for i := 0;i<len(obj.strkvs);i++{
			if obj.strkvs[i].K == Name{
				obj.strkvs = append(obj.strkvs[:i],obj.strkvs[i+1:]...)
				return
			}
		}
	}
	//解转义
	obj.UnEscapestrs()
	for i := 0;i<len(obj.strkvs);i++{
		if obj.strkvs[i].K == Name{
			obj.strkvs = append(obj.strkvs[:i],obj.strkvs[i+1:]...)
			return
		}
	}
}

func (obj *VObject)ExtractValue(Name string)*DxValue  {
	if !obj.keysUnescaped && strings.IndexByte(Name, '\\') < 0 {
		for i := 0;i<len(obj.strkvs);i++{
			if obj.strkvs[i].K == Name{
				result := obj.strkvs[i].V
				obj.strkvs = append(obj.strkvs[:i],obj.strkvs[i+1:]...)
				return result
			}
		}
	}
	//解转义
	obj.UnEscapestrs()
	for i := 0;i<len(obj.strkvs);i++{
		if obj.strkvs[i].K == Name{
			result := obj.strkvs[i].V
			obj.strkvs = append(obj.strkvs[:i],obj.strkvs[i+1:]...)
			return result
		}
	}
	return nil
}

func (obj *VObject)indexByName(name string)int  {
	if !obj.keysUnescaped && strings.IndexByte(name, '\\') < 0 {
		for i := 0;i<len(obj.strkvs);i++{
			if obj.strkvs[i].K == name{
				return i
			}
		}
	}
	//解转义
	obj.UnEscapestrs()
	for i := 0;i<len(obj.strkvs);i++{
		if obj.strkvs[i].K == name{
			return i
		}
	}
	return  -1
}

func (obj *VObject) ValueByName(name string) *DxValue {
	if !obj.keysUnescaped && strings.IndexByte(name, '\\') < 0 {
		for i := 0;i<len(obj.strkvs);i++{
			if obj.strkvs[i].K == name{
				return obj.strkvs[i].V
			}
		}
	}
	//解转义
	obj.UnEscapestrs()
	for i := 0;i<len(obj.strkvs);i++{
		if obj.strkvs[i].K == name{
			return obj.strkvs[i].V
		}
	}
	return nil
}

func (obj *VObject)UnEscapestrs()  {
	if obj.keysUnescaped {
		return
	}
	for i := range obj.strkvs {
		kv := &obj.strkvs[i]
		kv.K = DxCommonLib.ParserEscapeStr(DxCommonLib.FastString2Byte(kv.K))
	}
	obj.keysUnescaped = true
}

type	DxValue struct {
	DataType	ValueType
	fobject		VObject
	ownercache	*cache
	farr		[]*DxValue
	simpleV		[8]byte
	fstrvalue	string
}

func NewValue(tp ValueType)*DxValue  {
	switch tp {
	case VT_True:
		return valueTrue
	case VT_False:
		return valueFalse
	case VT_NULL:
		return valueNull
	case VT_INF:
		return valueINF
	case VT_NAN:
		return valueNAN
	}
	result := &DxValue{}
	result.Reset(tp)
	return result
}

func (v *DxValue)Reset(dt ValueType)  {
	if v == valueNAN || v == valueINF || v == valueTrue || v == valueFalse || v == valueNull{
		return
	}
	v.fobject.keysUnescaped = false
	if v.DataType != dt{
		switch v.DataType {
		case VT_Array:
			for i := 0;i<len(v.farr);i++{
				v.farr[i] = nil
			}
			v.farr = v.farr[:0]
		case VT_Object:
			for i := 0; i < len(v.fobject.strkvs);i++{
				v.fobject.strkvs[i].V = nil
			}
			v.fobject.strkvs = v.fobject.strkvs[:0]
		}
	}
	v.DataType = dt
	switch dt {
	case VT_Object:
		if v.fobject.strkvs == nil{
			v.fobject.strkvs = make([]strkv,0,32)
		}
		v.farr = nil
	case VT_Array:
		v.fobject.strkvs = nil
		if v.farr == nil{
			v.farr = make([]*DxValue,0,32)
		}
	default:
		v.farr = nil
		v.fobject.strkvs = nil
	}
	v.fstrvalue = ""
	DxCommonLib.ZeroByteSlice(v.simpleV[:])
}

func (v *DxValue)Type() ValueType {
	return v.DataType
}

func (v *DxValue)AsInt()int64  {
	switch v.DataType {
	case VT_True:
		return 1
	case VT_False:
		return 0
	case VT_Int:
		return *((*int64)(unsafe.Pointer(&v.simpleV[0])))
	case VT_Float,VT_DateTime:
		return int64(*((*float64)(unsafe.Pointer(&v.simpleV[0]))))
	case VT_String,VT_RawString:
		return DxCommonLib.StrToIntDef(v.fstrvalue,0)
	}
	return 0
}

func (v *DxValue)clone(c *cache)*DxValue  {
	if v.DataType >= VT_True{ //系统固定的值不变更
		return v
	}
	rootv := c.getValue(v.DataType)
	if v.DataType >= VT_Int && v.DataType <= VT_DateTime{
		copy(rootv.simpleV[:],v.simpleV[:])
		return rootv
	}
	switch v.DataType {
	case VT_String,VT_RawString:
		rootv.fstrvalue = v.fstrvalue
	case VT_Object:
		for i := 0; i<len(v.fobject.strkvs);i++{
			rkv := rootv.fobject.getKv()
			rkv.K = v.fobject.strkvs[i].K
			rkv.V = v.fobject.strkvs[i].V.clone(c)
		}
	case VT_Array:
		for i := 0;i < len(v.farr);i++{
			rootv.farr = append(rootv.farr, v.farr[i].clone(c))
		}
	}
	return rootv
}

func (v *DxValue)Clone(usecache bool)*DxValue  {
	var c *cache
	if usecache{
		c = getCache()
	}else{
		c = nil
	}
	return v.clone(c)
}

func (v *DxValue)RemoveKey(Name string)  {
	if v.DataType == VT_Object{
		v.fobject.Remove(Name)
	}
}

func (v *DxValue)ExtractValue(Name string)*DxValue  {
	if v.DataType == VT_Object{
		return v.fobject.ExtractValue(Name)
	}
	return nil
}

func (v *DxValue)RemoveIndex(idx int)  {
	if v.DataType == VT_Array && idx >= 0 && idx < len(v.farr){
		v.farr = append(v.farr[:idx],v.farr[idx+1:]...)
	}
}

func (v *DxValue)ForcePath(vt ValueType,paths ...string)*DxValue  {
	curv := v
	l := len(paths)
	var llastv,lastv *DxValue
	for i := 0;i<l - 1;i++{
		key := paths[i]
		if lastv != nil{
			 llastv = lastv
		}
		lastv = curv
		curv = curv.ValueByName(key)
		if curv == nil || curv.DataType > VT_Array{
			switch lastv.DataType {
			case VT_Array:
				if idx := DxCommonLib.StrToIntDef(key,-1);idx < 0{
					lastv.Reset(VT_Object)
					curv = lastv.SetKey(key,VT_Object)
				}else{
					curv = lastv.SetIndex(int(idx),VT_Object)
				}
			case VT_Object:
				curv = lastv.SetKey(key,VT_Object)
			default:
				if lastv == valueNull || lastv == valueTrue || lastv == valueFalse || lastv == valueINF || lastv == valueNAN{
					if llastv == nil{
						v = NewValue(VT_Object)
						lastv = nil
						curv = v
						continue
					}else{
						lastv = llastv.SetKey(paths[i-1],VT_Object)
					}
				}else{
					lastv.Reset(VT_Object)
				}
				curv = lastv.SetKey(key,VT_Object)
			}
		}
	}
	if l > 0{
		return curv.SetKey(paths[l-1],vt)
	}
	return curv
}

func (v *DxValue)SetInt(value int64)  {
	switch v.DataType {
	case VT_Int:
		*((*int64)(unsafe.Pointer(&v.simpleV[0]))) = value
	case VT_String:
		v.fstrvalue = strconv.FormatInt(value,10)
	default:
		v.Reset(VT_Int)
		*((*int64)(unsafe.Pointer(&v.simpleV[0]))) = value
	}
}

func (v *DxValue)String()string  {
	return v.AsString()
}

func (v *DxValue)AsString()string  {
	switch v.DataType {
	case VT_Int:
		return strconv.FormatInt(*((*int64)(unsafe.Pointer(&v.simpleV[0]))), 10)
	case VT_True:
		return "true"
	case VT_False:
		return "false"
	case VT_Float:
		vf := *((*float64)(unsafe.Pointer(&v.simpleV[0])))
		return strconv.FormatFloat(vf,'f',-1,64)
	case VT_DateTime:
		dt := *((*DxCommonLib.TDateTime)(unsafe.Pointer(&v.simpleV[0])))
		return dt.ToTime().Format("2006-01-02 15:04:05")
	case VT_String:
		return v.fstrvalue
	case VT_NULL:
		return "null"
	case VT_RawString:
		v.DataType = VT_String
		v.fstrvalue = DxCommonLib.ParserEscapeStr(DxCommonLib.FastString2Byte(v.fstrvalue))
	case VT_Object,VT_Array:
		return DxCommonLib.FastByte2String(Value2Json(v,false,nil))
	}
	return ""
}

func (v *DxValue)SetString(value string)  {
	if v.DataType != VT_String{
		v.Reset(VT_String)
	}
	v.fstrvalue = value
}

func (v *DxValue)SetKeyString(Name,value string)  {
	v.SetKey(Name,VT_String).fstrvalue = value
}

func (v *DxValue)SetIndexString(idx int,value string)  {
	v.SetIndex(idx,VT_String).fstrvalue = value
}

func (v *DxValue)SetKeyInt(Name string,value int64)  {
	v.SetKey(Name,VT_Int).SetInt(value)
}

func (v *DxValue)SetIndexInt(idx int,value int64)  {
	v.SetIndex(idx,VT_Int).SetInt(value)
}

func (v *DxValue)SetKeyFloat(Name string,value float64)  {
	v.SetKey(Name,VT_Float).SetFloat(value)
}

func (v *DxValue)SetIndexFloat(idx int,value float64)  {
	v.SetIndex(idx,VT_Float).SetFloat(value)
}

func (v *DxValue)SetKeyBool(Name string,value bool)  {
	if value{
		v.SetKey(Name,VT_True)
	}else{
		v.SetKey(Name,VT_False)
	}
}

func (v *DxValue)SetIndexBool(idx int,value bool)  {
	if value{
		v.SetIndex(idx,VT_True)
	}else{
		v.SetIndex(idx,VT_False)
	}
}

func (v *DxValue)SetKeyTime(Name string,value time.Time)  {
	v.SetKey(Name,VT_DateTime).SetFloat(float64(DxCommonLib.Time2DelphiTime(&value)))
}

func (v *DxValue)SetIndexTime(idx int,value time.Time)  {
	v.SetIndex(idx,VT_DateTime).SetFloat(float64(DxCommonLib.Time2DelphiTime(&value)))
}

func (v *DxValue)AsBool()bool  {
	if v.DataType >= VT_Int && v.DataType <= VT_DateTime{
		return *((*int64)(unsafe.Pointer(&v.simpleV[0]))) > 0
	}
	switch v.DataType {
	case VT_True:
		return true
	case VT_String,VT_RawString:
		return strings.EqualFold(v.fstrvalue,"true")
	}
	return false
}

func (v *DxValue)SetBool(value bool)  {
	if v.DataType >= VT_True{
		return
	}
	if v.DataType >= VT_Int{
		if value{
			*((*int64)(unsafe.Pointer(&v.simpleV[0]))) = 1
		}else{
			*((*int64)(unsafe.Pointer(&v.simpleV[0]))) = 0
		}
	}
}

func (v *DxValue)AsFloat()float64  {
	switch v.DataType {
	case VT_True:
		return 1
	case VT_False:
		return 0
	case VT_Int:
		return float64(*((*int64)(unsafe.Pointer(&v.simpleV[0]))))
	case VT_Float,VT_DateTime:
		return *((*float64)(unsafe.Pointer(&v.simpleV[0])))
	case VT_String,VT_RawString:
		return DxCommonLib.StrToFloatDef(v.fstrvalue,0)
	}
	return 0
}

func (v *DxValue)AsDateTime()DxCommonLib.TDateTime  {
	switch v.DataType {
	case VT_Int,VT_Float,VT_DateTime:
		return (DxCommonLib.TDateTime)(v.AsFloat())
	case VT_String,VT_RawString:
		if t,err := time.Parse("2006-01-02T15:04:05Z",v.fstrvalue);err == nil{
			return DxCommonLib.Time2DelphiTime(&t)
		}else if t,err = time.Parse("2006-01-02 15:04:05",v.fstrvalue);err == nil{
			return DxCommonLib.Time2DelphiTime(&t)
		}else if t,err = time.Parse("2006/01/02 15:04:05",v.fstrvalue);err == nil{
			return DxCommonLib.Time2DelphiTime(&t)
		}
	}
	return -1
}

func (v *DxValue)AsGoTime()time.Time  {
	switch v.DataType {
	case VT_Int,VT_Float,VT_DateTime:
		return (DxCommonLib.TDateTime)(v.AsFloat()).ToTime()
	case VT_String,VT_RawString:
		if t,err := time.Parse("2006-01-02T15:04:05Z",v.fstrvalue);err == nil{
			return t
		}else if t,err = time.Parse("2006-01-02 15:04:05",v.fstrvalue);err == nil{
			return t
		}else if t,err = time.Parse("2006/01/02 15:04:05",v.fstrvalue);err == nil{
			return t
		}
	}
	return time.Time{}
}

func (v *DxValue)SetFloat(value float64)  {
	if v.DataType != VT_Float{
		v.Reset(VT_Float)
	}
	*((*float64)(unsafe.Pointer(&v.simpleV[0]))) = value
}

func (v *DxValue)AsObject()*VObject  {
	if v.DataType == VT_Object{
		return &v.fobject
	}
	return nil
}

func (v *DxValue)ValueByName(Name string)*DxValue  {
	switch v.DataType {
	case VT_Object:
		return v.fobject.ValueByName(Name)
	case VT_Array:
		idx := int(DxCommonLib.StrToIntDef(Name,-1))
		curarrlen := len(v.farr)
		if idx < 0 || idx > curarrlen - 1{
			return nil
		}
		return v.farr[idx]
	default:
		return nil
	}
}

func (v *DxValue)ValueByPath(paths ...string)*DxValue  {
	if v == nil{
		return nil
	}
	curv := v
	for i := 0;i<len(paths);i++{
		key := paths[i]
		curv = curv.ValueByName(key)
		if curv == nil{
			return nil
		}
	}
	return curv
}

//传递路径节点数组
func (v *DxValue)StringByPath(DefaultValue string, paths ...string)string  {
	result := v.ValueByPath(paths...)
	if result == nil{
		return DefaultValue
	}
	return result.AsString()
}

func (v *DxValue)BoolByPath(DefaultValue bool, paths ...string)bool  {
	result := v.ValueByPath(paths...)
	if result == nil{
		return DefaultValue
	}
	return result.AsBool()
}

func (v *DxValue)IntByPath(DefaultValue int64, paths ...string)int64  {
	result := v.ValueByPath(paths...)
	if result == nil{
		return DefaultValue
	}
	return result.AsInt()
}

func (v *DxValue)FloatByPath(DefaultValue float64, paths ...string)float64  {
	result := v.ValueByPath(paths...)
	if result == nil{
		return DefaultValue
	}
	return result.AsFloat()
}

func (v *DxValue)DateTimeByPath(DefaultValue DxCommonLib.TDateTime, paths ...string)DxCommonLib.TDateTime  {
	result := v.ValueByPath(paths...)
	if result == nil{
		return DefaultValue
	}
	return result.AsDateTime()
}

func (v *DxValue)GoTimeByPath(DefaultValue time.Time, paths ...string)time.Time  {
	result := v.ValueByPath(paths...)
	if result == nil{
		return DefaultValue
	}
	return result.AsGoTime()
}

func (v *DxValue)SetKey(Name string,tp ValueType)*DxValue  {
	if v.DataType == VT_Array{
		idx := DxCommonLib.StrToIntDef(Name,-1)
		if idx != -1{
			return v.SetIndex(int(idx),tp)
		}
	}
	if v.DataType != VT_Object{
		v.Reset(VT_Object)
	}
	idx := v.fobject.indexByName(Name)
	if idx >= 0{
		result := v.fobject.strkvs[idx].V
		if result == valueTrue || result == valueFalse || result == valueINF || result == valueNAN || result == valueNull{
			result = NewValue(tp)
			v.fobject.strkvs[idx].V = result
		}else{
			result.Reset(tp)
		}
		return result
	}
	kv := v.fobject.getKv()
	kv.K = Name
	kv.V = NewValue(tp)
	return kv.V
}

func (v *DxValue)SetIndex(idx int,tp ValueType)*DxValue  {
	if v.DataType != VT_Array{
		v.Reset(VT_Array)
	}
	l := len(v.farr)
	if idx >= 0 && idx < l{
		result := v.farr[idx]
		if result != nil && result.DataType != tp{
			if result == valueTrue || result == valueNull || result == valueFalse || result == valueINF || result == valueNAN{
				result = NewValue(tp)
				v.farr[idx] = result
			}else{
				result.Reset(tp)
			}
		}else if result == nil{
			result = NewValue(tp)
			v.farr[idx] = result
		}
		return result
	}else{
		result := NewValue(tp)
		v.farr = append(v.farr,result)
		return result
	}
}

func (v *DxValue)InsertValue(idx int,tp ValueType)*DxValue  {
	if v.DataType != VT_Array{
		v.Reset(VT_Array)
	}
	result := NewValue(tp)
	l := len(v.farr)
	if idx < 0{
		v.farr = append(v.farr[:0],result)
	}else if idx < l{
		rarr := v.farr[idx:]
		v.farr = append(v.farr[:idx],result)
		v.farr = append(v.farr,rarr...)
	}else{
		v.farr = append(v.farr,result)
	}
	return result
}