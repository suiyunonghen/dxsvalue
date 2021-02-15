package dxsvalue

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestNewValueFromBson(t *testing.T) {
	bt,_ := ioutil.ReadFile("d:\\bson.dat")
	Value,err := NewValueFromBson(bt,true,true)
	if err != nil{
		fmt.Println(err)
		Value = NewValue(VT_Object)
		Value.SetKeyString("Name","测试人")
		Value.SetKeyInt("Age",32)
		Value.SetKeyDouble("score",234.523)
		Value.SetKeyTime("now",time.Now())
		childValue := Value.SetKeyCached("Childs",VT_Object,Value.ValueCache())
		childValue.SetKeyCached("Name",VT_String,Value.ValueCache()).SetString("默默子")
		childValue.SetKeyCached("HashCode",VT_Int,Value.ValueCache()).SetInt(323)
	}else{
		fmt.Println(len(bt))
		fmt.Println(Value.String())
	}
	dst,err := Value2Bson(Value,nil)
	if err != nil{
		fmt.Println(err)
		return
	}


	fmt.Println(len(dst))
	NewValue,err := NewValueFromBson(dst,true,true)
	ioutil.WriteFile("d:\\2.bson",dst, os.ModePerm)
	fmt.Println(NewValue.String())


	NewValue.Clear()
	NewValue.SetKeyCached("Binary",VT_Binary,NewValue.ValueCache()).SetBinaryFromFile("d:\\map.html")
	NewValue.SetKeyCached("姓名",VT_String,NewValue.ValueCache()).SetString("不得闲")
	dst,err = Value2Bson(NewValue,nil)
	if err != nil{
		fmt.Println(err)
		return
	}

	ioutil.WriteFile("d:\\3.bson",dst, os.ModePerm)


	Value,err = NewValueFromBson(dst,true,true)
	if err != nil{
		fmt.Println(err)
		return
	}
	fmt.Println(string(Value.ValueByName("姓名").String()))

}
