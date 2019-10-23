package main

import (
	"fmt"
	"libvirt-exporter/exporter"
	"net/http"
	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	prometheus.MustRegister(exporter.NewLibvirtExporter())
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":9999", nil); err != nil {
		fmt.Println(err)
	}
}
