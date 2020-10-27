package dxsvalue

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"
)

var(
	vtest = []byte(`
-
 - bb: 
    gg: 32
    mm: asdf
 - aa
 -
  - cc
  - dd   
`)

	vtestB = []byte(`
- bb
- aa
-
  - cc
  - dd
`)
	vtestC = []byte(`
- bb
- aa
`)
	vmap = []byte(`
project:
  port: 8080
  name: &projectName epshealth-airobot-common

jwt:
  issuer: *projectName
  secret: c2VybmFtZSI6InRlc3QiLCJleHAiOjE2MDA5MjE0MjEsImlzcyI6ImVwc2hlYWx0aCIsIm5iZiI6MTYwMDkxNDIyMX0
  expires: 2h`)

	vmaparr = []byte(`
test:
  - 33
  - 44
  - 55
mm: sfasdf
istel: true`)
)

func TestYaml(t *testing.T)  {
	parser := newyamParser()
	defer freeyamlParser(parser)
	parser.fParsingValues = make([]yamlNode,0,10)
	parser.parseData = vmaparr
	parser.parse()
	fmt.Println(parser.root.String())

}

func TestDxValue_ForcePath(t *testing.T) {
	v := NewValueFrom(map[string]interface{}{
		"Name":"不得闲",
		"Age":33,
		"Weight":102.34,
	},true)
	fmt.Println(v.String())
	//v := NewValue(VT_Object)
	//v.SetKeyBool("a",true)
	v.ForcePath(VT_False,"a","b")
	v.ForcePath(VT_String,"a","b","c").SetString("Asdfaf")

	fmt.Println(v.String())

	v.Reset(VT_Array)
	v.SetIndex(0,VT_String).SetString("asdfadf")
	v.SetIndexString(1,"234")
	v.SetIndex(2,VT_Object)
	v.SetIndexBool(3,true)
	fmt.Println(v.String())
	v.SetIndexString(3,"ASdf")
	fmt.Println(v.String())
	v1,_ := NewValueFromJson([]byte(`["asdfadf","234",{},"ASdf"]`),false,false)
	fmt.Println("结果 ",v1.String())
	v2 := v1.Clone(false)
	fmt.Println(v2.String())
	return
}

func TestParseJsonValue(t *testing.T) {
	str := `{"Result":0,"Name":"不得闲","Age":36,"Weight":167.3,"arr":[ {"gg":23},23 ]}`
	v,err := NewValueFromJson([]byte(str),true,false)
	if err != nil{
		fmt.Println("发生错误：",err)
	}
	fmt.Println(v.StringByPath("","Result"))
	fmt.Println(v.StringByPath("","arr","0","gg"))
	v.SetKeyString("Parent","测试Parent")
	chld := v.SetKey("Childs",VT_Object)
	chld.SetKeyString("Name","TestName")
	chld.SetKeyInt("Age",3)
	fmt.Println(v.StringByPath("","Childs","Name"))
	fmt.Println(v.StringByPath("","Parent"))
	fmt.Println(string(Value2Json(v,JSE_OnlyAnsiChar,false,nil)))

	bt := make([]byte,0,1024)
	fmt.Println(string(formatValue(v,JSE_OnlyAnsiChar,false,bt,0)))
	FreeValue(v)

}

func TestNewValue(t *testing.T) {
	v := NewValue(VT_Object)
	v.SetKeyString("Name","不得闲")
	v.SetKeyInt("Age",36)
	v.SetKeyFloat("Weight",23.5)
	arrv := v.SetKey("Children",VT_Array)
	child := arrv.SetIndex(0,VT_Object)
	child.SetKeyString("Name","Child1")
	child.SetKeyString("Sex","boy")
	child.SetKeyInt("Age",3)

	child = arrv.SetIndex(1,VT_Object)
	child.SetKeyString("Name","Chil\"d2\"")
	child.SetKeyString("Sex","girl")
	child.SetKeyInt("Age",3)
	fmt.Println(string(Value2Json(v,JSE_OnlyAnsiChar,false,nil)))
	fmt.Println(string(Value2Json(v,JSE_AllEscape,false,nil)))
	fmt.Println(string(Value2Json(v,JSE_NoEscape,false,nil)))
}

type People struct {
	Name		string
	Age			int
	Weight		float32
	IsMen		bool
}

func TestDxValue_SetKeyvalue(t *testing.T) {
	value := NewObject(true)
	p := People{Name:`{"DxSoft":"gg"}`,Age:20,Weight:23.24,IsMen:true}
	value.SetKeyvalue("one",p,value.ValueCache())
	value.SetKeyvalue("two",&p,value.ValueCache())
	fmt.Println(string(Value2Json(value,JSE_OnlyAnsiChar,false,nil)))
	value.Clear()
	value.SetKeyString("Name","DxSoft")
}

func TestNewValueFromMsgPack(t *testing.T) {
	b,_ := ioutil.ReadFile(`DataProxy.config.msgPack`)
	v,err := NewValueFromMsgPack(b,false,false)
	if err != nil{
		fmt.Println(err)
		return
	}
	fmt.Println(v.String())

	b = formatValue(v,JSE_OnlyAnsiChar,false,nil,0)
	fmt.Println(string(b))
	b = Value2MsgPack(v,nil)
	v1,err := NewValueFromMsgPack(b,false,false)
	fmt.Print(v1.String())
}

func TestDxValue_MergeWith(t *testing.T) {
	v1 := NewValue(VT_Object)
	rootv := v1.SetKey("People",VT_Object)
	rootv.SetKeyString("Name","DxSoft")
	rootv.SetKeyString("Childs","ChildName1")

	v2 := NewValue(VT_Object)
	rootv = v2.SetKey("People",VT_Object)
	rootv.SetKeyString("Sex","Men")
	rootv.SetKeyInt("Age",20)
	rootv.SetKeyString("Childs","ChildName2")
	fmt.Println(v1.String())
	fmt.Println(v2.String())

	v1.MergeWith(v2, func(keypath string,oldv *DxValue, newv *DxValue) MergeOp {
		fmt.Println(keypath)
		if keypath == "People/Childs"{
			return MO_Replace
		}
		return MO_Normal
	})
	fmt.Println(v1.String())

	//数组合并
	varr1 := NewValue(VT_Array)
	varr1.SetIndexString(0,"Dxsoft")
	varr1.SetIndexString(1,"People")
	varr1.SetIndexInt(2,20)

	varr2 := NewValue(VT_Array)
	varr2.SetIndexString(0,"Dxsoft")
	varr2.SetIndexString(1,"People")
	varr2.SetIndexInt(2,40)

	varr1.MergeWith(varr2,nil)
	fmt.Println(varr1.String())
}

func TestDxValue_LoadFromJson(t *testing.T) {
	strBody := `
{
"LastUpdate":"1899-12-30 00:00:00"
}
`

	value := NewObject(true)
	err := value.LoadFromJson([]byte(strBody),true)
	if err != nil{
		fmt.Println(err)
	}
	value.SetKeyTime("LastUpdate",time.Time{})
	dst := Value2MsgPack(value,nil)
	value.Clear()
	if value.LoadFromMsgPack(dst,false)!=nil{
		fmt.Println("sadfa")
	}else{
		fmt.Println(value.String())
	}

}

func TestDxValue_AsDateTime(t *testing.T) {
	rec := NewValue(VT_Object)
	rec.SetKeyString("url","www.baidu.com")
	rec.SetKeyString("Host","234")
	rec.SetKeyInt("state",4)
	rec.SetKeyInt("FileSize",20)
	rec.SetKeyTime("LastUpdate",time.Now())
	rec.SetKeyString("Content-Type","zip")
	rec.SetKeyString("ContentEncoding","br")

	pkg := NewValue(VT_Object)
	pkg.SetKeyString("Name", "Method")
	pkg.SetKeyInt("Type", 1) //1是方法
	pkg.SetKeyValue("Params", rec)

	bt := Value2MsgPack(pkg,nil)
	recv := NewValue(VT_Object)
	err := recv.LoadFromMsgPack(bt,false)
	if err != nil{
		fmt.Println(err)
	}else{
		fmt.Println(recv.String())
	}

}