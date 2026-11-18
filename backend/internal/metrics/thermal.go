package metrics

import (
	"os"
	"strconv"
	"strings"
)

// ReadCPUTempC lee la temperatura del SoC en grados Celsius desde sysfs (sin vcgencmd,
// sin privilegios especiales -- a diferencia de throttle.go, que si necesita hostctl).
func ReadCPUTempC(sysPath string) (float64, error) {
	data, err := os.ReadFile(sysPath + "/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0, err
	}
	milliC, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	if err != nil {
		return 0, err
	}
	return milliC / 1000, nil
}
