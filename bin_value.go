package dxsvalue
//msgpack的二进制类型

func (v *DxValue)ExtType()byte{
	if v.DataType == VT_ExBinary && len(v.fbinary) > 0{
		return v.fbinary[0]
	}
	return 0
}

func (v *DxValue)BinLen()int  {
	switch v.DataType {
	case VT_ExBinary:
		if len(v.fbinary)>0{
			return len(v.fbinary[1:])
		}
		return 0
	case VT_Binary:
		return len(v.fbinary)
	default:
		return 0
	}
}

func (v *DxValue)SetExtType(t byte)  {
	if v.DataType == VT_ExBinary{
		if len(v.fbinary) > 0{
			v.fbinary[0] = t
		}else{
			v.fbinary = append(v.fbinary,t)
		}
	}
}

