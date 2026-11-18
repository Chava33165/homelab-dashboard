package metrics

import (
	"os"
	"strconv"
	"strings"
)

func ReadUptime(procPath string) (float64, error) {
	data, err := os.ReadFile(procPath + "/uptime")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, nil
	}
	return strconv.ParseFloat(fields[0], 64)
}
