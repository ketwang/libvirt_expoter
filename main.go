package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"libvirt_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main()  {
	prometheus.MustRegister(exporter.NewLibvirtExporter())
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":9999", nil); err != nil {
		fmt.Println(err)
	}
}