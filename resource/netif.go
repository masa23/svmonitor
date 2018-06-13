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
			stat["rx"][item], err = strconv.ParseInt(s[i+1], 10, 64)
			if err != nil {
				return netifStats, err
			}
			stat["tx"][item], err = strconv.ParseInt(s[i+8], 10, 64)
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

/*
root@ubuntu:~# cat /proc/net/dev
Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
  eth0: 6141744392 15601694    0    0    0     0          0     88574 68061616282 17354415    0    0    0     0       0          0
vethUKX23M: 34579323204 9818802    0 4530    0     0          0         0 955547501 7837238    0  160    0     0       0          0
vethQX6BM2:  161980    3526    0    0    0     0          0         0 41295216  210813    0    0    0     0       0          0
vethAGJPPT: 25417704  295622    0    0    0     0          0         0 699614568  446035    0    0    0     0       0          0
   br0: 1745031441 1094302    0    0    0     0          0         0 194154473  799664    0    0    0     0       0          0
   br1: 1903517    6547    0    0    0     0          0         0    26666     377    0    0    0     0       0          0
virbr0-nic:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0
  eth1: 41299669  210964    0    0    0     0          0      6194   186564    3876    0    0    0     0       0          0
    lo:  327446    4537    0    0    0     0          0         0   327446    4537    0    0    0     0       0          0
veth66EN97: 6768934   79498    0    0    0     0          0         0 200511623  195443    0    0    0     0       0          0
virbr0:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0
*/
