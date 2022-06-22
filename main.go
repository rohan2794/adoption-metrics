package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	//"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"

	//"google.golang.org/appengine/log"
	yaml "gopkg.in/yaml.v2"
)

var (
	config Config
)

const (
	collector = "query_exporter"
)

// =============================
// Config config structure
// =============================
type Config struct {
	Metrics map[string]struct {
		Api         string
		Type        string
		Description string
		Labels      []string
		Value       string
		metricDesc  *prometheus.Desc
	}
}

type AdoptionMetric struct {
	User               string      `json:"user"`
	Name               string      `json:"name"`
	Namespace          string      `json:"namespace"`
	Repository_type    string      `json:"repository_type"`
	Status             int         `json:"status"`
	Description        string      `json:"description"`
	Is_private         bool        `json:"is_private"`
	Is_automated       bool        `json:"is_automated"`
	Can_edit           bool        `json:"can_edit"`
	Star_count         int         `json:"star_count"`
	Pull_count         int64       `json:"pull_count"`
	Last_updated       string      `json:"last_updated"`
	Date_registered    string      `json:"date_registered"`
	Collaborator_count int         `json:"collaborator_count"`
	Affiliation        string      `json:"affiliation"`
	Hub_user           string      `json:"hub_user"`
	Has_starred        bool        `json:"has_starred"`
	Full_description   string      `json:"full_description"`
	Permissions        Permissions `json:"permissions"`
}

type Permissions struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
	Admin bool `json:"admin"`
}

func main() {
	var err error
	var configFile, bind string
	// =====================
	// Get OS parameter
	// =====================
	flag.StringVar(&configFile, "config", "./config.yml", "configuration file")
	flag.StringVar(&bind, "bind", "0.0.0.0:9104", "bind")
	flag.Parse()

	// =====================
	// Load config & yaml
	// =====================
	var b []byte
	if b, err = ioutil.ReadFile(configFile); err != nil {
		log.Errorf("Failed to read config file: %s", err)
		os.Exit(1)
	}

	// Load yaml
	if err = yaml.Unmarshal(b, &config); err != nil {
		log.Errorf("Failed to load config: %s", err)
		os.Exit(1)
	}

	// ========================
	// Regist handler
	// ========================
	log.Infof("Regist version collector - %s", collector)
	prometheus.Register(version.NewCollector(collector))
	prometheus.Register(&QueryCollector{})

	// Regist http handler
	log.Infof("HTTP handler path - %s", "/metrics")
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		h := promhttp.HandlerFor(prometheus.Gatherers{
			prometheus.DefaultGatherer,
		}, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	})

	// start server
	log.Infof("Starting http server - %s", bind)
	if err = http.ListenAndServe(bind, nil); err != nil {
		log.Errorf("Failed to start http server: %s", err)
	}

}

// =============================
// QueryCollector exporter
// =============================
type QueryCollector struct{}

// Describe prometheus describe
func (e *QueryCollector) Describe(ch chan<- *prometheus.Desc) {
	for metricName, metric := range config.Metrics {
		metric.metricDesc = prometheus.NewDesc(
			prometheus.BuildFQName(collector, "", metricName),
			metric.Description,
			metric.Labels, nil,
		)
		config.Metrics[metricName] = metric
		log.Infof("metric description for \"%s\" registerd", metricName)
	}
}

// Collect prometheus collect
func (e *QueryCollector) Collect(ch chan<- prometheus.Metric) {

	// Execute each queries in metrics
	for name, metric := range config.Metrics {
		data, err := GetMetrics(metric.Api)
		if err != nil {
			log.Errorf("Failed to get mmetrics, API call failed, API: %s, Error: %v", name, err)
		}
		// Metric labels
		labelVals := []string{}

		labelVals = append(labelVals, name)

		// Metric value
		val := float64(data.Pull_count)

		// Add metric
		switch strings.ToLower(metric.Type) {
		case "counter":
			ch <- prometheus.MustNewConstMetric(metric.metricDesc, prometheus.CounterValue, val, labelVals...)
		case "gauge":
			ch <- prometheus.MustNewConstMetric(metric.metricDesc, prometheus.GaugeValue, val, labelVals...)
		default:
			log.Errorf("Fail to add metric for %s: %s is not valid type", name, metric.Type)
			continue
		}
	}
}

func GetMetrics(api string) (AdoptionMetric, error) {
	if api == "" {
		return AdoptionMetric{}, fmt.Errorf("API not found")
	}
	var jsonResponse []byte
	var err error

	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		log.Info("Error in GET request. API: %s, Error: %v", api, err)
	}
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Info("Error while making GET request,API: %s, Error: %v", api, err)
	} else {
		defer resp.Body.Close()
		jsonResponse, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Info("Error while reading data, Error %v", err)
		}
	}
	if err != nil {
		return AdoptionMetric{}, err
	}
	var response AdoptionMetric
	err = json.Unmarshal(jsonResponse, &response)
	if err != nil {
		log.Info("Failed to unmarshal", "string", string(jsonResponse))
		return AdoptionMetric{}, err
	}
	return response, nil
}
