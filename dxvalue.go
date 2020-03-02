package dxsvalue

import (
	"bytes"
	"github.com/suiyunonghen/DxCommonLib"
	"reflect"
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
	VT_Double
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
	ownercache	*ValueCache
	simpleV		[8]byte
	fstrvalue	string
	fbinary		[]byte		//二进制数据
	farr		[]*DxValue
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

func (v *DxValue)ValueCache()*ValueCache  {
	return v.ownercache
}

func (v *DxValue)Clear()  {
	if v.ownercache != nil{
		v.ownercache.Reset(false) //根不回收
	}
	v.fobject.keysUnescaped = false
	switch v.DataType {
	case VT_Object:
		for i := 0; i < len(v.fobject.strkvs);i++{
			v.fobject.strkvs[i].V = nil
			v.fobject.strkvs[i].K = ""
		}
		v.fobject.strkvs = v.fobject.strkvs[:0]
	case VT_Array:
		v.fobject.strkvs = nil
		for i := 0;i<len(v.farr);i++{
			v.farr[i] = nil
		}
		v.farr = v.farr[:0]
	default:
		v.farr = nil
		v.fobject.strkvs = nil
	}
	v.fbinary = nil
	v.fstrvalue = ""
}

func NewObject(cached bool)*DxValue  {
	var result *DxValue
	if cached{
		c := &ValueCache{
			fisroot:	false,
			Value:    make([]DxValue,0,8),
		}
		c.Value = append(c.Value, DxValue{DataType:   VT_Object,
			ownercache: c,})
		result = &c.Value[len(c.Value)-1]
	}else{
		result = &DxValue{
			DataType:   VT_Object,
			ownercache: nil,
		}
	}
	result.fobject.strkvs =  make([]strkv,0,8)
	return result
}

func NewArray(cached bool)*DxValue  {
	var result *DxValue
	if cached{
		c := &ValueCache{
			fisroot:	false,
			Value:    make([]DxValue,0,8),
		}
		c.Value = append(c.Value, DxValue{DataType:   VT_Object,
			ownercache: c,})
		result = &c.Value[len(c.Value)-1]
	}else{
		result = &DxValue{
			DataType:   VT_Object,
			ownercache: nil,
		}
	}
	result.farr =  make([]*DxValue,0,8)
	return result
}

func NewCacheValue(tp ValueType)*DxValue  {
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
	c := getCache()
	//缓存模式下，会公用这个cacheBuffer
	return c.getValue(tp)
}

func (v *DxValue)LoadFromMsgPack(b []byte,sharebinary bool)error  {
	_,err := parseMsgPack2Value(b,v,sharebinary)
	return err
}

func (v *DxValue)LoadFromJson(b []byte,sharebinary bool)error  {
	_,err := parseJson2Value(b,v,sharebinary)
	return err
}

func (v *DxValue)Reset(dt ValueType)  {
	if v == valueNAN || v == valueINF || v == valueTrue || v == valueFalse || v == valueNull{
		return
	}
	v.DataType = dt
	v.fobject.keysUnescaped = false
	switch dt {
	case VT_Object:
		if v.fobject.strkvs == nil{
			v.fobject.strkvs = make([]strkv,0,8)
		}else{
			for i := 0; i < len(v.fobject.strkvs);i++{
				v.fobject.strkvs[i].V = nil
				v.fobject.strkvs[i].K = ""
			}
			v.fobject.strkvs = v.fobject.strkvs[:0]
		}
		v.farr = nil
	case VT_Array:
		v.fobject.strkvs = nil
		if v.fobject.strkvs == nil{
			v.fobject.strkvs = make([]strkv,0,8)
		}else{
			for i := 0;i<len(v.farr);i++{
				v.farr[i] = nil
			}
			v.farr = v.farr[:0]
		}
	default:
		v.farr = nil
		v.fobject.strkvs = nil
	}
	v.fbinary = nil
	v.fstrvalue = ""
}

func (v *DxValue)Type() ValueType {
	return v.DataType
}

func (v *DxValue)Find(key string)*DxValue  {
	return v.ValueByName(key)
}

func (v *DxValue)StringByName(key string,defstr string)string  {
	return v.AsString(key,defstr)
}

func (v *DxValue)IntByName(key string,def int)int  {
	return v.AsInt(key,def)
}

func (v *DxValue)BoolByName(Key string,defv bool)bool  {
	return v.AsBool(Key,defv)
}

func (v *DxValue)FloatByName(Key string,defv float32)float32  {
	return v.AsFloat(Key,defv)
}

func (v *DxValue)DoubleByName(Key string,defv float64)float64  {
	return v.AsDouble(Key,defv)
}

type MergeOp byte
const(
	MO_Custom MergeOp = iota
	MO_Replace //碰到相同的就替换
	MO_Normal  //执行默认操作，对于object的，碰到相同的就合并并集，Array的也是取并集
)

type MergeFunc	func(keyPath string, oldv *DxValue,newv *DxValue)MergeOp

func mergeObject(parentpath []byte,v,value *DxValue,mergefunc MergeFunc)[]byte  {
	mergeOp := MO_Normal
	for i := 0;i<len(value.fobject.strkvs);i++{
		idx := v.fobject.indexByName(value.fobject.strkvs[i].K)
		if idx != -1{
			if mergefunc!=nil{
				if len(parentpath) != 0{
					parentpath = append(parentpath,'/')
				}
				parentpath = append(parentpath,value.fobject.strkvs[i].K...)
				mergeOp = mergefunc(DxCommonLib.FastByte2String(parentpath),v.fobject.strkvs[idx].V,value.fobject.strkvs[i].V)
			}
			if mergeOp == MO_Normal{
				if v.fobject.strkvs[idx].V.DataType <= VT_Array{
					return mergeObject(parentpath,v.fobject.strkvs[idx].V,value.fobject.strkvs[i].V,mergefunc)
				}else{
					//合并成数组
					oldobj := v.fobject.strkvs[idx].V.clone(nil)
					v.fobject.strkvs[idx].V.Reset(VT_Array)
					v.fobject.strkvs[idx].V.SetIndexValue(0,oldobj)
					v.fobject.strkvs[idx].V.SetIndexValue(1,value.fobject.strkvs[i].V.clone(nil))
				}
			}else if mergeOp == MO_Replace{
				//直接替换
				v.fobject.strkvs[idx].V.Reset(VT_NULL)
				v.fobject.strkvs[idx].V = value.fobject.strkvs[i].V.clone(nil)
			}
		}else{
			//没有，直接增加
			v.SetKeyValue(value.fobject.strkvs[i].K,value.fobject.strkvs[i].V.clone(nil))
		}
	}
	return parentpath
}

func (v *DxValue)MergeWith(value *DxValue,mergefunc MergeFunc)  {
	if value == nil || v == nil || value.DataType != v.DataType || v.DataType > VT_Array{
		return
	}
	if v.DataType == VT_Object{
		mergeObject(make([]byte,0,16),v,value,mergefunc)
		return
	}

	//直接将两个数组合并
	var willAdd []*DxValue
	for i := 0;i<len(value.farr);i++{
		newv := value.farr[i]
		hasFound := false
		for j := 0;j<len(v.farr);j++{
			if newv.Equal(v.farr[j]){
				hasFound = true
				break
			}
		}
		if !hasFound{
			willAdd = append(willAdd,newv.clone(nil))
		}
	}
	if len(willAdd) > 0{
		v.farr = append(v.farr,willAdd...)
	}
}

func (v *DxValue)Equal(value *DxValue)bool  {
	if v == nil && value == nil{
		return true
	}
	if v == nil && value != nil || v != nil && value == nil{
		return false
	}

	if v.DataType != value.DataType {
		if v.DataType == VT_String && value.DataType == VT_RawString ||
			value.DataType == VT_String && v.DataType == VT_RawString{
			return strings.EqualFold(v.String(),value.String())
		}
		if v.DataType >= VT_Int && v.DataType <= VT_Double && value.DataType >= VT_Int && value.DataType <= VT_Double{
			return v.Double() == value.Double()
		}
		return false
	}
	switch v.DataType {
	case VT_String,VT_RawString:
		return v.fstrvalue == value.fstrvalue
	case VT_Binary,VT_ExBinary:
		return bytes.Compare(v.fbinary,value.fbinary) == 0
	case VT_Object:
		if len(v.fobject.strkvs) != len(value.fobject.strkvs){
			return false
		}
		v.fobject.UnEscapestrs()
		value.fobject.UnEscapestrs()
		for i := 0;i<len(v.fobject.strkvs);i++{
			if v.fobject.strkvs[i].K != value.fobject.strkvs[i].K ||
				!v.fobject.strkvs[i].V.Equal(value.fobject.strkvs[i].V){
				return false
			}
		}
	case VT_Array:
		if len(v.farr) != len(v.farr){
			return false
		}
		for i := 0;i<len(v.farr);i++{
			if !v.farr[i].Equal(value.farr[i]){
				return false
			}
		}
	default:
		if v.DataType >= VT_ExBinary{
			return v.DataType == value.DataType
		}
		return bytes.Compare(v.simpleV[:],value.simpleV[:]) == 0
	}
	return true
}

func (v *DxValue)TimeByName(Key string,defv time.Time)time.Time  {
	return v.AsDateTime(Key,DxCommonLib.Time2DelphiTime(&defv)).ToTime()
}

func (v *DxValue)Int()int64  {
	switch v.DataType {
	case VT_True:
		return 1
	case VT_False:
		return 0
	case VT_Int:
		return *((*int64)(unsafe.Pointer(&v.simpleV[0])))
	case VT_Double,VT_DateTime:
		return int64(*((*float64)(unsafe.Pointer(&v.simpleV[0]))))
	case VT_Float:
		return int64(*((*float32)(unsafe.Pointer(&v.simpleV[0]))))
	case VT_String,VT_RawString:
		return DxCommonLib.StrToIntDef(v.fstrvalue,0)
	}
	return 0
}

func (v *DxValue)clone(c *ValueCache)*DxValue  {
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
		rootv.fobject.keysUnescaped = v.fobject.keysUnescaped
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

func (v *DxValue)CopyFrom(fromv *DxValue)  {
	v.Reset(fromv.DataType)
	switch fromv.DataType {
	case VT_Object:
		c := v.ownercache
		v.fobject.keysUnescaped = fromv.fobject.keysUnescaped
		for i := 0; i<len(fromv.fobject.strkvs);i++{
			rkv := fromv.fobject.getKv()
			rkv.K = fromv.fobject.strkvs[i].K
			rkv.V = fromv.fobject.strkvs[i].V.clone(c)
		}
	case VT_Array:
		c := v.ownercache
		for i := 0;i < len(fromv.farr);i++{
			v.farr = append(v.farr, fromv.farr[i].clone(c))
		}
	case VT_String,VT_RawString:
		v.fstrvalue = fromv.fstrvalue
	default:
		if v.DataType >= VT_Int && v.DataType <= VT_DateTime{
			copy(v.simpleV[:],fromv.simpleV[:])
		}
	}
}

func (v *DxValue)Clone(usecache bool)*DxValue  {
	var c *ValueCache
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
	switch v.DataType {
	case VT_Int:
		return strconv.FormatInt(*((*int64)(unsafe.Pointer(&v.simpleV[0]))), 10)
	case VT_True:
		return "true"
	case VT_False:
		return "false"
	case VT_Double:
		vf := *((*float64)(unsafe.Pointer(&v.simpleV[0])))
		return strconv.FormatFloat(vf,'f',-1,64)
	case VT_Float:
		vf := *((*float32)(unsafe.Pointer(&v.simpleV[0])))
		return strconv.FormatFloat(float64(vf),'f',-1,64)
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
		return v.fstrvalue
	case VT_Object,VT_Array:
		return DxCommonLib.FastByte2String(Value2Json(v,false,nil))
	}
	return ""
}

func (v *DxValue)AsString(Name string,Default string)string  {
	return v.StringByPath(Default,Name)
}

func (v *DxValue)AsBool(Name string,def bool)bool  {
	return v.BoolByPath(def,Name)
}

func (v *DxValue)AsInt(Name string,def int)int  {
	return int(v.IntByPath(int64(def),Name))
}

func (v *DxValue)AsInt64(Name string,def int64)int64  {
	return v.IntByPath(int64(def),Name)
}

func (v *DxValue)AsFloat(Name string,def float32)float32  {
	return v.FloatByPath(def,Name)
}

func (v *DxValue)AsDouble(Name string,def float64)float64  {
	return v.DoubleByPath(def,Name)
}

func (v *DxValue)AsDateTime(Name string,def DxCommonLib.TDateTime)DxCommonLib.TDateTime  {
	return v.DateTimeByPath(def,Name)
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

func (v *DxValue)SetIndexTime(idx int,t *time.Time)  {
	v.SetIndex(idx,VT_DateTime).SetDouble(float64(DxCommonLib.Time2DelphiTime(t)))
}

func (v *DxValue)SetKeyInt(Name string,value int64)  {
	v.SetKey(Name,VT_Int).SetInt(value)
}

func (v *DxValue)SetIndexInt(idx int,value int64)  {
	v.SetIndex(idx,VT_Int).SetInt(value)
}

func (v *DxValue)SetKeyFloat(Name string,value float32)  {
	v.SetKey(Name,VT_Float).SetFloat(value)
}

func (v *DxValue)SetIndexFloat(idx int,value float32)  {
	v.SetIndex(idx,VT_Float).SetFloat(value)
}

func (v *DxValue)SetKeyDouble(Name string,value float64)  {
	v.SetKey(Name,VT_Double).SetDouble(value)
}

func (v *DxValue)SetIndexDouble(idx int,value float64)  {
	v.SetIndex(idx,VT_Double).SetDouble(value)
}

func (v *DxValue)SetKeyBool(Name string,value bool)  {
	if value{
		v.SetKeyValue(Name,valueTrue)
	}else{
		v.SetKeyValue(Name,valueFalse)
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
	v.SetKey(Name,VT_DateTime).SetDouble(float64(DxCommonLib.Time2DelphiTime(&value)))
}

func (v *DxValue)Bool()bool  {
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

func (v *DxValue)Double()float64  {
	switch v.DataType {
	case VT_True:
		return 1
	case VT_False:
		return 0
	case VT_Int:
		return float64(*((*int64)(unsafe.Pointer(&v.simpleV[0]))))
	case VT_Double,VT_DateTime:
		return *((*float64)(unsafe.Pointer(&v.simpleV[0])))
	case VT_Float:
		return float64(*((*float32)(unsafe.Pointer(&v.simpleV[0]))))
	case VT_String,VT_RawString:
		return DxCommonLib.StrToFloatDef(v.fstrvalue,0)
	}
	return 0
}

func (v *DxValue)Float()float32  {
	switch v.DataType {
	case VT_True:
		return 1
	case VT_False:
		return 0
	case VT_Int:
		return float32(*((*int64)(unsafe.Pointer(&v.simpleV[0]))))
	case VT_Double,VT_DateTime:
		return float32(*((*float64)(unsafe.Pointer(&v.simpleV[0]))))
	case VT_Float:
		return *((*float32)(unsafe.Pointer(&v.simpleV[0])))
	case VT_String,VT_RawString:
		return float32(DxCommonLib.StrToFloatDef(v.fstrvalue,0))
	}
	return 0
}


func (v *DxValue)DateTime()DxCommonLib.TDateTime  {
	switch v.DataType {
	case VT_Int,VT_Double,VT_Float,VT_DateTime:
		return (DxCommonLib.TDateTime)(v.Double())
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

func (v *DxValue)GoTime()time.Time  {
	switch v.DataType {
	case VT_Int,VT_Float,VT_Double, VT_DateTime:
		return (DxCommonLib.TDateTime)(v.Double()).ToTime()
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

func (v *DxValue)SetDouble(value float64)  {
	if v.DataType != VT_Double{
		v.Reset(VT_Double)
	}
	*((*float64)(unsafe.Pointer(&v.simpleV[0]))) = value
}

func (v *DxValue)SetFloat(value float32)  {
	if v.DataType != VT_Float{
		v.Reset(VT_Float)
	}
	*((*float32)(unsafe.Pointer(&v.simpleV[0]))) = value
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
	return result.String()
}

func (v *DxValue)BoolByPath(DefaultValue bool, paths ...string)bool  {
	result := v.ValueByPath(paths...)
	if result == nil{
		return DefaultValue
	}
	return result.Bool()
}

func (v *DxValue)IntByPath(DefaultValue int64, paths ...string)int64  {
	result := v.ValueByPath(paths...)
	if result == nil{
		return DefaultValue
	}
	return result.Int()
}

func (v *DxValue)FloatByPath(DefaultValue float32, paths ...string)float32  {
	result := v.ValueByPath(paths...)
	if result == nil{
		return DefaultValue
	}
	return result.Float()
}

func (v *DxValue)DoubleByPath(DefaultValue float64, paths ...string)float64  {
	result := v.ValueByPath(paths...)
	if result == nil{
		return DefaultValue
	}
	return result.Double()
}

func (v *DxValue)DateTimeByPath(DefaultValue DxCommonLib.TDateTime, paths ...string)DxCommonLib.TDateTime  {
	result := v.ValueByPath(paths...)
	if result == nil{
		return DefaultValue
	}
	return result.DateTime()
}

func (v *DxValue)GoTimeByPath(DefaultValue time.Time, paths ...string)time.Time  {
	result := v.ValueByPath(paths...)
	if result == nil{
		return DefaultValue
	}
	return result.GoTime()
}

func (v *DxValue)SetKeyCached(Name string,tp ValueType,c *ValueCache)*DxValue  {
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
			result = c.getValue(tp)
			v.fobject.strkvs[idx].V = result
		}else if result.DataType != tp {
			result.Reset(tp)
		}
		return result
	}
	kv := v.fobject.getKv()
	kv.K = Name
	kv.V = c.getValue(tp)
	return kv.V
}

func (v *DxValue)SetKey(Name string,tp ValueType)*DxValue  {
	return v.SetKeyCached(Name,tp,v.ownercache)
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
		v.SetKeyCached(Name,VT_DateTime,cache).SetDouble(float64(DxCommonLib.Time2DelphiTime(&realv)))
	case *time.Time:
		v.SetKeyCached(Name,VT_DateTime,cache).SetDouble(float64(DxCommonLib.Time2DelphiTime(realv)))
	case float32:
		v.SetKeyCached(Name,VT_Float,cache).SetFloat(realv)
	case *float32:
		v.SetKeyCached(Name,VT_Float,cache).SetFloat(*realv)
	case float64:
		v.SetKeyCached(Name,VT_Double,cache).SetDouble(realv)
	case *float64:
		v.SetKeyCached(Name,VT_Double,cache).SetDouble(*realv)
	case []byte:
		v.SetKeyCached(Name,VT_Binary,cache).SetBinary(realv)
	case *[]byte:
		v.SetKeyCached(Name,VT_Binary,cache).SetBinary(*realv)
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
	default:
		//判断一下是否是结构体
		reflectv := reflect.ValueOf(value)
		if !reflectv.IsValid(){
			return
		}
		if reflectv.Kind() == reflect.Ptr{
			reflectv = reflectv.Elem()
		}
		switch reflectv.Kind(){
		case reflect.Struct:
			newv := v.SetKeyCached(Name,VT_Object,cache)
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
				case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64,
					reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
					newv.SetKeyCached(kv.String(),VT_Int,cache).SetInt(rvalue.Int())
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
		}

	}
}

func (v *DxValue)SetKeyValue(Name string,value *DxValue)  {
	if v.DataType == VT_Array{
		idx := DxCommonLib.StrToIntDef(Name,-1)
		if idx != -1{
			v.SetIndexValue(int(idx),value)
			return
		}
	}
	if v.DataType != VT_Object{
		v.Reset(VT_Object)
	}
	idx := v.fobject.indexByName(Name)
	if idx >= 0{
		v.fobject.strkvs[idx].V = value
	}else{
		kv := v.fobject.getKv()
		kv.K = Name
		kv.V = value
	}
}

func (v *DxValue)SetIndexValue(idx int,value *DxValue)  {
	if v.DataType != VT_Array{
		v.Reset(VT_Array)
	}
	l := len(v.farr)
	if idx >= 0 && idx < l{
		v.farr[idx] = value
	}else{
		v.farr = append(v.farr,value)
	}
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
			}else if result.DataType != tp {
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

func (v *DxValue)Count()int  {
	switch v.DataType {
	case VT_Object:
		return len(v.fobject.strkvs)
	case VT_Array:
		return len(v.farr)
	}
	return 0
}

func (v *DxValue)IntByIndex(idx int,def int)int  {
	value := v.ValueByIndex(idx)
	if value != nil{
		return int(value.Int())
	}
	return def
}

func (v *DxValue)StringByIndex(idx int,def string)string  {
	value := v.ValueByIndex(idx)
	if value != nil{
		return value.String()
	}
	return def
}

func (v *DxValue)BoolByIndex(idx int,def bool)bool  {
	value := v.ValueByIndex(idx)
	if value != nil{
		return value.Bool()
	}
	return def
}

func (v *DxValue)FloatByIndex(idx int,def float32)float32  {
	value := v.ValueByIndex(idx)
	if value != nil{
		return value.Float()
	}
	return def
}

func (v *DxValue)DoubleByIndex(idx int,def float64)float64  {
	value := v.ValueByIndex(idx)
	if value != nil{
		return value.Double()
	}
	return def
}

func (v *DxValue)ValueByIndex(idx int)*DxValue{
	switch v.DataType {
	case VT_Object:
		if idx >= 0 && idx < len(v.fobject.strkvs){
			return v.fobject.strkvs[idx].V
		}
	case VT_Array:
		if idx >= 0 && idx < len(v.farr){
			return v.farr[idx]
		}
	}
	return nil
}

func (v *DxValue)KeyNameByIndex(idx int)string  {
	if v.DataType == VT_Object{
		if idx >= 0 && idx < len(v.fobject.strkvs){
			return v.fobject.strkvs[idx].K
		}
	}
	return ""
}

func (v *DxValue)Visit(f func(Key string,value *DxValue) bool)  {
	switch v.DataType {
	case VT_Object:
		for i := 0;i<len(v.fobject.strkvs);i++{
			if !f(v.fobject.strkvs[i].K,v.fobject.strkvs[i].V){
				return
			}
		}
	case VT_Array:
		for i := 0;i<len(v.farr);i++{
			if !f("",v.farr[i]){
				return
			}
		}
	}
}