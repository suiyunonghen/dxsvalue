package dxsvalue

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewValueFromBson(t *testing.T) {
	bt,_ := ioutil.ReadFile("d:\\bson.dat")
	Value,err := NewValueFromBson(bt,true,true)
	if err != nil{
		fmt.Println(err)
		return
	}
	fmt.Println(len(bt))
	fmt.Println(Value.String())
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
