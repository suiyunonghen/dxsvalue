package dxsvalue

import (
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"testing"
)

type MarshalTest struct {
	Name		string
	Age			int
	Desp		string
	Info		struct{
		Name1		string
		Name2		string
		Age2		int
	}
}

func BenchmarkJsonMa(b *testing.B)  {
	b.Run("std", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			var t MarshalTest
			for pb.Next() {
				json.Marshal(&t)
			}
		})
	})

	b.Run("DxValue", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			var t MarshalTest
			for pb.Next() {
				Marshal(&t)
			}
		})
	})
}

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
	/*b.Run("DxRecord", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(buf)))
		b.RunParallel(func(pb *testing.PB) {
			rc := dvalue.NewRecord()
			for pb.Next() {
				rc.JsonParserFromByte(buf,false,false)
				rc.Bytes(false)
			}
		})
	})*/
	b.Run("DxValue", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(buf)))
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				v,_ := NewValueFromJson(buf,false,false)
				Value2Json(v,JSE_OnlyAnsiChar,false,nil)
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

func BenchmarkMsgPackParse(b *testing.B) {
	buf, err := ioutil.ReadFile("DataProxy.config.msgPack")
	if err != nil {
		return
	}
	/*b.Run("DxRecord", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(buf)))
		b.RunParallel(func(pb *testing.PB) {
			rc := dvalue.NewRecord()
			r := bytes.NewReader(buf)
			w := bytes.NewBuffer(make([]byte,0,1024))
			for pb.Next() {
				r.Reset(buf)
				rc.LoadMsgPackReader(r)
				//rc.SaveMsgPackFile()
				dvalue.NewEncoder(w).EncodeRecord(rc)
			}
		})
	})*/
	b.Run("DxValue", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(buf)))
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				v,_ := NewValueFromMsgPack(buf,false,false)
				Value2MsgPack(v,nil)
			}
		})
	})
}