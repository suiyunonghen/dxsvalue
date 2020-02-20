package dxsvalue
//msgpack的二进制类型

func (v *DxValue)ExtType()byte{
	if v.DataType == VT_ExBinary && len(v.fbinary) > 0{
		return v.fbinary[0]
	}
	return 0
}


