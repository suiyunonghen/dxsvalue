package dxsvalue

import (
	"fmt"
	"testing"
)

func TestDxValue_AsInt(t *testing.T) {
	var v DxValue
	v.SetInt(20)
	fmt.Println(v.AsString())
	fmt.Println(v.AsInt())
	v.SetString("不得闲")
	fmt.Println(v.AsString())
	fmt.Println(v.AsInt())

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
	chld := v.NewValue("Childs",VT_Object)
	chld.SetKeyString("Name","TestName")
	chld.SetKeyInt("Age",3)
	fmt.Println(v.StringByPath("","Childs","Name"))
	fmt.Println(v.StringByPath("","Parent"))
	fmt.Println(string(Value2Json(v,nil)))
	FreeValue(v)

}
