package dxsvalue

import (
	"io/ioutil"
	"os"
)

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

func (v *DxValue)Binary()[]byte  {
	switch v.DataType {
	case VT_Binary:
		return v.fbinary
	case VT_ExBinary:
		if len(v.fbinary) > 0{
			return v.fbinary[1:]
		}
		return nil
	default:
		return nil
	}
}

func (v *DxValue)SetBinary(b []byte,valueClone bool)  {
	switch v.DataType {
	case VT_Binary:
		if valueClone{
			v.fbinary = append(v.fbinary[:0],b...)
		}else{
			v.fbinary = b
		}
	case VT_ExBinary:
		if len(v.fbinary) > 0{
			v.fbinary = append(v.fbinary[:0],v.fbinary[0])
			v.fbinary = append(v.fbinary,b...)
		}else{
			v.Reset(VT_Binary)
			if valueClone{
				v.fbinary = append(v.fbinary[:0],b...)
			}else{
				v.fbinary = b
			}
		}
	}
}

func (v *DxValue)SaveBinaryToFile(fileName string)error  {
	if  v.DataType == VT_Binary || v.DataType == VT_ExBinary{
		if file,err := os.OpenFile(fileName,os.O_CREATE | os.O_TRUNC,0644);err == nil{
			defer file.Close()
			if v.DataType == VT_Binary{
				_,err := file.Write(v.fbinary)
				return err
			}
			if v.DataType == VT_ExBinary && len(v.fbinary)>0{
				_,err := file.Write(v.fbinary[1:])
				return err
			}
		}else{
			return err
		}
	}
	return nil
}

func (v *DxValue)SetBinaryFromFile(fileName string)error  {
	if v.DataType == VT_Binary || v.DataType == VT_ExBinary{
		b,err := ioutil.ReadFile(fileName)
		if err != nil{
			return err
		}
		v.SetBinary(b,false)
	}
	return nil
}
