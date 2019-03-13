package influxdb

import (
  "os"
  "log"
	"time"
	"encoding/json"

	"github.com/xxwassyxx/sensu-plugins-go/lib/tls"
	"github.com/influxdata/influxdb/client/v2"
)

type Config struct {
  Url string `json:"url"`
  DB string `json:"database"`
  User string `json:"user"`
  Password string `json:"password"`
  TlsCa string `json:"tls_ca"`
  TlsCert string `json:"tls_cert"`
  TlsKey string `json:"tls_key"`
  Insecure bool `json:"insecure_skip_verify"`
}

var config Config

func LoadConfiguration(file string) Config {
  configFile, err := os.Open(file)
  defer configFile.Close()
  if err != nil {
      log.Println(err.Error())
  }
  jsonParser := json.NewDecoder(configFile)
  jsonParser.Decode(&config)
  return config
}

func InfluxDBConnect(conf string) client.Client {
	config = LoadConfiguration(conf)

	tlsConfig, err := tls.TLSConfig(config.TlsCa, config.TlsCert ,config.TlsKey, config.Insecure)
	if err != nil {
		log.Fatalln("Error: ", err)
	}

  c, err := client.NewHTTPClient(client.HTTPConfig{
    Addr:     config.Url,
    Username: config.User,
    Password: config.Password,
    TLSConfig: tlsConfig,
  })
  if err != nil {
    log.Fatalln("Error: ", err)
  }
  return c
}

func InfluxDBClient(measurement_name string, fields map[string]interface{}, tags map[string]string, conf string) {
	c := InfluxDBConnect(conf)
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  config.DB,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a point and add to batch
	pt, err := client.NewPoint(measurement_name, tags, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := c.Write(bp); err != nil {
		log.Fatal(err)
	}

	// Close client resources
	if err := c.Close(); err != nil {
    		log.Fatal(err)
	}
}