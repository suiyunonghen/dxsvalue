package dxsvalue

import (
	"bytes"
	"fmt"
	"github.com/suiyunonghen/DxCommonLib"
	"math"
	"sync"
)

func spaceTrim(r rune) bool {
	return r == ' ' || r == '\n'
}

type yamlNode struct {
	isMapArrayNode	bool
	spaceCount		int			//空格个数
	v				*DxValue
}

type yamlParser struct {
	ifFirstObjEle		bool				//是否是对象的第一个元素
	lastSpaceCount		int
	root				*DxValue
	fparentCache		*ValueCache
	parseData			[]byte
	fParsingValues		[]yamlNode			//正在解析中的Value
	useageValues		[]strkv	//引用值信息
}


var(
	yamlparserPool	sync.Pool
)

func newyamParser()*yamlParser  {
	v := yamlparserPool.Get()
	if v != nil{
		return v.(*yamlParser)
	}
	return &yamlParser{
		fParsingValues:  make([]yamlNode,0,32),
		useageValues: make([]strkv,0,32),
	}
}

func freeyamlParser(parser *yamlParser)  {
	parser.reset(nil)
	yamlparserPool.Put(parser)
}

func (parser *yamlParser)reset(data []byte)  {
	parser.parseData = data
	for i := 0;i<len(parser.fParsingValues);i++{
		parser.fParsingValues[i].v = nil
	}
	parser.fParsingValues = parser.fParsingValues[:0]
	for i := 0;i<len(parser.useageValues);i++{
		parser.useageValues[i].V = nil
		parser.useageValues[i].K = ""
	}
	parser.useageValues = parser.useageValues[:0]
}

func (parser *yamlParser)setCachedUseage(cacheName string,cacheValue *DxValue)  {
	for i := 0;i<len(parser.useageValues);i++{
		if parser.useageValues[i].K == cacheName{
			parser.useageValues[i].V = cacheValue
			return
		}
	}
	parser.useageValues = append(parser.useageValues,strkv{cacheName,cacheValue})
}

func (parser *yamlParser)getUseage(cacheName string)*DxValue  {
	for i := 0;i<len(parser.useageValues);i++{
		if parser.useageValues[i].K == cacheName{
			return parser.useageValues[i].V
		}
	}
	return nil
}

func (parser *yamlParser)popLast(popCount int)  {
	l := len(parser.fParsingValues)
	if l > 0{
		if popCount > l{
			popCount = l
		}
		for i,j := l - 1,popCount;i>=0 && j > 0;i,j = i-1,j-1{
			parser.fParsingValues[i].v = nil
		}
		parser.fParsingValues = parser.fParsingValues[:l-popCount]
	}
}

func (parser *yamlParser)parseArray(dataLine []byte,spaceCount int)error  {
	var currentValue *DxValue
	isvalue := false
	if len(dataLine) > 0{
		if dataLine[0] != ' '{
			//判定一下当前的空格数量
			if spaceCount <= parser.lastSpaceCount{
				return fmt.Errorf("无效的Yaml格式：%s",string(dataLine))
			}
			//只有大于的时候，才可能认为是上个内容的字符串
		}
		dataLine = bytes.TrimFunc(dataLine,spaceTrim)
		isvalue = len(dataLine) > 0
	}
	lastIndex := len(parser.fParsingValues) - 1
	if lastIndex < 0{
		currentValue = NewArray(true)
		parser.root = currentValue
		parser.fparentCache = currentValue.ValueCache()
		parser.fParsingValues = append(parser.fParsingValues,yamlNode{false,-1,currentValue})
	}else{
		currentValue = parser.fParsingValues[lastIndex].v
	}
	isstrvalue := false
	if currentValue.DataType == VT_Object{
		currentValue.Reset(VT_Array)
	}
	if currentValue.DataType == VT_Array{
		if isvalue{
			currentValue = currentValue.SetIndexCached(math.MaxInt64,VT_String,parser.fparentCache)
		}else{
			currentValue = currentValue.SetIndexCached(math.MaxInt64,VT_Array,parser.fparentCache)
		}
		parser.fParsingValues = append(parser.fParsingValues,yamlNode{false,spaceCount,currentValue})
	}else{
		isstrvalue = true
	}
	if isvalue{
		if isstrvalue{
			currentValue.SetString(currentValue.String() + " "+string(dataLine))
		}else{
			k,v,hasKeySplit := parser.splitKv(dataLine)
			if !hasKeySplit{ //不存在键值分隔符
				parser.parseStringValue(currentValue,k)

			}else{
				if len(k) == 0{
					return fmt.Errorf("无效的Yaml格式：%s",string(dataLine))
				}
				//map结构增加
				currentValue.Reset(VT_Object)
				//要加上数组标记"- "的长度2
				parser.fParsingValues[len(parser.fParsingValues) - 1].isMapArrayNode = true
				parser.ifFirstObjEle = true
				//parser.fParsingValues[len(parser.fParsingValues) - 1].spaceCount = parser.fParsingValues[len(parser.fParsingValues) - 1].spaceCount + 2
				if len(v) == 0{
					currentValue = currentValue.SetKeyCached(string(k),VT_NULL,parser.fparentCache)
				}else{
					currentValue = currentValue.SetKeyCached(string(k),VT_String,parser.fparentCache)
					parser.parseStringValue(currentValue,v)
				}
				parser.fParsingValues = append(parser.fParsingValues,yamlNode{true,spaceCount,currentValue})

			}
		}
	}
	return nil
}

func (parser *yamlParser)parseLine(lineData []byte,spaceCount int)error  {
	dataLine := lineData[spaceCount:]
	if dataLine[0] == '#'{
		//注释，不做处理
		return nil
	}
	lastIndex := bytes.IndexByte(dataLine,'#')
	if lastIndex > -1{ //去掉注释
		dataLine = dataLine[:lastIndex]
	}
	if len(dataLine) == 0{
		//注释不处理
		return nil
	}

	isArray := dataLine[0] == '-'
	lastIndex = len(parser.fParsingValues) - 1
	lastSpaceCount := parser.lastSpaceCount
	if parser.ifFirstObjEle && lastIndex >= 0{
		if parser.fParsingValues[lastIndex].isMapArrayNode{ // - name: //这种数组和object集合的
			lastSpaceCount += 2
		}
	}

	if  lastSpaceCount > spaceCount{
		//大于的时候，需要将上一个值写入的弹出,然后还需要将值的parent弹出
		parser.popLast(2)
		//然后查找和当前的spaceCount相同的组
		l := len(parser.fParsingValues)
		popcount := l
		for i := l - 1;i>0;i--{
			if parser.fParsingValues[i].spaceCount < spaceCount{
				break
			}else{
				parser.fParsingValues[i].v = nil
				popcount--
			}
		}
		parser.fParsingValues = parser.fParsingValues[:popcount]
	}else if lastSpaceCount == spaceCount{
		parser.popLast(1)
	}
	var err error
	if isArray{
		err = parser.parseArray(dataLine[1:],spaceCount)
	}else{
		err = parser.parseObject(dataLine,spaceCount)
	}

	parser.lastSpaceCount = spaceCount
	return err
}

func (parser *yamlParser)parseObject(dataLine []byte,spaceCount int)error  {
	var currentValue *DxValue
	//非数组

	lastIndex := len(parser.fParsingValues) - 1
	lastSpaceCount := parser.lastSpaceCount
	if lastIndex < 0{
		currentValue = NewObject(true)
		parser.root = currentValue
		parser.fparentCache = currentValue.ValueCache()
		parser.fParsingValues = append(parser.fParsingValues,yamlNode{false,-1,currentValue})
	}else{
		if parser.ifFirstObjEle && parser.fParsingValues[lastIndex].isMapArrayNode{ // - name: //这种数组和object集合的
			lastSpaceCount += 2
		}
		currentValue = parser.fParsingValues[lastIndex].v
	}
	if parser.ifFirstObjEle{
		parser.ifFirstObjEle = false
	}

	key,value,hasKeySplit := parser.splitKv(dataLine)
	if string(key) == "name"{
		fmt.Println("asdf")
	}


	if lastIndex < 0{
		if len(key) == 0 || !hasKeySplit{
			return fmt.Errorf("无效的Yaml格式：%s",string(dataLine))
		}
	}
	if  lastSpaceCount == spaceCount{
		//和上一步是同一级，判定一下是否具备KV结构
		if len(key) == 0 || !hasKeySplit{
			return fmt.Errorf("无效的Yaml格式：%s",string(dataLine))
		}
		currentValue = currentValue.SetKeyCached(string(key),VT_String,parser.fparentCache)
		parser.fParsingValues = append(parser.fParsingValues,yamlNode{false,spaceCount,currentValue})
		//currentValue.SetString(string(value))
		parser.parseStringValue(currentValue,value)
	}else if spaceCount > lastSpaceCount{
		//子集
		if len(key) == 0{
			//字符串
			if currentValue.DataType == VT_Object || currentValue.DataType == VT_Array{
				return fmt.Errorf("无效的Yaml格式：%s",string(dataLine))
			}
			oldstr := currentValue.String()
			currentValue.SetString(oldstr + " :"+string(value))
			return nil
		}
		if currentValue.DataType != VT_Object{
			currentValue.Reset(VT_Object)
		}
		vkey := string(key)
		if len(value) == 0{
			currentValue = currentValue.SetKeyCached(vkey,VT_Object,parser.fparentCache)
		}else{
			//判定一下是否是引用界定值
			if value[0] != '&'{
				if value[0] == '*'{ //引用其他值信息

				}else{
					currentValue = currentValue.SetKeyCached(vkey,VT_String,parser.fparentCache)
					parser.parseStringValue(currentValue,value)
				}
			}else{
				//引用的
				currentValue = currentValue.SetKeyCached(vkey,VT_Object,parser.fparentCache)
				parser.setCachedUseage(vkey,currentValue)
			}
		}
		parser.fParsingValues = append(parser.fParsingValues,yamlNode{false,spaceCount,currentValue})
	}else{
		//完成闭包处理，回到上一个对象处理结构
		if len(key) == 0 || !hasKeySplit || currentValue.DataType != VT_Object{
			return fmt.Errorf("无效的Yaml格式：%s",string(dataLine))
		}
		if len(value) == 0{
			currentValue = currentValue.SetKeyCached(string(key),VT_Object,parser.fparentCache)
		}else{
			currentValue = currentValue.SetKeyCached(string(key),VT_String,parser.fparentCache)
			//currentValue.SetString(string(value))
			parser.parseStringValue(currentValue,value)
		}
		parser.fParsingValues = append(parser.fParsingValues,yamlNode{false,spaceCount,currentValue})
	}

	return nil
}

func (parser *yamlParser)splitKv(dataline []byte)(k,v []byte,bool2 bool)  {
	dataline = bytes.TrimFunc(dataline,spaceTrim)
	splitindex := bytes.IndexByte(dataline,':')
	if splitindex < 0{
		return dataline,nil,false
	}
	return bytes.TrimFunc(dataline[:splitindex],spaceTrim),bytes.TrimFunc(dataline[splitindex+1:],spaceTrim),true
}

func (parser *yamlParser)parseStringValue(target *DxValue,vstr []byte)  {
	vlen := len(vstr)
	if vlen == 4 && bytes.Compare(vstr,truebyte) == 0{
		target.SetBool(true)
		return
	}
	if vlen == 5 && bytes.Compare(vstr,falebyte) == 0{
		target.SetBool(false)
		return
	}
	if vfloat,err := DxCommonLib.ParseFloat(DxCommonLib.FastByte2String(vstr));err == nil{
		target.SetDouble(vfloat)
		return
	}
	//判定一下是否有使用引用参数的
	if vstr[0] == '{' && vstr[vlen - 1] == '}'{
		//解析Json Map格式的串结构
	}else if vstr[0] == '[' && vstr[vlen - 1] == ']'{
		//解析Json Array格式的串结构
	}else{

		target.SetString(string(vstr))
	}
}

func (parser *yamlParser)parse()error  {
	datalen := len(parser.parseData)
	if datalen == 0{
		return nil
	}
	istart := 0
	curLineEnd := 0
	parser.lastSpaceCount = -1
	for {
		spaceCount := 0
		isStart := true
		curLineEnd = datalen
		for i := istart;i<datalen;i++{
			if parser.parseData[i] == '\n'{
				curLineEnd = i
				break
			}
			if isStart{
				if parser.parseData[i] == ' '{
					spaceCount++
				}else{
					isStart = false
				}
			}
		}
		if curLineEnd - istart < 2{
			istart = curLineEnd + 1
			if istart == datalen{
				break
			}
			continue
		}
		err := parser.parseLine(parser.parseData[istart: curLineEnd],spaceCount)
		if err != nil{
			return err
		}

		istart = curLineEnd + 1
		if istart >= datalen{
			break
		}
	}
	return nil
}
