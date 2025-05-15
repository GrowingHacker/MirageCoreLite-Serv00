package statusService

import (
	"mymodule/xraycoreHelper"
)

func CheckStatus(xray *xraycoreHelper.XrayService) bool {
	if xray.Instance == nil {
		return false
	}
	return xray.Instance.Running
}
