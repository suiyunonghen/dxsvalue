package dxsvalue

import "sync"

var(
	cachePool	sync.Pool
)

type	cache struct {
	fisroot	bool
	Value	[]DxValue
	cacheBuffer	[]byte
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
		c.Value = c.Value[:0]
		cachePool.Put(c)
	}
}
