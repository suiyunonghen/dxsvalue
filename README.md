# DxValue
万能值变量，其目的是将各种编码类型（目前已经支持了JSON,MsgPack,Yaml,BSON）都可以使用本库来进行操作处理，
本来以前写了个DxValue，之后参考了一些其他的相关解码编码库，感觉不是很理想，所以本版本是重构之后的优化版本，
之前将不同的类型值通过指针进行区分，对于某些使用操作不太统一方便，所以本次重写，全部使用一个结构代替，同时
内部支持cache缓存创建，也便于回收，支持说明：
- JSON完全支持
- MsgPack完全支持
- Yaml简要支持，目前支持嵌入，引用归并等，暂时不支持行内类Json格式解析
- BSON基本上完全支持，对于一些特有编码类型的如minkey等未做支持

## 用法

> go get github.com/suiyunonghen/dxsvalue

1、NewValue
```go
    v := dxsvalue.NewValue(VT_Object)
    v.SetKeyString("Name","不得闲")
    v.SetKeyInt("Age",36)
    v.SetKeyFloat("Weight",23.5)
    //添加一个数组
    arrv := v.SetKey("Children",VT_Array)
    //数组中增加一个对象
    child := arrv.SetIndex(0,VT_Object)
    child.SetKeyString("Name","Child1")
    child.SetKeyString("Sex","boy")
    child.SetKeyInt("Age",3)
    //数组中增加一个对象
    child = arrv.SetIndex(1,VT_Object)
    child.SetKeyString("Name","Child2")
    child.SetKeyString("Sex","girl")
    child.SetKeyInt("Age",3)
    fmt.Println(string(Value2Json(v,nil)))    
```
以上的代码最终的输出结果为
```
{
    "Name": "不得闲",
    "Age": 36,
    "Weight": 23.5,
    "Children": [
        {
            "Name": "Child1",
            "Sex": "boy",
            "Age": 3
        },
        {
            "Name": "Child2",
            "Sex": "girl",
            "Age": 3
        }
    ]
}
```

2、NewValueFromJson
>使用本函数主要是从JSON中构建一个DxValue对象进行操作，有两个参数，参数2指定是否使用cache，如果使用
>了cache，之后可以使用FreeValue对Value进行回收，此时使用的Value都是cache中的对象结构，用法：
```go
    str := `{"Result":0,"Name":"不得闲","Age":36,"Weight":167.3,"arr":[ {"gg":23},23 ]}`
	v,err := dxsvalue.NewValueFromJson([]byte(str),true)
	if err != nil{
		fmt.Println("发生错误：",err)
	}
    defer dxsvalue.FreeValue(v)
```
除以上之外，还有NewValueFromBson,NewValueFromMsgPack,NewValueFromYaml，使用方法差不多


3、获取值
> DxValue可以使用String,Int,Bool,Float,DateTime,GoTime等来获取简单类型Value的相对
>应的数据值，相应的使用SetXXX函数来设置相应的数据值,AsString,AsInt等可以用来获取相应的
>key对应的值
>
>
>获取某个节点的值，可以使用ValueByName()来查找到某个节点，然后相应的使用AsXXX等函数来获取，同时也可以
>使用ValueByPath获取一个路径下面的节点，比如要获取 a/b/c ValueByPath('a','b','c'),相应的也有
>StringByPath,BoolByPath等
>
>
4、设定值
> 使用SetXX类的函数，SetKeyXXX类型的主要用来设定K-V结构的数据设定主要针对Object类型，SetIndex类型主
>要针对数组类的数据设定