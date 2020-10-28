package dxsvalue

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewValueFromBson(t *testing.T) {
	bt,_ := ioutil.ReadFile("d:\\1.bson")
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
	}
	fmt.Println(len(dst))
	NewValue,err := NewValueFromBson(dst,true,true)
	ioutil.WriteFile("d:\\2.bson",dst, os.ModePerm)
	fmt.Println(NewValue.String())


}
