package main

import (
  "fmt"
  "os"
  "net"
  "log"
	"strings"
  "encoding/json"

  "github.com/shirou/gopsutil/disk"
  "golang.org/x/sys/windows"
  "github.com/xxwassyxx/sensu-plugins-go/lib/check"
  "github.com/xxwassyxx/sensu-plugins-go/lib/influxdb"
)

type DiskInfo struct {
	Path string
	Free int
}

type Output struct {
	Name     string   `json:"name"`
	Command  string   `json:"command"`
	Status   int      `json:"status"`
	Output   string   `json:"output"`
	Ttl      int      `json:"ttl,omitempty"`
	Kb   string   		`json:"kb,omitempty"`
	Handlers []string `json:"handlers,omitempty"`
}

var conf  string

func main() {
	var (
		warn  int
		crit  int
		sensu_values Output
		handlers string
		kb string
		excl string
	)

	c := check.New("windows_check_disk")
	c.Option.IntVarP(&warn, "warn", "w", 9000, "WARN")
	c.Option.IntVarP(&crit, "crit", "c", 7000, "CRIT")
	c.Option.StringVarP(&conf, "conf", "C", "influx.json", "CONF")
	c.Option.StringVarP(&excl, "excl", "e", "389", "EXCLUDE")
	c.Option.StringVarP(&kb, "kb", "k", "389", "KB")
	c.Option.StringVarP(&handlers, "handlers", "H", "metrics", "HANLERS")
	c.Init()

	Excl := strings.Split(excl, ",")
	usage, err := getStats()
	if err != nil {
		c.Error(err)
	}
	for _, d := range usage {
		if stringInSlice(d.Path, Excl) == false {
			switch {
			case d.Free <= int(crit):
				sensu_values = Output{
					Name:     fmt.Sprintf("windows_check_disk_%s", strings.TrimSuffix(d.Path, ":\\")),
					Status:   2,
					Output:   fmt.Sprintf("Disk %s is currently at %d free", d.Path, d.Free),
					Kb:   		fmt.Sprintf("https://ww5.autotask.net/autotask/Knowledgebase/knowledgebase_view.aspx?isPopUp=1&objectID=%s", kb),
					Ttl:      60,
					Handlers: strings.Split(handlers, ","),
				}
				var output_json []byte
				output_json, _ = json.Marshal(sensu_values)
				conn, err := net.Dial("udp", "127.0.0.1:3030")
				if err != nil {
					log.Println("Problem sending JSON to socket: %v", err)
				} else {
					fmt.Fprintf(conn, string(output_json))
				}
			case d.Free <= int(warn):
				sensu_values = Output{
					Name:     fmt.Sprintf("windows_check_disk_%s", strings.TrimSuffix(d.Path, ":\\")),
					Status:   1,
					Output:   fmt.Sprintf("Disk %s is currently at %d free", d.Path, d.Free),
					Kb:   		fmt.Sprintf("https://ww5.autotask.net/autotask/Knowledgebase/knowledgebase_view.aspx?isPopUp=1&objectID=%s", kb),
					Ttl:      60,
					Handlers: strings.Split(handlers, ","),
				}
				var output_json []byte
				output_json, _ = json.Marshal(sensu_values)
				conn, err := net.Dial("udp", "127.0.0.1:3030")
				if err != nil {
					log.Println("Problem sending JSON to socket: %v", err)
				} else {
					fmt.Fprintf(conn, string(output_json))
				}
			default:
				sensu_values = Output{
					Name:     fmt.Sprintf("windows_check_disk_%s", strings.TrimSuffix(d.Path, ":\\")),
					Status:   0,
					Output:   fmt.Sprintf("Disk %s is currently at %d free", d.Path, d.Free),
					Kb:   		fmt.Sprintf("https://ww5.autotask.net/autotask/Knowledgebase/knowledgebase_view.aspx?isPopUp=1&objectID=%s", kb),
					Ttl:      60,
					Handlers: strings.Split(handlers, ","),
				}
				var output_json []byte
				output_json, _ = json.Marshal(sensu_values)
				conn, err := net.Dial("udp", "127.0.0.1:3030")
				if err != nil {
					log.Println("Problem sending JSON to socket: %v", err)
				} else {
					fmt.Fprintf(conn, string(output_json))
				}
			}
		}
	}
	c.Ok("Ran successfully")
}

func stringInSlice(a string, list []string) bool {
  for _, b := range list {
    if b == a {
      return true
    }
  }
  return false
}

func getdrives() (r []string){
  for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ"{
    _, err := os.Open(string(drive)+":\\")
    if err == nil {
      r = append(r, string(drive))
    }
  }
  return
}

func getStats() ([]DiskInfo, error) {
	diskinfo := make([]DiskInfo, 0)
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	drives := getdrives()
	for _, d := range drives {
		if windows.GetDriveType(windows.StringToUTF16Ptr(d + ":\\")) == 3 {
			disk, err := disk.Usage(d + ":\\")
			if err != nil {
				log.Println(err)
				return nil, err
			}
			fields := map[string]interface{}{
				"total":        int(disk.Total)/1024/1024,
				"free":         int(disk.Free)/1024/1024,
				"used":         int(disk.Used)/1024/1024,
				"used_percent": disk.UsedPercent,
				"inodes_total": int(disk.InodesTotal),
				"inodes_free":  int(disk.InodesFree),
				"inodes_used":  int(disk.InodesUsed),
				"inodes_used_percent":  disk.InodesUsedPercent,
			}
			tags := map[string]string{
				"host": name,
				"path": strings.TrimSuffix(disk.Path, "\\"),
			}
			influxdb.InfluxDBClient("disk", fields, tags, conf)
			diskinfo = append(diskinfo, DiskInfo{Path: disk.Path, Free: int(disk.Free)/1024/1024})
		}
	}
  return diskinfo, nil
}