package dxsvalue

import (
	"fmt"
	"github.com/suiyunonghen/DxCommonLib"
)

func parseUint8(b []byte)(byte,[]byte,error)  {
	if len(b) > 0{
		return b[0],b[1:],nil
	}
	return 0,b,ErrParseInt
}

func parseUint16(b []byte)(uint16,[]byte,error)  {
	if len(b) > 1{
		return (uint16(b[0]) << 8) | uint16(b[1]), b[2:], nil
	}
	return 0,b,ErrParseInt
}

func parseUint32(b []byte)(uint32,[]byte,error)  {
	if len(b) > 3{
		return (uint32(b[0]) << 24) | (uint32(b[1]) << 16) | (uint32(b[2]) << 8) | uint32(b[3]), b[4:], nil
	}
	return 0,b,ErrParseInt
}

func parseUint64(b []byte)(uint64,[]byte,error)  {
	if len(b) > 7{
		return (uint64(b[0]) << 56) | (uint64(b[1]) << 48) | (uint64(b[2]) << 40) | (uint64(b[3]) << 32) |
			(uint64(b[4]) << 24) | (uint64(b[5]) << 16) | (uint64(b[6]) << 8) | uint64(b[7]), b[8:], nil
	}
	return 0,b,ErrParseInt
}

func parseLen(code MsgPackCode,b []byte)(int,[]byte,error)  {
	if code.IsFixedStr(){
		return int(code & CodeFixedStrMask),b, nil
	}
	switch code {
	case CodeStr8, CodeBin8:
		l,b,err := parseUint8(b)
		return int(l),b, err
	case CodeStr16, CodeBin16:
		l,b,err := parseUint16(b)
		return int(l),b, err
	case CodeStr32, CodeBin32:
		l,b,err := parseUint32(b)
		return int(l),b, err
	}
	return 0,b,ErrParseObjLen
}

func parseExtLen(code MsgPackCode,b []byte)(int ,[]byte,error)  {
	switch code {
	case CodeFixExt1:
		return 1,b, nil
	case CodeFixExt2:
		return 2,b, nil
	case CodeFixExt4:
		return 4,b, nil
	case CodeFixExt8:
		return 8,b, nil
	case CodeFixExt16:
		return 16,b, nil
	case CodeExt8:
		n,b, err := parseUint8(b)
		return int(n),b, err
	case CodeExt16:
		n, b,err := parseUint16(b)
		return int(n),b, err
	case CodeExt32:
		n,b, err := parseUint32(b)
		return int(n),b, err
	default:
		return 0,b, fmt.Errorf("msgpack: invalid code=%x decoding ext length", code)
	}
}

func parseMapLen(code MsgPackCode, b []byte)(int ,[]byte,error)  {
	if code >= CodeFixedMapLow && code <= CodeFixedMapHigh {
		return int(code & CodeFixedMapMask), b,nil
	}
	switch code {
	case CodeMap16:
		u16,b,err := parseUint16(b)
		if err != nil{
			err = ErrParseObjLen
		}
		return int(u16),b,err
	case CodeMap32:
		u32,b,err := parseUint32(b)
		if err != nil{
			err = ErrParseObjLen
		}
		return int(u32),b,err
	}
	return 0,b,ErrInvalidateCode
}

func parseArrLen(code MsgPackCode, b []byte)(int ,[]byte,error)  {
	if code >= CodeFixedArrayLow && code <= CodeFixedArrayHigh {
		return int(code & CodeFixedArrayMask), b,nil
	}
	switch code {
	case CodeArray16:
		n,b, err := parseUint16(b)
		return int(n),b, err
	case CodeArray32:
		n,b, err := parseUint32(b)
		return int(n),b, err
	}
	return 0,b,ErrInvalidateCode
}

func parseRawStringByte(code MsgPackCode,b []byte)(result []byte,tail []byte,err error)  {
	stlen,b,err := parseLen(code,b)
	if err != nil{
		return nil,b,ErrParseObjLen
	}
	if stlen <= 0 {
		return nil,b, nil
	}
	haslen := len(b)
	if haslen >= stlen{
		return b[:stlen],b[stlen:],nil
	}
	return nil,b,fmt.Errorf("msgpack: string data truncated,totalen=%d,realLen=%d",stlen,haslen)
}

func parseString(code MsgPackCode,b []byte,c *cache)(result *DxValue,tail []byte,err error)  {
	stlen,b,err := parseLen(code,b)
	if err != nil{
		return nil,b,ErrParseObjLen
	}
	if stlen <= 0 {
		result = c.getValue(VT_String)
		return result,b, nil
	}
	haslen := len(b)
	if haslen >= stlen{
		result = c.getValue(VT_String)
		result.fstrvalue = DxCommonLib.FastByte2String(b[:stlen])
		return result,b[stlen:],nil
	}
	return nil,b,fmt.Errorf("msgpack: string data truncated,totalen=%d,realLen=%d",stlen,haslen)
}
