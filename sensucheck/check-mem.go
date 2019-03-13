package main

import (
    "fmt"
    "os"
    "log"

    "github.com/shirou/gopsutil/mem"
    "github.com/xxwassyxx/sensu-plugins-go/lib/check"
    "github.com/xxwassyxx/sensu-plugins-go/lib/influxdb"
)

var conf  string

func main() {
	var (
		warn  int
		crit  int
	)

	c := check.New("windows_check_mem")
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

  vm, err := mem.VirtualMemory()
  if err != nil {
		log.Println("error %v", err)
	}

	fields := map[string]interface{}{
		"total":             int64(vm.Total),
		"available":         int64(vm.Available),
		"used":              int64(vm.Used),
		"free":              int64(vm.Free),
		"cached":            int64(vm.Cached),
		"buffered":          int64(vm.Buffers),
		"active":            int64(vm.Active),
		"inactive":          int64(vm.Inactive),
		"wired":             int64(vm.Wired),
		"slab":              int64(vm.Slab),
		"used_percent":      100 * float64(vm.Used) / float64(vm.Total),
		"available_percent": 100 * float64(vm.Available) / float64(vm.Total),
		"commit_limit":      int64(vm.CommitLimit),
		"committed_as":      int64(vm.CommittedAS),
		"dirty":             int64(vm.Dirty),
		"high_free":         int64(vm.HighFree),
		"high_total":        int64(vm.HighTotal),
		"huge_page_size":    int64(vm.HugePageSize),
		"huge_pages_free":   int64(vm.HugePagesFree),
		"huge_pages_total":  int64(vm.HugePagesTotal),
		"low_free":          int64(vm.LowFree),
		"low_total":         int64(vm.LowTotal),
		"mapped":            int64(vm.Mapped),
		"page_tables":       int64(vm.PageTables),
		"shared":            int64(vm.Shared),
		"swap_cached":       int64(vm.SwapCached),
		"swap_free":         int64(vm.SwapFree),
		"swap_total":        int64(vm.SwapTotal),
		"vmalloc_chunk":     int64(vm.VMallocChunk),
		"vmalloc_total":     int64(vm.VMallocTotal),
		"vmalloc_used":      int64(vm.VMallocUsed),
		"write_back":        int64(vm.Writeback),
		"write_back_tmp":    int64(vm.WritebackTmp),
	}
	tags := map[string]string{"host": name}
  influxdb.InfluxDBClient("mem", fields, tags, conf)

  return vm.UsedPercent, nil
}