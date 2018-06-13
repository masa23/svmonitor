package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	graphite "github.com/marpaia/graphite-golang"
	"github.com/masa23/svmonitor"
	"github.com/masa23/svmonitor/resource"
)

var conf svmonitor.Config

func main() {
	var configFile string
	var err error

	flag.StringVar(&configFile, "config", "./config.yaml", "config file path")
	flag.Parse()
	conf, err = svmonitor.ConfigLoad(configFile)
	if err != nil {
		panic(err)
	}
	sendChan := make(chan []graphite.Metric, conf.Graphite.SendBuffer)
	go graphiteSendMetrics(sendChan)
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		now := time.Now()
		target := now.Truncate(conf.ReportInterval)
		d := target.Add(conf.ReportInterval).Sub(now)

		timer.Reset(d)
		t := <-timer.C

		metrics, err := createGraphiteMetrics(t)
		if err != nil {
			panic(err)
		}
		sendChan <- metrics
		t = t.Truncate(conf.ReportInterval).Add(-conf.ReportInterval)
	}
}

func graphiteSendMetrics(sendChan chan []graphite.Metric) {
	g, err := graphite.NewGraphite(conf.Graphite.Host, conf.Graphite.Port)
	if err != nil {
		panic(err)
	}
	defer g.Disconnect()
	for {
		metrics := <-sendChan
		for {
			err = g.SendMetrics(metrics)
			if err != nil {
				for {
					g.Disconnect()
					g, err = graphite.NewGraphite(conf.Graphite.Host, conf.Graphite.Port)
					if err != nil {
						time.Sleep(time.Second)
						continue
					}
					break
				}
			}
			break
		}
	}
}

func createGraphiteMetrics(t time.Time) ([]graphite.Metric, error) {
	var metrics []graphite.Metric

	if conf.Metric.CPU.Enable {
		cpus, err := resource.CPUStat()
		if err != nil {
			return metrics, err
		}

		if conf.Metric.CPU.AllCores {
			for cpu, stat := range cpus {
				for item, value := range stat {
					metrics = append(metrics, graphite.Metric{
						Name:      fmt.Sprintf("%s.%s.%s.%s", conf.Graphite.Prefix, "cpu", cpu, item),
						Value:     strconv.FormatInt(value, 10),
						Timestamp: t.Unix(),
					})
				}
			}
		} else {
			for item, value := range cpus["cpu"] {
				metrics = append(metrics, graphite.Metric{
					Name:      fmt.Sprintf("%s.%s.%s.%s", conf.Graphite.Prefix, "cpu", "cpu", item),
					Value:     strconv.FormatInt(value, 10),
					Timestamp: t.Unix(),
				})
			}
		}
	}

	if conf.Metric.Network.Enable {
		netif, err := resource.NetIfStat()
		if err != nil {
			return metrics, err
		}
		for iface, ifstat := range netif {
			if !inArrayString(iface, conf.Metric.Network.Interfaces) {
				continue
			}
			for rxtx, stat := range ifstat {
				for item, value := range stat {
					metrics = append(metrics, graphite.Metric{
						Name:      fmt.Sprintf("%s.%s.%s.%s.%s", conf.Graphite.Prefix, "netif", iface, rxtx, item),
						Value:     strconv.FormatInt(value, 10),
						Timestamp: t.Unix(),
					})
				}
			}
		}
	}
	return metrics, nil
}

func inArrayString(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
