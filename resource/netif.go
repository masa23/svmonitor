package resource

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

var netifStatsOld map[string]map[string]map[string]int64
var netifItems = []string{"bytes", "packets", "errs", "drop", "fifo", "compressed", "multicast"}

func NetIfCount() (map[string]map[string]map[string]int64, error) {
	stats := make(map[string]map[string]map[string]int64)
	netifStats := make(map[string]map[string]map[string]int64)
	fd, err := os.Open("/proc/net/dev")
	if err != nil {
		return netifStats, err
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)

	for scanner.Scan() {
		l := scanner.Text()
		s := strings.Fields(l)

		if len(s) < 17 {
			continue
		}
		ifname := strings.TrimRight(s[0], ":")

		stat := make(map[string]map[string]int64)
		stat["rx"] = make(map[string]int64)
		stat["tx"] = make(map[string]int64)
		for i, item := range netifItems {
			stat["rx"][item], _ = strconv.ParseInt(s[i+1], 10, 64)
			if err != nil {
				return netifStats, err
			}
			stat["tx"][item], err = strconv.ParseInt(s[i+9], 10, 64)
			if err != nil {
				return netifStats, err
			}
		}
		stats[ifname] = stat
	}

	if err := scanner.Err(); err != nil {
		return netifStats, err
	}

	if netifStatsOld == nil {
		netifStatsOld = stats
	}

	for iface, stat := range netifStatsOld {
		netifStats[iface] = make(map[string]map[string]int64)
		netifStats[iface]["rx"] = make(map[string]int64)
		netifStats[iface]["tx"] = make(map[string]int64)
		for _, item := range netifItems {
			// reset counter old stat set 0
			if stats[iface]["rx"][item] < stat["rx"][item] {
				stat["rx"][item] = 0
			}
			if stats[iface]["tx"][item] < stat["tx"][item] {
				stat["tx"][item] = 0
			}
			netifStats[iface]["rx"][item] = stats[iface]["rx"][item] - stat["rx"][item]
			netifStats[iface]["tx"][item] = stats[iface]["tx"][item] - stat["tx"][item]
		}
	}
	netifStatsOld = stats

	return netifStats, nil
}
