package main

import (
    "fmt"
    "os"
    "log"

    "github.com/shirou/gopsutil/cpu"
    "github.com/xxwassyxx/sensu-plugins-go/lib/check"
    "github.com/xxwassyxx/sensu-plugins-go/lib/influxdb"
)

var conf  string

func main() {
	var (
		warn  int
		crit  int
	)

	c := check.New("windows_check_cpu")
	c.Option.IntVarP(&warn, "warn", "w", 80, "WARN")
	c.Option.IntVarP(&crit, "crit", "c", 90, "CRIT")
	c.Option.StringVarP(&conf, "conf", "C", "influx.json", "CONF")
	c.Init()

	usage, err := getStats()
	if err != nil {
		c.Error(err)
	}

	switch {
	case usage >= float64(crit):
		c.Critical(fmt.Sprintf("%.0f%%", usage))
	case usage >= float64(warn):
		c.Warning(fmt.Sprintf("%.0f%%", usage))
	default:
		c.Ok(fmt.Sprintf("%.0f%%", usage))
	}
}

func getStats() (float64, error) {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}

  cpuTotal, err := cpu.Times(false)
  if err != nil {
		log.Println("error %v", err)
	}

  perCPU, err := cpu.Times(true)
  if err != nil {
		log.Println("error %v", err)
	}
	
	times := append(perCPU, cpuTotal...)

	for _, cts := range times {
		tags := map[string]string{
			"cpu": cts.CPU,
		}
		tags["host"] = name

		totalDelta := cts.User + cts.System + cts.Nice + cts.Iowait + cts.Irq + cts.Softirq + cts.Steal + cts.Idle

		fields := map[string]interface{}{
			"usage_user":       100 * cts.User / totalDelta,
			"usage_system":     100 * cts.System / totalDelta,
			"usage_idle":       100 * cts.Idle / totalDelta,
			"usage_nice":       100 * cts.Nice / totalDelta,
			"usage_iowait":     100 * cts.Iowait / totalDelta,
			"usage_irq":        100 * cts.Irq / totalDelta,
			"usage_softirq":    100 * cts.Softirq / totalDelta,
			"usage_steal":      100 * cts.Steal / totalDelta,
			"usage_guest":      100 * cts.Guest / totalDelta,
			"usage_guest_nice": 100 * cts.GuestNice / totalDelta,
		}
  	influxdb.InfluxDBClient("cpu", fields, tags, conf)
  }

  total := cpuTotal[0].User + cpuTotal[0].System + cpuTotal[0].Nice + cpuTotal[0].Iowait + cpuTotal[0].Irq + cpuTotal[0].Softirq + cpuTotal[0].Steal + cpuTotal[0].Idle
  return 100 * (cpuTotal[0].System + cpuTotal[0].User ) / total, nil
}