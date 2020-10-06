package resource

import (
	"encoding/json"
	"os/exec"
)

var aristaIfStatsOld map[string]map[string]map[string]float64

type PacketCounter struct {
	OutOctets float64 `json:"outOctets"`
	InOctets  float64 `json:"InOctets"`
}

func AristaIfStat() (map[string]map[string]map[string]float64, error) {
	stats := make(map[string]map[string]map[string]float64)
	aristaIfStats := make(map[string]map[string]map[string]float64)

	out, err := exec.Command("Cli", "-c", "show int counters | json").Output()
	if err != nil {
		return aristaIfStats, err
	}
	var j map[string]map[string]PacketCounter
	json.Unmarshal(out, &j)

	for iface, v := range j["interfaces"] {
		stats[iface] = make(map[string]map[string]float64)
		stats[iface]["rx"] = make(map[string]float64)
		stats[iface]["tx"] = make(map[string]float64)
		stats[iface]["rx"]["octets"] = v.InOctets
		stats[iface]["tx"]["octets"] = v.OutOctets
	}

	if aristaIfStatsOld == nil {
		aristaIfStatsOld = stats
	}

	for iface, stat := range aristaIfStatsOld {
		aristaIfStats[iface] = make(map[string]map[string]float64)
		aristaIfStats[iface]["rx"] = make(map[string]float64)
		aristaIfStats[iface]["tx"] = make(map[string]float64)
		item := "octets"
		// reset counter old stat set 0
		if stats[iface]["rx"][item] < stat["rx"][item] {
			stat["rx"][item] = 0
		}
		if stats[iface]["tx"][item] < stat["tx"][item] {
			stat["tx"][item] = 0
		}
		aristaIfStats[iface]["rx"][item] = stats[iface]["rx"][item] - stat["rx"][item]
		aristaIfStats[iface]["tx"][item] = stats[iface]["tx"][item] - stat["tx"][item]
	}
	aristaIfStatsOld = stats

	return aristaIfStats, nil
}
