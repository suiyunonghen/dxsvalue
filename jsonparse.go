package dxsvalue

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/suiyunonghen/DxCommonLib"
	"io/ioutil"
	"os"
	"strconv"
)

type JsonErrType	uint8

const(
	JET_NoObjBack JsonErrType = iota		//缺少}
	JET_NoArrBack
	JET_NoKeyStart
	JET_NoKeyEnd
	JET_NoStrStart
	JET_NoStrEnd
	JET_NoKVSplit
	JET_NoValueSplit
	JET_Invalidate
)

func (tp JsonErrType)String()string  {
	switch tp {
	case JET_NoObjBack:
		return "缺少对象的结束返回符}"
	case JET_NoArrBack:
		return "缺少数组结构的结束返回符]"
	case JET_NoKeyStart:
		return "缺少键值开始的\""
	case JET_NoKeyEnd:
		return "缺少键值结束的\""
	case JET_NoStrStart:
		return "缺少字符串开始\""
	case JET_NoStrEnd:
		return "缺少字符串结束\""
	case JET_NoKVSplit:
		return "缺少键值分隔符："
	case JET_NoValueSplit:
		return "缺少值分隔符,"
	}
	return "未知的错误"
}

type ErrorParseJson struct {
	Type				JsonErrType
	InvalidIndex		int
	parseB				[]byte
}

func (err *ErrorParseJson)Error()string  {
	return fmt.Sprintf("解析在数据%s在位置%d处发生错误：%s",DxCommonLib.FastByte2String(err.parseB),err.InvalidIndex,err.Type.String())
}

//解析从fastjson中的代码修改
//如果成功，tail中返回的是剩下的字节内容，否则发生错误的话，tail中返回的是当前正在解析的数据
func parseJsonObj(b []byte,c *cache)(result *DxValue,tail []byte,err error)  {
	oldb := b
	b,skiplen := skipWB(b)
	if len(b) == 0{
		return nil,oldb,&ErrorParseJson{
			Type:         JET_NoObjBack,
			InvalidIndex: skiplen,
			parseB:oldb,
		}
	}
	result = c.getValue(VT_Object)
	if b[0] == '}'{
		tail = b[1:]
		return
	}
	for{
		kv := result.fobject.getKv()
		oldb := b
		b,skiplen = skipWB(b)
		if len(b) == 0 || b[0] != '"' {
			return nil, oldb, &ErrorParseJson{
				Type:         JET_NoKeyStart,
				InvalidIndex: skiplen,
				parseB:oldb,
			}
		}
		kv.K, b, err = parseJsonKey(b[1:],c!=nil)
		if err != nil{
			result = nil
			return
		}
		oldb = b
		b,skiplen = skipWB(b)
		if len(b) == 0 || b[0] != ':' {
			result = nil
			err = &ErrorParseJson{
				Type:         JET_NoKVSplit,
				InvalidIndex: skiplen,
				parseB:       oldb,
			}
			return
		}
		oldb = b
		b,skiplen = skipWB(b[1:])
		//解析Value
		kv.V, b, err = parseJsonValue(b,c)
		if err != nil {
			result = nil
			return
		}
		oldb = b
		b,skiplen = skipWB(b)
		if len(b) == 0 {
			result = nil
			err = &ErrorParseJson{
				Type:         JET_NoObjBack,
				InvalidIndex: skiplen,
				parseB:       oldb,
			}
			return
		}
		if b[0] == ',' {
			b = b[1:]
			continue
		}
		if b[0] == '}' {
			return result, b[1:], nil
		}
		err = &ErrorParseJson{
			Type:         JET_NoValueSplit,
			InvalidIndex: skiplen,
			parseB:       oldb,
		}
		return nil, oldb, err
	}
}

func parseJsonKey(b []byte,useCache bool) (key string, tail []byte, err error) {
	l := len(b)
	for i := 0; i < l; i++ {
		if b[i] == '"' {
			if useCache{
				return string(b[:i]), b[i+1:], nil
			}
			return string(b[:i]), b[i+1:], nil
		}
		if b[i] == '\\' { //有转义的
			key,tail,err = parseJsonString(b,useCache)
			if jpe,ok := err.(*ErrorParseJson);ok{
				if jpe.Type == JET_NoStrEnd{
					jpe.Type = JET_NoKeyEnd
				}
			}
		}
	}
	return "", nil, &ErrorParseJson{
		Type:         JET_NoKeyStart,
		InvalidIndex: l,
	}
}

func parseJsonString(b []byte,useCache bool) (value string, tail []byte, err error) {
	n := bytes.IndexByte(b, '"')
	if n < 0 {
		return "", b, &ErrorParseJson{
			Type:         JET_NoStrEnd,
			InvalidIndex: 0,
		}
	}
	if n == 0 || b[n-1] != '\\' {//不是转义的"
		if useCache{
			return string(b[:n]), b[n+1:], nil
		}
		return string(b[:n]), b[n+1:], nil
	}

	ss := b
	for {
		i := n - 1
		for i > 0 && b[i-1] == '\\' {
			i--
		}
		if uint(n-i)%2 == 0 {
			if useCache{
				return string(ss[:len(ss)-len(b)+n]), b[n+1:], nil
			}
			return string(ss[:len(ss)-len(b)+n]), b[n+1:], nil
		}
		b = b[n+1:]

		n = bytes.IndexByte(b, '"')
		if n < 0 {
			return string(ss), ss, &ErrorParseJson{
				Type:         JET_NoStrEnd,
				InvalidIndex: 0,
			}
		}
		if n == 0 || b[n-1] != '\\' {
			if useCache{
				return string(ss[:len(ss)-len(b)+n]), b[n+1:], nil
			}
			return string(ss[:len(ss)-len(b)+n]), b[n+1:], nil
		}
	}
}

func parseJsonArray(b []byte,c *cache)(result *DxValue,tail []byte,err error)  {
	oldb := b
	b,skiplen := skipWB(b)
	if len(b) == 0 {
		err = &ErrorParseJson{
			Type:         JET_NoArrBack,
			InvalidIndex: skiplen,
			parseB:       oldb,
		}
		return nil, oldb, err
	}
	result = c.getValue(VT_Array)
	if b[0] == ']' {
		return result, b[1:], nil
	}
	var v *DxValue
	for {
		oldb = b
		b,skiplen = skipWB(b)
		v, b, err = parseJsonValue(b,c)
		if err != nil {
			return nil,b,err
		}
		result.farr = append(result.farr, v)

		oldb = b
		b,skiplen = skipWB(b)
		if len(b) == 0 {
			return nil, oldb, &ErrorParseJson{
				Type:         JET_NoArrBack,
				InvalidIndex: skiplen,
				parseB:       oldb,
			}
		}
		if b[0] == ',' {
			b = b[1:]
			continue
		}
		if b[0] == ']' {
			b = b[1:]
			return result, b, nil
		}
		err = &ErrorParseJson{
			Type:         JET_NoValueSplit,
			InvalidIndex: skiplen,
			parseB:       oldb,
		}
		return nil, oldb, err
	}
}

var(
	truebyte = []byte("true")
	falebyte = []byte("false")
	nullbyte = []byte("null")
	nanbyte = []byte("nan")
	infbyte = []byte("inf")
	valueTrue = &DxValue{DataType: VT_True}
	valueFalse = &DxValue{DataType: VT_False}
	valueNAN = &DxValue{DataType: VT_NAN}
	valueINF = &DxValue{DataType: VT_INF}
	valueNull  = &DxValue{DataType: VT_NULL}
)


func NewValueFromJson(b []byte,useCache bool)(*DxValue,error)  {
	var c *cache
	if !useCache{
		c = nil
	}else{
		c = getCache()
	}
	b,skiplen := skipWB(b)
	v, tail, err := parseJsonValue(b,c)
	if err != nil {
		return nil, err
	}
	b,skiplen = skipWB(tail)
	if len(b) > 0 {
		err = &ErrorParseJson{
			Type:         JET_Invalidate,
			InvalidIndex: skiplen,
			parseB:       tail,
		}
		return nil, err
	}
	return v,nil
}

func parseJsonValue(b []byte,c *cache)(result *DxValue,tail []byte,err error)  {
	if len(b) == 0{
		return nil,nil,&ErrorParseJson{
			Type:         0,
			InvalidIndex: 0,
			parseB:       b,
		}
	}
	if b[0] == '{'{
		return parseJsonObj(b[1:],c)
	}
	if b[0] == '[' {
		return parseJsonArray(b[1:],c)
	}
	if b[0] == '"' {
		ss, tail, err := parseJsonString(b[1:],c!=nil)
		if err != nil {
			return nil, tail, err
		}
		//先判断一下是否是Json的日期格式
		var result *DxValue
		dt := DxCommonLib.ParserJsonTime(ss)
		if dt < 0{
			result = c.getValue(VT_RawString)
			result.fstrvalue = ss
		}else{
			result = c.getValue(VT_DateTime)
			result.SetDouble(float64(dt))
		}
		return result, tail, nil
	}
	if b[0] == 't' {
		if len(b) < 4 || bytes.Compare(b[:4],truebyte) != 0 {
			return nil, b, errors.New("无效的Json格式")
		}
		return valueTrue, b[4:], nil
	}
	if b[0] == 'f' {
		if len(b) < 5 || bytes.Compare(b[:5],falebyte) != 0 {
			return nil, b,errors.New("无效的Json格式")
		}
		return valueFalse, b[5:], nil
	}
	if b[0] == 'n' {
		blen := len(b)
		if blen < 4 || bytes.Compare(b[:4],nullbyte) != 0 {
			if blen >= 3 && bytes.Compare(b[:3],nanbyte) != 0 {
				return valueNAN, b[3:], nil
			}
			return nil, b,errors.New("无效的Json格式")
		}
		return valueNull, b[4:], nil
	}
	return parseJsonNum(b,c)
}

func parseJsonNum(b []byte,c *cache) (num *DxValue, tail []byte,err error) {
	isfloat := false
	for i := 0; i < len(b); i++ {
		ch := b[i]
		isfloat = ch == '.'
		if (ch >= '0' && ch <= '9') || isfloat || ch == '-' || ch == 'e' || ch == 'E' || ch == '+' {
			continue
		}
		if i == 0 || i == 1 && (b[0] == '-' || b[0] == '+') {
			if len(b[i:]) >= 3 {
				xs := b[i : i+3]
				if bytes.EqualFold(xs, infbyte) {
					return valueINF, b[i+3:], nil
				}
				if bytes.EqualFold(xs, nanbyte){
					return valueNAN, b[i+3:], nil
				}
			}
			return nil, b, errors.New("无效的Json格式")
		}
		ns := b[:i]
		b = b[i:]
		if isfloat{
			v := DxCommonLib.StrToFloatDef(DxCommonLib.FastByte2String(ns),0)
			num = c.getValue(VT_Float)
			num.SetFloat(float32(v))
		}else{
			v := DxCommonLib.StrToIntDef(DxCommonLib.FastByte2String(ns),0)
			num = c.getValue(VT_Int)
			num.SetInt(v)
		}
		return num, b, nil
	}
	return nil, nil, nil
}

func skipWB(b []byte) (r []byte,skiplen int) {
	if len(b) == 0 || b[0] > 0x20 {
		return b,0
	}
	if len(b) == 0 || b[0] != 0x20 && b[0] != 0x0A && b[0] != 0x09 && b[0] != 0x0D {
		return b,0
	}
	for i := 1; i < len(b); i++ {
		if b[i] != 0x20 && b[i] != 0x0A && b[i] != 0x09 && b[i] != 0x0D {
			return b[i:],i-1
		}
	}
	return nil,0
}


func Value2Json(v *DxValue,escapestr bool, dst []byte)[]byte  {
	if dst == nil{
		dst = make([]byte,0,256)
	}
	switch v.DataType {
	case VT_Object:
		dst = append(dst,'{')
		for i := 0;i<len(v.fobject.strkvs);i++{
			if i != 0{
				dst = append(dst,`,"`...)
			}else{
				dst = append(dst,'"')
			}
			if escapestr{
				if v.fobject.keysUnescaped{
					dst = DxCommonLib.EscapeJsonbyte(v.fobject.strkvs[i].K,dst)
				}else{
					dst = append(dst,v.fobject.strkvs[i].K...)
				}
			}else{
				if !v.fobject.keysUnescaped{
					v.fobject.UnEscapestrs()
				}
				dst = append(dst,v.fobject.strkvs[i].K...)
			}
			dst = append(dst,`":`...)
			dst = Value2Json(v.fobject.strkvs[i].V,escapestr,dst)
		}
		dst = append(dst,'}')
	case VT_String:
		dst = append(dst,'"')
		if escapestr {
			dst = DxCommonLib.EscapeJsonbyte(v.fstrvalue,dst)
		}else{
			dst = append(dst,DxCommonLib.FastString2Byte(v.fstrvalue)...)
		}
		dst = append(dst,'"')
	case VT_Array:
		dst = append(dst, '[')
		for i := 0;i<len(v.farr);i++{
			if i != 0{
				dst = append(dst, ',')
			}
			if v.farr[i] != nil{
				dst = Value2Json(v.farr[i],escapestr,dst)
			}else{
				dst = append(dst,'n','u','l','l')
			}
		}
		dst = append(dst, ']')
	case VT_Float,VT_Double:
		dst = strconv.AppendFloat(dst,v.Double(),'f',-1,64)
	case VT_RawString:
		dst = append(dst,'"')
		dst = append(dst,DxCommonLib.FastString2Byte(v.fstrvalue)...)
		dst = append(dst,'"')
	case VT_True:
		dst = append(dst,"true"...)
	case VT_False:
		dst = append(dst,"false"...)
	case VT_Int:
		dst = strconv.AppendInt(dst,v.Int(),10)
	case VT_DateTime:
		dst = append(dst,"/Date("...)
		unixs := int64((DxCommonLib.TDateTime)(v.Float()).ToTime().Unix()*1000)
		dst = strconv.AppendInt(dst,unixs,10)
		dst = append(dst,")/"...)
	case VT_NULL:
		dst = append(dst,'n','u','l','l')
	}
	return dst
}

func NewValueFromJsonFile(fileName string,usecache bool)(*DxValue,error)  {
	databytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil,err
	}
	if len(databytes) > 2 && databytes[0] == 0xEF && databytes[1] == 0xBB && databytes[2] == 0xBF{//BOM
		databytes = databytes[3:]
	}
	return NewValueFromJson(databytes,usecache)
}

func Value2File(v *DxValue, fileName string,BOMFile bool)error{
	if file,err := os.OpenFile(fileName,os.O_CREATE | os.O_TRUNC,0644);err == nil{
		defer file.Close()
		if BOMFile{
			file.Write([]byte{0xEF,0xBB,0xBF})
		}
		dst := Value2Json(v,false,nil)
		_,err := file.Write(dst)
		return err
	}else{
		return err
	}
}