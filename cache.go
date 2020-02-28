package dxsvalue

import (
	"sync"
)

var(
	ValueCachePool	sync.Pool
)

type	ValueCache struct {
	fisroot	bool
	Value	[]DxValue
}

func (c *ValueCache)getValue(t ValueType)*DxValue  {
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

func getCache()*ValueCache  {
	var c *ValueCache
	v := ValueCachePool.Get()
	if v == nil{
		c = &ValueCache{
			fisroot:	true,
			Value:    make([]DxValue,0,8),
		}
	}else{
		c = v.(*ValueCache)
		c.fisroot = true
	}
	return c
}

func (c *ValueCache)Reset(toRoot bool)  {
	if !toRoot{
		c.Value = c.Value[:1]
	}else{
		c.Value = c.Value[:0]
	}
}

//释放Value回收ValueCache
func FreeValue(v *DxValue)  {
	c := v.ownercache
	v.ownercache = nil
	if c!=nil{
		c.Reset(true)
		ValueCachePool.Put(c)
	}
}
