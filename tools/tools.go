package tools

import "github.com/vmware/govmomi/vim25/types"

//StatusConvert 状态转换器
func StatusConvert(status types.ManagedEntityStatus) float64 {
	switch status {
	case types.ManagedEntityStatusGreen:
		return float64(1)
	case types.ManagedEntityStatusGray:
		return float64(2)
	case types.ManagedEntityStatusYellow:
		return float64(3)
	case types.ManagedEntityStatusRed:
		return float64(4)
	default:
		return float64(5)
	}
}
