package main

import (
	"fmt"
  "net"
  "log"
	"strings"
  "encoding/json"

  "github.com/shirou/gopsutil/winservices"
  "github.com/xxwassyxx/sensu-plugins-go/lib/check"
)

type Output struct {
	Name     string   `json:"name"`
	Command  string   `json:"command"`
	Status   int      `json:"status"`
	Output   string   `json:"output"`
	Ttl      int      `json:"ttl,omitempty"`
	Kb   string   		`json:"kb,omitempty"`
	Handlers []string `json:"handlers,omitempty"`
}


type ServiceInfo struct {
	ServiceName string
	DisplayName string
	State       int
}

var conf  string

func main() {
	var (
		handlers string
		kb string
	)

	oserv := map[string]string{}

	c := check.New("windows_check_services")
	c.Option.StringVarP(&conf, "conf", "C", "influx.json", "CONF")
	c.Option.StringToStringVarP(&oserv, "serv", "s", map[string]string{}, "SERV")
	c.Option.StringVarP(&kb, "kb", "k", "389", "KB")
	c.Option.StringVarP(&handlers, "handlers", "H", "metrics", "HANLERS")
	c.Init()

	services, err := getStats()
	if err != nil {
		c.Error(err)
	}

	for k, v := range oserv {
		var sensu_values Output
		if stringInSlice(v, services) {
			sensu_values = Output{
				Name:     fmt.Sprintf("Check_services_%s", v),
				Status:   2,
				Output:   fmt.Sprintf("The service %s is currently not running", k),
				Kb:   		fmt.Sprintf("https://ww5.autotask.net/autotask/Knowledgebase/knowledgebase_view.aspx?isPopUp=1&objectID=%s", kb),
				Ttl:      60,
				Handlers: strings.Split(handlers, ","),
			}
		} else {
			sensu_values = Output{
				Name:     fmt.Sprintf("Check_services_%s", v),
				Status:   0,
				Output:   fmt.Sprintf("The service %s is running", k),
				Kb:   		fmt.Sprintf("https://ww5.autotask.net/autotask/Knowledgebase/knowledgebase_view.aspx?isPopUp=1&objectID=%s", kb),
				Ttl:      60,
				Handlers: strings.Split(handlers, ","),
			}
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
	c.Ok("Ran successfully")
}

func stringInSlice(a string, list []ServiceInfo) bool {
  for _, b := range list {
    if b.ServiceName == a {
      return true
    }
  }
  return false
}

func getStats() ([]ServiceInfo, error) {
	serviceInfo := make([]ServiceInfo, 0)
  services, err := winservices.ListServices()
  if err != nil {
		log.Println(err)
	}
	
	for _, svc := range services {
		s, err := winservices.NewService(svc.Name)
		if err != nil {
			continue
		}
		
		var ss winservices.ServiceStatus
		ss, err = s.QueryStatus()
		if ss.State != 4 && ss.State != 0 {
			serviceInfo = append(serviceInfo, ServiceInfo{ServiceName: svc.Name, State: int(ss.State)})
		}
	}

  return serviceInfo, nil
}