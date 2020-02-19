module github.com/suiyunonghen/dxfastvalue

go 1.13

require (
	github.com/json-iterator/go v1.1.7
	github.com/suiyunonghen/DxCommonLib v0.1.8
	github.com/suiyunonghen/DxValue v1.0.4
)

replace (
	github.com/suiyunonghen/DxCommonLib => /../DxCommonLib
	github.com/suiyunonghen/DxValue => /../DxValue
)
