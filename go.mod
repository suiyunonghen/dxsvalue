module github.com/suiyunonghen/dxsvalue

go 1.13

require (
	github.com/json-iterator/go v1.1.7
	github.com/suiyunonghen/DxCommonLib v0.2.0
	github.com/suiyunonghen/DxValue v1.1.0
)

replace (
	github.com/suiyunonghen/DxCommonLib => /../DxCommonLib
	github.com/suiyunonghen/DxValue => /../DxValue
)
