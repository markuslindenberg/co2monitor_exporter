package main

import (
	"net/http"

	"github.com/markuslindenberg/co2monitor_exporter/co2monitor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	exporterName = "co2monitor_exporter"
	namespace    = "co2monitor"
)

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9673").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		device        = kingpin.Flag("device", "hidraw device").Default("").String()
	)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print(exporterName))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting", exporterName, version.Info())
	log.Infoln("Build context", version.BuildContext())

	co2 := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "co2_ppm",
	})
	temp := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "temp_celsius",
	})
	humidity := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "humidity_rh",
	})

	monitor, err := co2monitor.Open(*device)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			op, value, err := monitor.Read()
			if err != nil {
				log.Fatal(err)
			}
			switch op {
			case co2monitor.OpCo2:
				co2.Set(float64(value))
			case co2monitor.OpTemp:
				temp.Set(co2monitor.TempToCelsius(value))
			case co2monitor.OpHum:
				humidity.Set(co2monitor.HumidityToRH(value))
			}
		}
	}()

	log.Infoln("Listening on", *listenAddress)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>CO2 Monitor Exporter</title></head>
             <body>
             <h1>CO2 Monitor Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
