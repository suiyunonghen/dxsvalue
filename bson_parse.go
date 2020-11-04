package dxsvalue

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/suiyunonghen/DxCommonLib"
	"math"
	"strconv"
	"time"
	"unsafe"
)

var(
	ErrInvalidBSON = errors.New("invalidate bson format")
	ErrInvalidArrayBsonIndex = errors.New("invalidate bson Array index")
	ErrBadBSONDocLen = errors.New("BSONDocument size invalidate")
)

/**
BSON文档（对象）由一个有序的元素列表构成。每个元素由一个字段名、一个类型和一个值组成。字段名为字符串。
所以BSON文档必然是Object
 */
func parseBsonDocument(b []byte,c *ValueCache,sharebinary bool,documentValue *DxValue)(docLen,parseIndex int,err error)  {
	if len(b) <= 4{
		return 0,0,ErrBadBSONDocLen
	}
	//文档长度
	docLen = int(binary.LittleEndian.Uint32(b[:4]))
	if docLen > len(b){
		return docLen,0,ErrBadBSONDocLen
	}
	parseIndex = 4
	var element *DxValue
	var keyName string
	//读取文档体，由一个个的元素组成
	//每个元素为一个类型+字段名+值组成
	for parseIndex < docLen{
		//读取类型
		btype := BsonType(b[parseIndex])
		parseIndex++
		//读取字段,cstring,ansic
		keyEnd := bytes.IndexByte(b[parseIndex:],0)
		if keyEnd < 0{
			return docLen,parseIndex,ErrInvalidBSON
		}
		keyEnd = parseIndex+keyEnd
		//UTF8
		if sharebinary{
			keyName = DxCommonLib.FastByte2String(b[parseIndex:keyEnd])
		}else{
			keyName = string(b[parseIndex:keyEnd])
		}
		parseIndex = keyEnd
		if err != nil{
			return docLen,parseIndex, ErrInvalidBSON
		}
		parseIndex++
		switch btype {
		case BSON_Array:
			element = documentValue.SetKeyCached(keyName,VT_Array,c)
			arraylen,parseIdx,err := parseBsonArray(b[parseIndex:],c,element,sharebinary)
			if err == nil{
				parseIndex += arraylen
			}else{
				return docLen,parseIndex + parseIdx,err
			}
		case BSON_EmbeddedDocument:
			element = documentValue.SetKeyCached(keyName,VT_Object,c)
			doclen,parseIdx,err := parseBsonDocument(b[parseIndex:],c,sharebinary,element)
			if err == nil{
				parseIndex += doclen
			}else{
				return docLen,parseIndex + parseIdx,err
			}
		default:
			element = documentValue.SetKeyCached(keyName,VT_NULL,c)
			_,parseidx,err := parseValue(btype,b[parseIndex:],element,sharebinary)
			if err == nil{
				parseIndex += parseidx
			}else{
				return docLen,parseIndex + parseidx,err
			}
		}

		if parseIndex == docLen - 1 && b[parseIndex] == 0{
			break
		}
	}
	return docLen,parseIndex,nil
}

//数组结构解析和文档结构解析差不多，将数组的索引号设定为docment的key
func parseBsonArray(b []byte,c *ValueCache,value *DxValue,shareBinary bool)(docLen,parseIndex int,err error)  {
	//文档长度
	docLen = int(binary.LittleEndian.Uint32(b[:4]))
	if docLen > len(b){
		return docLen,0,ErrBadBSONDocLen
	}
	parseIndex = 4
	keyName := ""
	//读取元素
	//每个元素为一个类型+字段名+值组成,字段名为数组的索引序号
	var element *DxValue
	for parseIndex < docLen{
		btype := BsonType(b[parseIndex])
		parseIndex++
		//读取字段,cstring,ansic
		keyEnd := bytes.IndexByte(b[parseIndex:],0)
		if keyEnd < 0{
			return docLen,parseIndex,ErrInvalidBSON
		}
		keyEnd = parseIndex+keyEnd
		if shareBinary{
			keyName = DxCommonLib.FastByte2String(b[parseIndex:keyEnd])
		}else {
			keyName = string(b[parseIndex:keyEnd])
		}
		parseIndex = keyEnd
		if err != nil{
			return docLen,parseIndex, ErrInvalidBSON
		}
		arridx := int(DxCommonLib.StrToIntDef(keyName,-1))
		if arridx == -1{
			return docLen,parseIndex,ErrInvalidArrayBsonIndex
		}
		parseIndex++
		switch btype {
		case BSON_Array:
			element = value.SetIndexCached(arridx,VT_Array,c)
			arraylen,parseIdx,err := parseBsonArray(b[parseIndex:],c,element,shareBinary)
			if err == nil{
				parseIndex += arraylen
			}else{
				return docLen,parseIndex + parseIdx,err
			}
		case BSON_EmbeddedDocument:
			element = value.SetIndexCached(arridx,VT_Object,c)
			doclen,parseIdx,err := parseBsonDocument(b[parseIndex:],c,shareBinary,element)
			if err == nil{
				parseIndex += doclen
			}else{
				return docLen,parseIndex + parseIdx,err
			}
		default:
			element = value.SetIndexCached(arridx,VT_NULL,c)
			_,parseidx,err := parseValue(btype,b[parseIndex:],element,shareBinary)
			if err == nil{
				parseIndex += parseidx
			}else{
				return docLen,parseIndex + parseidx,err
			}
		}
		if parseIndex == docLen - 1 && b[parseIndex] == 0{
			break
		}
	}
	return docLen,parseIndex,nil
}

func parseValue(btype BsonType, b []byte,value *DxValue,shareBinary bool)(valueLen,parseIndex int,err error)  {
	switch btype {
	case BSON_Double:
		u64 := binary.LittleEndian.Uint64(b[:8])
		value.SetDouble(*(*float64)(unsafe.Pointer(&u64)))
		return 8,8,nil
	case BSON_String:
		//UTF8
		valueLen = int(binary.LittleEndian.Uint32(b[:4]))
		parseIndex += 4
		start := parseIndex
		parseIndex += valueLen
		//由于会多写一个字符串结束符0，所以，实际长度应该减去1
		if shareBinary{
			value.SetString(DxCommonLib.FastByte2String(b[start:parseIndex - 1]))
		}else{
			value.SetString(string(b[start:parseIndex - 1]))
		}
	case BSON_Binary:
		value.DataType = VT_ExBinary
		value.ExtType = uint8(btype)
		//先读取长度
		valueLen = int(binary.LittleEndian.Uint32(b[:4]))
		parseIndex = 4
		//子类型
		b = b[parseIndex:]
		subType := b[0]
		if subType == 2{
			parseIndex ++
			b = b[parseIndex:]
			//再重新处理一遍
			valueLen = int(binary.LittleEndian.Uint32(b[:4]))
			parseIndex += 4
			b = b[parseIndex:]
			if valueLen > len(b){
				return valueLen,parseIndex,ErrInvalidBSON
			}
			value.DataType = VT_Binary
			value.ExtType = 0
		}else{
			//VT_ExBinary要将subType带上
			valueLen ++
		}
		parseIndex += valueLen
		if shareBinary{
			value.SetBinary(b[:valueLen],false)
		}else {
			value.SetBinary(b[:valueLen],true)
		}
	case BSON_Undefined:
	case BSON_ObjectID:
		//12字节
		value.DataType = VT_Binary
		value.ExtType = uint8(btype)
		value.SetBinary(b[:12],!shareBinary)
		return 12,12,nil
	case BSON_Boolean:
		//一个字节
		if b[0] == 1{
			value.DataType = VT_True
		}else{
			value.DataType = VT_False
		}
		return 1,1,nil
	case BSON_DateTime:
		//Unix 纪元(1970 年 1 月 1 日)以来的毫秒数
		vunix := time.Duration(binary.LittleEndian.Uint64(b[:8])) * time.Millisecond
		value.SetTime(time.Date(1970,1,1,0,0,0,0,time.Local).Add(vunix))
		return 8,8,nil
	case BSON_Null:
		//啥都不干
	case BSON_Regex:
	case BSON_DBPointer:
	case BSON_JavaScript:
	case BSON_Symbol:
	case BSON_CodeWithScope:
	case BSON_Int32:
		value.SetInt(int64(binary.LittleEndian.Uint32(b[:4])))
		return 4,4,nil
	case BSON_Timestamp:
		//64 位值 前4个字节是一个增量，后4个字节是一个时间戳。
		value.ExtType = uint8(btype)
		value.DataType = VT_Int
		copy(value.simpleV[:],b[:8])
		return 8,8,nil
	case BSON_Int64:
		value.SetInt(int64(binary.LittleEndian.Uint64(b[:8])))
		return 8,8,nil
	case BSON_Decimal128:
		value.DataType = VT_Binary
		value.ExtType = uint8(btype)
		value.SetBinary(b[:16],!shareBinary)
		return 16,16,nil
	case BSON_MinKey:
	case BSON_MaxKey:
	}
	return
}

func NewValueFromBson(b []byte,useCache bool,sharebinary bool)(*DxValue,error)  {
	var c *ValueCache
	if !useCache{
		c = nil
	}else{
		c = getCache()
	}
	root := c.getValue(VT_Object)
	_,_,err := parseBsonDocument(b,c,sharebinary,root)
	if err != nil{
		return nil, err
	}
	return root,err
}


func Value2Bson(v *DxValue, dst []byte)([]byte,error)  {
	if v.DataType != VT_Object{
		return nil,errors.New("只有Object可以转换")
	}
	return writeObjBsonValue(v,dst),nil
}

func writeObjBsonValue(v *DxValue, dst []byte)[]byte  {
	v.fobject.UnEscapestrs()
	//先留出文档长度
	lenindex := len(dst)
	dst = append(dst,0,0,0,0)
	for i := 0;i<len(v.fobject.strkvs);i++{
		//写入文档元素
		//类型
		dst = writeBsonElementType(v.fobject.strkvs[i].V,dst)
		//写入字段名,Ansic模式
		/*bt,err := DxCommonLib.GBKString(v.fobject.strkvs[i].K)
		if err != nil{
			return nil, err
		}*/
		//还是直接UTF8吧
		bt := DxCommonLib.FastString2Byte(v.fobject.strkvs[i].K)
		dst = append(dst,bt...)
		dst = append(dst,0)
		if v.fobject.strkvs[i].V == nil{
			continue
		}
		//写入value值
		switch v.fobject.strkvs[i].V.DataType {
		case VT_Object:
			dst = writeObjBsonValue(v.fobject.strkvs[i].V,dst)
		case VT_Array:
			dst = writeArrayBsonValue(v.fobject.strkvs[i].V,dst)
		default:
			dst = writeSimpleBsonValue(v.fobject.strkvs[i].V,dst)
		}
	}
	dst = append(dst,0) //增加一个结束符
	totallen := len(dst) - lenindex
	//写入实际的长度
	binary.LittleEndian.PutUint32(dst[lenindex:],uint32(totallen))
	return dst
}

func writeArrayBsonValue(v *DxValue, dst []byte)[]byte  {
	//先留出文档长度
	lenindex := len(dst)
	dst = append(dst,0,0,0,0)
	for i := 0;i<len(v.farr);i++{
		//写入文档元素
		//类型
		dst = writeBsonElementType(v.farr[i],dst)
		//写入字段名
		idxStr := strconv.Itoa(i)
		bt := DxCommonLib.FastString2Byte(idxStr)
		dst = append(dst,bt...)
		dst = append(dst,0)
		if v.farr[i] == nil{
			continue
		}
		//写入value值
		switch v.farr[i].DataType {
		case VT_Object:
			dst = writeObjBsonValue(v.farr[i],dst)
		case VT_Array:
			dst = writeArrayBsonValue(v.farr[i],dst)
		default:
			dst = writeSimpleBsonValue(v.farr[i],dst)
		}
	}
	dst = append(dst,0) //增加一个结束符
	totallen := len(dst) - lenindex
	//写入实际的长度
	binary.LittleEndian.PutUint32(dst[lenindex:],uint32(totallen))
	return dst
}

func putLittI32(i int32,dst []byte)[]byte  {
	l := len(dst)
	dst = append(dst,0,0,0,0)
	binary.LittleEndian.PutUint32(dst[l:],uint32(i))
	return dst
}

func putLittI64(i int64,dst []byte)[]byte  {
	l := len(dst)
	dst = append(dst,0,0,0,0,0,0,0,0)
	binary.LittleEndian.PutUint64(dst[l:],uint64(i))
	return dst
}

func writeSimpleBsonValue(v *DxValue,dst []byte)[]byte  {
	switch v.DataType {
	case VT_String,VT_RawString:
		//写入长度
		vstr := DxCommonLib.FastString2Byte(v.String())
		//加上一个字符结束空白
		dst = putLittI32(int32(len(vstr)+1),dst)
		//写入内容
		dst = append(dst,vstr...)
		//字符串结束
		dst = append(dst,0)
	case VT_Int:
		vint := v.Int()
		if v.ExtType == uint8(BSON_Timestamp){
			dst = append(dst,v.simpleV[:]...)
		}else{
			if vint < int64(math.MinInt32) || vint > int64(math.MaxInt32){
				dst = putLittI64(vint,dst)
			}else{
				dst = putLittI32(int32(vint),dst)
			}
		}
	case VT_Float,VT_Double:
		u64 := *(*int64)(unsafe.Pointer(&v.simpleV[0]))
		dst = putLittI64(u64,dst)
	case VT_DateTime:
		unixMillisecond := int64(v.GoTime().Sub(time.Date(1970,1,1,0,0,0,0,time.Local))/time.Millisecond)
		dst = putLittI64(unixMillisecond,dst)
	case VT_Binary:
		switch v.ExtType {
		case uint8(BSON_ObjectID):
			dst = append(dst,v.fbinary[:12]...)
		case uint8(BSON_Decimal128):
			dst = append(dst,v.fbinary[:16]...)
		default:
			//先写入长度
			dst = putLittI32(int32(len(v.fbinary)),dst)
			//写入subtype
			dst = append(dst,uint8(BinaryGeneric)) //普通类型
			//写入二进制
			dst = append(dst,v.fbinary...)
		}
	case VT_ExBinary:
		if v.ExtType == byte(BSON_Binary){
			//实际长度需要减去一个子类型占据的位
			dst = putLittI32(int32(len(v.fbinary)-1),dst)
		}else{
			//当普通二进制处理
			//先写入长度
			dst = putLittI32(int32(len(v.fbinary)),dst)
			//写入subtype
			dst = append(dst,uint8(BinaryGeneric)) //普通类型
		}
		//写入二进制
		dst = append(dst,v.fbinary...)
	case VT_True:
		dst = append(dst,1)
	case VT_False:
		dst = append(dst,0)
	}
	return dst
}

func writeBsonElementType(element *DxValue,dst []byte)[]byte  {
	if element == nil{
		dst = append(dst,byte(BSON_Null))
		return dst
	}
	switch element.DataType {
	case VT_Object:
		dst = append(dst,byte(BSON_EmbeddedDocument))
	case VT_Array:
		dst = append(dst,byte(BSON_Array))
	case VT_String,VT_RawString:
		dst = append(dst,byte(BSON_String))
	case VT_Int:
		v := element.Int()
		if element.ExtType == uint8(BSON_Timestamp){
			dst = append(dst,byte(BSON_Timestamp))
		}else{
			if v < int64(math.MinInt32) || v > int64(math.MaxInt32){
				dst = append(dst,byte(BSON_Int64))
			}else{
				dst = append(dst,byte(BSON_Int32))
			}
		}
	case VT_Float,VT_Double:
		dst = append(dst,byte(BSON_Double))
	case VT_DateTime:
		dst = append(dst,byte(BSON_DateTime))
	case VT_Binary,VT_ExBinary:
		if element.ExtType != 0{
			dst = append(dst,element.ExtType)
		}else{
			dst = append(dst,byte(BSON_Binary))
		}
		/*switch element.ExtType {
		case uint8(BSON_ObjectID):
			dst = append(dst,byte(BSON_ObjectID))
		case uint8(BSON_Decimal128):
			dst = append(dst,byte(BSON_Decimal128))
		default:
			dst = append(dst,byte(BSON_Binary))
		}*/
	case VT_NULL:
		dst = append(dst,byte(BSON_Null))
	case VT_True,VT_False:
		dst = append(dst,byte(BSON_Boolean))
	}
	return dst
}
