module github.com/suiyunonghen/dxfastvalue

go 1.13

require (
	github.com/json-iterator/go v1.1.7
	github.com/suiyunonghen/DxCommonLib v0.1.3
	github.com/suiyunonghen/DxValue v0.0.0
	github.com/valyala/fastjson v1.4.5
)

replace (
	github.com/suiyunonghen/DxCommonLib => /../DxCommonLib
	github.com/suiyunonghen/DxValue => /../DxValue
	github.com/valyala/fastjson => ../../valyala/fastjson
)
