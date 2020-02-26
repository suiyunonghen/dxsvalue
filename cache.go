package dxsvalue

import (
	"sync"
)

var(
	cachePool	sync.Pool
)

type	cache struct {
	fisroot	bool
	Value	[]DxValue
}

func (c *cache)getValue(t ValueType)*DxValue  {
	if c == nil{
		return NewValue(t)
	}
	if cap(c.Value) > len(c.Value) {
		c.Value = c.Value[:len(c.Value)+1]
	} else {
		c.Value = append(c.Value, DxValue{})
	}
	result := &c.Value[len(c.Value)-1]
	result.Reset(t)
	if c.fisroot{
		c.fisroot = false
		result.ownercache = c
	}
	return result
}

func getCache()*cache  {
	var c *cache
	v := cachePool.Get()
	if v == nil{
		c = &cache{
			fisroot:	true,
			Value:    make([]DxValue,0,8),
		}
	}else{
		c = v.(*cache)
		c.fisroot = true
	}
	return c
}


//释放Value回收Cache
func FreeValue(v *DxValue)  {
	c := v.ownercache
	v.ownercache = nil
	if c!=nil{
		for i := 0;i<len(c.Value);i++{
			switch c.Value[i].DataType {
			case VT_Object:
				for j := 0;j<len(c.Value[i].fobject.strkvs);j++{
					c.Value[i].fobject.strkvs[j].V = nil
					c.Value[i].fobject.strkvs[j].K = ""
				}
				c.Value[i].fobject.strkvs = c.Value[i].fobject.strkvs[:0]
			case VT_Array:
				for j := 0;j<len(c.Value[i].farr);j++{
					c.Value[i].farr[j] = nil
				}
				c.Value[i].farr = c.Value[i].farr[:0]
			case VT_Binary,VT_ExBinary:
				c.Value[i].fbinary = nil
			case VT_String,VT_RawString:
				c.Value[i].fstrvalue = ""
			default:
				//DxCommonLib.ZeroByteSlice(v.simpleV[:])
			}
		}
		c.Value = c.Value[:0]
		cachePool.Put(c)
	}
}
