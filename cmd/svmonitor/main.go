package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/k0kubun/pp"
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
	g, err := graphite.NewGraphite(conf.Graphite.Host, conf.Graphite.Port)
	if err != nil {
		panic(err)
	}
	defer g.Disconnect()
	go func() {
		for {
			metrics := <-sendChan
			for {
				pp.Println(metrics)
				err = g.SendMetrics(metrics)
				if err != nil {
					g.Disconnect()
					pp.Println(err)
					g, err = graphite.NewGraphite(conf.Graphite.Host, conf.Graphite.Port)
					time.Sleep(time.Millisecond * 100)
				}
				break
			}
		}
	}()

	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		now := time.Now()
		target := now.Truncate(conf.ReportInterval)
		d := target.Add(conf.ReportInterval).Sub(now)

		timer.Reset(d)
		t := <-timer.C

		metrics, err := sendGraphiteMetrics(t)
		if err != nil {
			panic(err)
		}
		sendChan <- metrics
		t = t.Truncate(conf.ReportInterval).Add(-conf.ReportInterval)
	}
}

func sendGraphiteMetrics(t time.Time) ([]graphite.Metric, error) {
	var metrics []graphite.Metric
	cpus, err := resource.CPU()
	if err != nil {
		return metrics, err
	}

	for cpu, stat := range cpus {
		for item, value := range stat {
			metrics = append(metrics, graphite.Metric{
				Name:      fmt.Sprintf("%s.%s.%s.%s", conf.Graphite.Prefix, "cpu", cpu, item),
				Value:     strconv.FormatInt(value, 10),
				Timestamp: t.Unix(),
			})
		}
	}
	return metrics, nil
}
