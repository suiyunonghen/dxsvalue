package dxsvalue

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestDxValue_ForcePath(t *testing.T) {
	v := NewValue(VT_Object)
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
	v1,_ := NewValueFromJson([]byte(`["asdfadf","234",{},"ASdf"]`),false)
	fmt.Println("结果 ",v1.String())
	v2 := v1.Clone(false)
	fmt.Println(v2.String())
	return
}

func TestParseJsonValue(t *testing.T) {
	str := `{"Result":0,"Name":"不得闲","Age":36,"Weight":167.3,"arr":[ {"gg":23},23 ]}`
	v,err := NewValueFromJson([]byte(str),true)
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
	fmt.Println(string(Value2Json(v,true,nil)))
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
	child.SetKeyString("Name","Child2")
	child.SetKeyString("Sex","girl")
	child.SetKeyInt("Age",3)
	fmt.Println(string(Value2Json(v,true,nil)))
}

func TestNewValueFromMsgPack(t *testing.T) {
	b,_ := ioutil.ReadFile(`DataProxy.config.msgPack`)
	v,err := NewValueFromMsgPack(b,false)
	if err != nil{
		fmt.Println(err)
		return
	}
	fmt.Println(v.String())
	b = Value2MsgPack(v,nil)
	v1,err := NewValueFromMsgPack(b,false)
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