package stopservice

import (
	"mymodule/xraycoreHelper"
)

func Stop(xray *xraycoreHelper.XrayService) {
	xray.Stop()
}
