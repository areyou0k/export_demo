package main

import (
	"flag"
	"net/http"
	"vmware_exporter/exporter"
	"vmware_exporter/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var (
	Version       = "ESXI监控"
	listenAddress = flag.String("web.listen-address", ":9601", "Address to listen on for web interface and telemetry.")
	metricPath    = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	landingPage   = []byte("<html><head><title>SYS Exporter " + Version +
		"</title></head><body><h1>SYS Exporter " + Version + "</h1><p><a href='" + *metricPath + "'>Metrics</a></p></body></html>")
)

func main() {
	//初始化日志
	logger.Setup()
	// 解析定义的监听端⼝等信息
	flag.Parse()

	exporter := exporter.NewExporter()
	prometheus.MustRegister(exporter)
	http.Handle(*metricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(landingPage)
	})
	log.Info().Msgf("Starting server %s", *listenAddress)
	log.Error().Msgf(" %s", http.ListenAndServe(*listenAddress, nil))
}
