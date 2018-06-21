package resource

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

var cpuStatsOld map[string]map[string]int64
var cpuItems = []string{"user", "nice", "system", "idle", "hwirq", "swirq", "steal"}

// CPU is cpu use resource parcents
func CPUStat() (map[string]map[string]int64, error) {
	cpuStats := make(map[string]map[string]int64)
	cpuUse := make(map[string]map[string]int64)
	fd, err := os.Open("/proc/stat")
	if err != nil {
		return cpuUse, err
	}
	defer fd.Close()
	scanner := bufio.NewScanner(fd)

	for scanner.Scan() {
		stat := make(map[string]int64)
		l := scanner.Text()
		if !strings.HasPrefix(l, "cpu") {
			break
		}
		s := strings.Fields(l)
		if len(s) < 8 {
			continue
		}
		for i, item := range cpuItems {
			n, _ := strconv.ParseInt(s[i+1], 10, 64)
			stat[item] = n
		}
		cpuStats[s[0]] = stat
	}

	if cpuStatsOld == nil {
		cpuStatsOld = cpuStats
	}

	for cpu := range cpuStatsOld {
		var all int64
		var all2 int64
		for _, item := range cpuItems {
			if cpuUse[cpu] == nil {
				cpuUse[cpu] = make(map[string]int64)
			}
			cpuUse[cpu][item] = cpuStats[cpu][item] - cpuStatsOld[cpu][item]
			all += cpuUse[cpu][item]
		}
		for _, item := range cpuItems {
			if cpuUse[cpu][item] > 0 {
				cpuUse[cpu][item] = (cpuUse[cpu][item] * 100) / all
				all2 += cpuUse[cpu][item]
			}
		}
		// 端数はidleに入れ込んでしまう
		cpuUse[cpu]["idle"] += (100 - all2)
	}
	return cpuUse, nil
}
