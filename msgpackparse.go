package dxsvalue

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/suiyunonghen/DxCommonLib"
	"math"
	"time"
	"unsafe"
)

var(
	ErrUnKnownCode	= errors.New("msgpack: readCode Error")
	ErrParseObjLen	= errors.New("msgpack: parse Object length Error")
	ErrParseInt = errors.New("msgpack: parse integer error")
	ErrInvalidateCode = errors.New("msgpack: invalidate msgpack code")
	ErrMapKey = errors.New("currently only string can be a map key")
)

func readCode(b []byte)(MsgPackCode,[]byte,error)  {
	if len(b) > 1{
		return MsgPackCode(b[0]),b[1:],nil
	}
	return CodeUnkonw,nil,ErrUnKnownCode
}

func parseMsgPackObject(code MsgPackCode,b []byte,c *cache)(result *DxValue,tail []byte,err error)  {
	maplen,b,err := parseMapLen(code,b)
	if err != nil{
		return nil,b,err
	}
	if maplen == 0{
		return c.getValue(VT_Object),b,nil
	}
	if !MsgPackCode(b[0]).IsStr(){
		return nil,b,ErrMapKey
	}
	var rawbyte []byte
	var key string
	var value *DxValue
	result = c.getValue(VT_Object)
	for i := 0;i<maplen;i++{
		code,b,err = readCode(b)
		if err != nil{
			FreeValue(result)
			return nil,b,err
		}
		rawbyte,b,err = parseRawStringByte(code,b)
		if err != nil{
			FreeValue(result)
			return nil,b,err
		}
		if c != nil{
			key = DxCommonLib.FastByte2String(rawbyte)
		}else{
			key = string(rawbyte)
		}
		value,b,err = parseMsgPackValue(b,c)
		if err != nil{
			FreeValue(result)
			return nil,b,err
		}
		result.SetKeyValue(key,value)
	}
	tail = b
	return
}


func parseMsgPackArray(code MsgPackCode,b []byte,c *cache)(result *DxValue,tail []byte,err error){
	arrlen,b,err := parseArrLen(code,b)
	if err != nil{
		return nil,b,err
	}
	if arrlen == 0{
		return c.getValue(VT_Array),b,nil
	}
	var value *DxValue
	result = c.getValue(VT_Array)
	for i := 0;i<arrlen;i++{
		value,b,err = parseMsgPackValue(b,c)
		if err != nil{
			FreeValue(result)
			return nil,b,err
		}
		result.farr = append(result.farr,value)
	}
	return result,b,err
}

func parseMsgPackValue(b []byte,c *cache)(result *DxValue,tail []byte,err error)  {
	code,b,err := readCode(b)
	if err != nil{
		return nil,b,err
	}
	switch  {
	case code.IsStr():
		return parseString(code,b,c)
	case code.IsArray():
		return parseMsgPackArray(code,b,c)
	case code.IsMap():
		return parseMsgPackObject(code,b,c)
	case code.IsFixedNum():
		result = c.getValue(VT_Int)
		result.SetInt(int64(int8(code)))
		return
	case code == CodeUint8:
		ub,tail,err := parseUint8(b)
		if err != nil{
			return nil,tail,err
		}
		result = c.getValue(VT_Int)
		result.SetInt(int64(ub))
		return result,tail,nil
	case code == CodeInt8:
		ub,tail,err := parseUint8(b)
		if err != nil{
			return nil,tail,err
		}
		result = c.getValue(VT_Int)
		result.SetInt(int64(int8(ub)))
		return result,tail,nil
	case code == CodeUint16:
		u16,tail,err := parseUint16(b)
		if err != nil{
			return nil,tail,err
		}
		result = c.getValue(VT_Int)
		result.SetInt(int64(u16))
		return result,tail,nil
	case code == CodeInt16:
		u16,tail,err := parseUint16(b)
		if err != nil{
			return nil,tail,err
		}
		result = c.getValue(VT_Int)
		result.SetInt(int64(int16(u16)))
		return result,tail,nil
	case code == CodeUint32:
		u32,tail,err := parseUint32(b)
		if err != nil{
			return nil,tail,err
		}
		result = c.getValue(VT_Int)
		result.SetInt(int64(u32))
		return result,tail,nil
	case code == CodeInt32:
		u32,tail,err := parseUint32(b)
		if err != nil{
			return nil,tail,err
		}
		result = c.getValue(VT_Int)
		result.SetInt(int64(int32(u32)))
		return result,tail,nil
	case code == CodeInt64 || code == CodeUint64:
		u64,tail,err := parseUint64(b)
		if err != nil{
			return nil,tail,err
		}
		result = c.getValue(VT_Int)
		result.SetInt(int64(u64))
		return result,tail,nil
	case code == CodeFloat:
		u32,tail,err := parseUint32(b)
		if err != nil{
			return nil,tail,err
		}
		result = c.getValue(VT_Float)
		result.SetFloat(*(*float32)(unsafe.Pointer(&u32)))
		return result,tail,nil
	case code == CodeDouble:
		u64,tail,err := parseUint64(b)
		if err != nil{
			return nil,tail,err
		}
		result = c.getValue(VT_Double)
		result.SetDouble(*(*float64)(unsafe.Pointer(&u64)))
		return result,tail,nil
	case code == CodeTrue:
		result = valueTrue
		tail = b
		return
	case code == CodeFalse:
		result = valueFalse
		tail = b
		return
	case code == CodeNil:
		result = valueNull
		tail = b
		return
	case code.IsBin():
		//二进制
		blen,tail,err := parseLen(code,b)
		if err != nil{
			return nil,tail,err
		}
		haslen := len(b)
		if blen <= haslen{
			tail = b[blen:]
			result = c.getValue(VT_Binary)
			if c != nil{
				result.fbinary = b[: blen]
			}else{
				result.fbinary = append(result.fbinary,b[:blen]...)
			}
			return result,tail,nil
		}
		return nil,b, fmt.Errorf("msgpack: binay data truncated,totalen=%d,realLen=%d",blen,haslen)
	case code.IsExt():
		//扩展，需要判定一下这个扩展是否是日期时间格式
		exlen,b,err := parseExtLen(code,b)
		if err != nil{
			return result,tail,err
		}
		haslen := len(b)
		if haslen < exlen{
			return nil,b, fmt.Errorf("msgpack: binay data truncated,totalen=%d,realLen=%d",exlen,haslen)
		}
		tail = b[exlen+1:]
		if code.IsTime(b[0]){
			//日期时间
			result = c.getValue(VT_DateTime)
			switch code {
			case CodeFixExt4:
				sec := binary.BigEndian.Uint32(b[1:])
				t := time.Unix(int64(sec), 0)
				result.SetDouble(float64(DxCommonLib.Time2DelphiTime(&t)))
			case CodeFixExt8:
				//64位时间格式
				sec := binary.BigEndian.Uint64(b[1:])
				nsec := int64(sec >> 34)
				sec &= 0x00000003ffffffff
				t := time.Unix(int64(sec), nsec)
				result.SetDouble(float64(DxCommonLib.Time2DelphiTime(&t)))
			default:
				nsec := binary.BigEndian.Uint32(b[1:])
				sec := binary.BigEndian.Uint64(b[5:])
				t := time.Unix(int64(sec), int64(nsec))
				result.SetDouble(float64(DxCommonLib.Time2DelphiTime(&t)))
			}
		}else{
			result = c.getValue(VT_ExBinary)
			if c != nil{
				result.fbinary = b[:exlen+1] //多一个extype
			}else{
				result.fbinary = append(result.fbinary,b[:exlen+1]...)
			}
		}
	}
	return
}

func NewValueFromMsgPack(b []byte,useCache bool)(*DxValue,error)  {
	var c *cache
	if !useCache{
		c = nil
	}else{
		c = getCache()
		//缓存模式下，会公用这个cacheBuffer
		c.cacheBuffer = append(c.cacheBuffer[:0],b...)
		b = c.cacheBuffer
	}
	v, _, err := parseMsgPackValue(b,c)
	if err != nil {
		return nil, err
	}
	return v,nil
}

func Value2MsgPack(v *DxValue,dst []byte)[]byte  {
	if dst == nil{
		dst = make([]byte,0,256)
	}
	if v == nil{
		dst = append(dst,byte(CodeNil))
		return dst
	}
	switch v.DataType {
	case VT_Object:
		l := len(v.fobject.strkvs)
		dst = writeMapCode(l,dst)
		for i := 0;i < l;i++{
			kv := v.fobject.strkvs[i]
			l := len(kv.K)
			dst = writeStrCode(l,dst)
			dst = append(dst,DxCommonLib.FastString2Byte(kv.K)...)
			dst = Value2MsgPack(kv.V,dst)
		}
	case VT_Array:
		l := len(v.farr)
		dst = writeArrayCode(l,dst)
		for i := 0;i<l;i++{
			dst = Value2MsgPack(v.farr[i],dst)
		}
	case VT_DateTime:
		t := v.AsGoTime()
		secs := uint64(t.Unix())
		if secs>>34 == 0 {
			data := uint64(t.Nanosecond())<<34 | secs
			if data&0xffffffff00000000 == 0 {
				//先写入code，以及exttype
				//CodeFixExt4
				dst = append(dst, byte(CodeFixExt4),0xff)
				l := len(dst)
				dst = append(dst,0,0,0,0)
				b := dst[l:l+4]
				binary.BigEndian.PutUint32(b, uint32(data))
				return dst
			}
			dst = append(dst, byte(CodeFixExt8),0xff)
			l := len(dst)
			dst = append(dst,0,0,0,0,0,0,0,0)
			b := dst[l:l+8]
			binary.BigEndian.PutUint64(b, data)
			return dst
		}
		//96
		dst = append(dst, byte(CodeExt8),0xff)
		l := len(dst)
		dst = append(dst,0,0,0,0,0,0,0,0,0,0,0,0)
		b := dst[l:l+12]
		binary.BigEndian.PutUint32(b, uint32(t.Nanosecond()))
		binary.BigEndian.PutUint64(b[4:], secs)
		return dst
	case VT_ExBinary:
		blen := 0
		if len(v.fbinary) > 0{
			blen = len(v.fbinary[1:])
		}
		dst = writeExtCode(blen,dst)
		//直接写入内容
		dst = append(dst,v.fbinary...)
	case VT_Binary:
		blen := len(v.fbinary)
		dst = writeBinCode(blen,dst)
		dst = append(dst,v.fbinary...)
	case VT_Float:
		u32 := math.Float32bits(v.AsFloat())
		dst = append(dst,byte(CodeFloat),byte(u32 >> 24),byte(u32 >> 16),byte(u32 >> 8),byte(u32))
	case VT_Double:
		u64 := math.Float64bits(v.AsDouble())
		dst = append(dst,byte(CodeDouble),byte(u64 >> 56),byte(u64 >> 48),byte(u64 >> 40),byte(u64 >> 32),byte(u64 >> 24),byte(u64 >> 16),byte(u64 >> 8),byte(u64))
	case VT_Int:
		dst = writeInt(v.AsInt(),dst)
	case VT_String,VT_RawString:
		l := len(v.fstrvalue)
		dst = writeStrCode(l,dst)
		dst = append(dst,DxCommonLib.FastString2Byte(v.fstrvalue)...)
	case VT_False:
		dst = append(dst,byte(CodeFalse))
	case VT_True:
		dst = append(dst,byte(CodeTrue))
	case VT_NULL:
		dst = append(dst,byte(CodeNil))
	}
	return dst
}