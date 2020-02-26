package dxsvalue

import (
	"fmt"
	"math"
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
		result.fstrvalue = string(b[:stlen])
		return result,b[stlen:],nil
	}
	return nil,b,fmt.Errorf("msgpack: string data truncated,totalen=%d,realLen=%d",stlen,haslen)
}

func getExtcode(exbinLen int)MsgPackCode  {
	switch exbinLen {
	case 1:
		return CodeFixExt1
	case 2:
		return CodeFixExt2
	case 4:
		return CodeFixExt4
	case 8:
		return CodeFixExt8
	case 16:
		return CodeFixExt16
	}
	if exbinLen < 256 {
		return CodeExt8
	}
	if exbinLen < 65536 {
		return CodeExt16
	}
	return CodeExt32
}

func writeExtCode(exlen int,dst []byte)[]byte  {
	switch exlen {
	case 1:
		dst = append(dst,byte(CodeFixExt1))
	case 2:
		dst = append(dst,byte(CodeFixExt2))
	case 4:
		dst = append(dst,byte(CodeFixExt4))
	case 8:
		dst = append(dst,byte(CodeFixExt8))
	case 16:
		dst = append(dst,byte(CodeFixExt16))
	}
	if exlen < 256 {
		dst = append(dst,byte(CodeExt8),byte(exlen))
	}else if exlen < 65536 {
		dst = append(dst,byte(CodeExt16),byte(exlen >> 8),byte(exlen))
	}else{
		dst = append(dst,byte(CodeExt32),byte(exlen >> 24),byte(exlen >> 16),byte(exlen >> 8),byte(exlen))
	}
	return dst
}

func writeBinCode(binlen int,dst []byte)[]byte  {
	if binlen < 256 {
		dst = append(dst,byte(CodeBin8),byte(binlen))
	}else if binlen < 65536 {
		dst = append(dst,byte(CodeBin16),byte(binlen >> 8),byte(binlen))
	}else{
		dst = append(dst,byte(CodeBin32),byte(binlen >> 24),byte(binlen >> 16),byte(binlen >> 8),byte(binlen))
	}
	return dst
}

func writeStrCode(strlen int,dst []byte)[]byte  {
	switch {
	case strlen < 32:
		dst = append(dst,byte(CodeFixedStrLow) | byte(strlen))
	case strlen < 256:
		dst = append(dst,byte(CodeStr8),byte(strlen))
	case strlen < 65536:
		dst = append(dst,byte(CodeStr16),byte(strlen >> 8),byte(strlen))
	default:
		dst = append(dst,byte(CodeStr32),byte(strlen >> 24),byte(strlen >> 16),byte(strlen >> 8),byte(strlen))
	}
	return dst
}

func writeMapCode(maplen int,dst []byte)[]byte  {
	switch {
	case maplen < 16:
		dst = append(dst,byte(CodeFixedMapLow) | byte(maplen))
	case maplen < 65536:
		dst = append(dst,byte(CodeMap16),byte(maplen >> 8),byte(maplen))
	default:
		dst = append(dst,byte(CodeMap32),byte(maplen >> 24),byte(maplen >> 16),byte(maplen >> 8),byte(maplen))
	}
	return dst
}

func writeArrayCode(arrlen int,dst []byte)[]byte  {
	switch {
	case arrlen < 16:
		dst = append(dst,byte(CodeFixedArrayLow) | byte(arrlen))
	case arrlen < 65536:
		dst = append(dst,byte(CodeArray16),byte(arrlen >> 8),byte(arrlen))
	default:
		dst = append(dst,byte(CodeArray32),byte(arrlen >> 24),byte(arrlen >> 16),byte(arrlen >> 8),byte(arrlen))
	}
	return dst
}

func writeInt(n int64,dst []byte)[]byte  {
	if n >= 0{
		if n <= math.MaxInt8 {
			dst = append(dst,byte(n))
		}else if n <= math.MaxUint8 {
			dst = append(dst,byte(CodeUint8),byte(n))
		}else if n <= math.MaxUint16 {
			dst = append(dst,byte(CodeUint16),byte(n >> 8),byte(n))
		}else if n <= math.MaxUint32 {
			dst = append(dst,byte(CodeUint32),byte(n >> 24),byte(n >> 16),byte(n >> 8),byte(n))
		}else{
			dst = append(dst,byte(CodeUint64),byte(n >> 56),byte(n >> 48),byte(n >> 40),byte(n >> 32),byte(n >> 24),byte(n >> 16),byte(n >> 8),byte(n))
		}
		return dst
	}
	var low byte
	low = byte(NegFixedNumLow)
	if n >= int64(int8(low)) {
		dst = append(dst,byte(n))
	}else if n >= math.MinInt8 {
		dst = append(dst,byte(CodeInt8),byte(n))
	}else if n >= math.MinInt16{
		dst = append(dst,byte(CodeInt16),byte(n >> 8),byte(n))
	}else if n >= math.MinInt32 {
		dst = append(dst,byte(CodeInt32),byte(n >> 24),byte(n >> 16),byte(n >> 8),byte(n))
	}else{
		//64位
		dst = append(dst,byte(CodeInt64),byte(n >> 56),byte(n >> 48),byte(n >> 40),byte(n >> 32),byte(n >> 24),byte(n >> 16),byte(n >> 8),byte(n))
	}
	return dst
}