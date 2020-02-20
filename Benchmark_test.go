package dxsvalue

import (
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
	dvalue "github.com/suiyunonghen/DxValue"
	"io/ioutil"
	"testing"
)

func BenchmarkJsonParse(b *testing.B) {
	buf, err := ioutil.ReadFile("DataProxy.config.json")
	if err != nil {
		return
	}
	b.Run("std", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(buf)))
		b.RunParallel(func(pb *testing.PB) {
			mp := make(map[string]interface{}, 0)
			for pb.Next() {
				json.Unmarshal(buf, &mp)
				json.Marshal(mp)
			}
		})
	})
	b.Run("DxRecord", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(buf)))
		b.RunParallel(func(pb *testing.PB) {
			rc := dvalue.NewRecord()
			for pb.Next() {
				rc.JsonParserFromByte(buf,false,false)
			}
		})
	})
	b.Run("DxValue", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(buf)))
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				v,_ := NewValueFromJson(buf,true)
				Value2Json(v,false,nil)
			}
		})
	})
	b.Run("Jsoniter", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(buf)))
		b.RunParallel(func(pb *testing.PB) {
			mp := make(map[string]interface{},100)
			for pb.Next() {
				jsoniter.Unmarshal(buf,&mp)
				jsoniter.Marshal(mp)
			}
		})
	})
}

