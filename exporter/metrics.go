package exporter

import (
	"io/ioutil"
	"libvirt-exporter/pool"
	"libvirt-exporter/util"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"libvirt-exporter/dom"
)

type LibvirtExporter struct {
	LibvirtUp *prometheus.Desc

	LibvirtDomain                *prometheus.Desc
	LibvirtDomainUp              *prometheus.Desc
	LibvirtDomainConfigMem       *prometheus.Desc
	LibvirtDomainConfigCPUs      *prometheus.Desc
	LibvirtDomainVCPURunningTime *prometheus.Desc
	LibvirtDomainVCPUStealTime   *prometheus.Desc
	LibvirtDomainNicRxBytes      *prometheus.Desc
	LibvirtDomainNicRxPackets    *prometheus.Desc
	LibvirtDomainNicTxBytes      *prometheus.Desc
	LibvirtDomainNicTxPackets    *prometheus.Desc

	LibvirtPool          *prometheus.Desc
	LibvirtPoolUP        *prometheus.Desc
	LibvirtPoolCapacity  *prometheus.Desc
	LibvirtPoolAllocated *prometheus.Desc
}

func (e *LibvirtExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.LibvirtUp

	ch <- e.LibvirtDomain
	ch <- e.LibvirtDomainUp
	ch <- e.LibvirtDomainConfigMem
	ch <- e.LibvirtDomainConfigCPUs
	ch <- e.LibvirtDomainVCPURunningTime
	ch <- e.LibvirtDomainVCPUStealTime
	ch <- e.LibvirtDomainNicTxBytes
	ch <- e.LibvirtDomainNicTxPackets
	ch <- e.LibvirtDomainNicRxBytes
	ch <- e.LibvirtDomainNicRxPackets

	ch <- e.LibvirtPool
	ch <- e.LibvirtPoolUP
	ch <- e.LibvirtPoolAllocated
	ch <- e.LibvirtPoolCapacity
}

func (e *LibvirtExporter) Collect(ch chan<- prometheus.Metric) {
	d := dom.DomObj{}
	rtDoms, err := d.GetDoms()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtUp,
			prometheus.GaugeValue,
			0)
		return
	}

	p := pool.PoolObj{}
	pools, err := p.GetPools()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtUp,
			prometheus.GaugeValue,
			0,
		)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		e.LibvirtUp,
		prometheus.GaugeValue,
		1)

	g := sync.WaitGroup{}

	for _, rtDom := range rtDoms {
		g.Add(1)
		go func(rtDom dom.Dom) {
			defer g.Done()
			domMeta, err := rtDom.GetOverallState()
			if err != nil {
				panic(err)
			} else {
				ch <- prometheus.MustNewConstMetric(
					e.LibvirtDomain,
					prometheus.GaugeValue,
					1.0,
					[]string{domMeta.Name, domMeta.UUID}...,
				)
			}

			e.CollectDomain(domMeta, ch)
		}(rtDom)
	}

	for _, p := range pools {
		g.Add(1)
		go func(p pool.Pool) {
			defer g.Done()
			poolMeta, err := p.GetOverallState()
			if err != nil {
				panic(err)
			}

			ch <- prometheus.MustNewConstMetric(
				e.LibvirtPool,
				prometheus.GaugeValue,
				1.0,
				[]string{poolMeta.UUID, poolMeta.Name}...,
			)

			e.CollectPool(poolMeta, ch)
		}(p)
	}

	g.Wait()
}

func (e *LibvirtExporter) CollectPool(poolMeta *pool.PoolMeta, ch chan<- prometheus.Metric) {
	if poolMeta.State != pool.StateRunning {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtPoolUP,
			prometheus.GaugeValue,
			0,
			[]string{poolMeta.Name, poolMeta.UUID}...,
		)

		return
	}

	ch <- prometheus.MustNewConstMetric(
		e.LibvirtPoolUP,
		prometheus.GaugeValue,
		1,
		[]string{poolMeta.Name, poolMeta.UUID}...,
	)

	v, err := util.ConvertToBytes(poolMeta.Capacity)
	if err != nil {
		return
	}
	ch <- prometheus.MustNewConstMetric(
		e.LibvirtPoolCapacity,
		prometheus.GaugeValue,
		v,
		[]string{poolMeta.Name, poolMeta.UUID}...,
	)

	v, err = util.ConvertToBytes(poolMeta.Allocation)
	if err != nil {
		return
	}
	ch <- prometheus.MustNewConstMetric(
		e.LibvirtPoolAllocated,
		prometheus.GaugeValue,
		v,
		[]string{poolMeta.Name, poolMeta.UUID}...,
	)

}

func (e *LibvirtExporter) CollectDomain(domMeta *dom.DomMeta, ch chan<- prometheus.Metric) {
	name := domMeta.Name
	uuid := domMeta.UUID

	if domMeta.State != dom.STATE_RUNNING {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtDomainUp,
			prometheus.GaugeValue,
			0,
			[]string{name, uuid}...,
		)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		e.LibvirtDomainUp,
		prometheus.GaugeValue,
		1,
		[]string{name, uuid}...,
	)

	mem, err := util.ConvertToBytes(domMeta.Mem)
	if err != nil {
		return
	}
	ch <- prometheus.MustNewConstMetric(
		e.LibvirtDomainConfigMem,
		prometheus.GaugeValue,
		mem,
		[]string{name, uuid}...,
	)

	cpuNum, err := strconv.ParseFloat(domMeta.CPU, 64)
	if err != nil {
		return
	}
	ch <- prometheus.MustNewConstMetric(
		e.LibvirtDomainConfigCPUs,
		prometheus.GaugeValue,
		cpuNum,
		[]string{name, uuid}...,
	)

	for _, vcpu := range domMeta.VCpus.VCPUs {
		path := "/proc/" + vcpu.PID + "/schedstat"

		content, err := ioutil.ReadFile(path)
		if err != nil {
			continue
		}

		items := strings.Fields(string(content))
		if len(items) != 3 {
			continue
		}

		runningTime, err := strconv.ParseFloat(items[0], 64)
		if err != nil {
			continue
		}

		stealTime, err := strconv.ParseFloat(items[1], 64)
		if err != nil {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.LibvirtDomainVCPURunningTime,
			prometheus.GaugeValue,
			runningTime,
			[]string{name, uuid, vcpu.Index}...,
		)

		ch <- prometheus.MustNewConstMetric(
			e.LibvirtDomainVCPUStealTime,
			prometheus.GaugeValue,
			stealTime,
			[]string{name, uuid, vcpu.Index}...,
		)
	}

	content, err := ioutil.ReadFile("/proc/net/dev")
	if err != nil {
		return
	}

	for _, nic := range domMeta.Nics {
		for _, line := range strings.Split(string(content), "\n") {
			if !strings.HasPrefix(strings.TrimSpace(line), nic.Interface) {
				continue
			}

			items := strings.Fields(line)
			if len(items) != 17 {
				continue
			}

			txBytes, err := strconv.ParseFloat(items[1], 64)
			if err != nil {
				continue
			}

			txPackets, err := strconv.ParseFloat(items[2], 64)
			if err != nil {
				continue
			}

			rxBytes, err := strconv.ParseFloat(items[9], 64)
			if err != nil {
				continue
			}

			rxPackets, err := strconv.ParseFloat(items[10], 64)
			if err != nil {
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				e.LibvirtDomainNicRxBytes,
				prometheus.GaugeValue,
				rxBytes,
				[]string{name, uuid, nic.Interface}...,
			)

			ch <- prometheus.MustNewConstMetric(
				e.LibvirtDomainNicRxPackets,
				prometheus.GaugeValue,
				rxPackets,
				[]string{name, uuid, nic.Interface}...,
			)

			ch <- prometheus.MustNewConstMetric(
				e.LibvirtDomainNicTxBytes,
				prometheus.GaugeValue,
				txBytes,
				[]string{name, uuid, nic.Interface}...,
			)

			ch <- prometheus.MustNewConstMetric(
				e.LibvirtDomainNicTxPackets,
				prometheus.GaugeValue,
				txPackets,
				[]string{name, uuid, nic.Interface}...,
			)

			break
		}
	}
}

func NewLibvirtExporter() *LibvirtExporter {
	domainLabels := []string{"name", "uuid"}
	poolLabels := []string{"name", "uuid"}

	return &LibvirtExporter{
		// libvirt
		LibvirtUp: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "", "up"),
			"Whether scraping libvirt's metrics was successful.",
			nil,
			nil),

		// domain
		LibvirtDomain: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain", "metrics"),
			"Whether scraping domain metrics was successful.",
			domainLabels,
			nil,
		),

		LibvirtDomainUp: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain", "up"),
			"Whether domain is up.",
			domainLabels,
			nil,
		),

		// memory
		LibvirtDomainConfigMem: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain", "config_memory_bytes"),
			"Current allowed memory of the domain, in bytes.",
			domainLabels,
			nil),

		// cpu
		LibvirtDomainConfigCPUs: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain", "config_cpu_num"),
			"Current allowed cpu number of the domain.",
			domainLabels,
			nil),

		LibvirtDomainVCPURunningTime: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain", "vcpu_running_time_ns_total"),
			"all vcpu used user time, in ns.",
			append(domainLabels, "index"),
			nil),

		LibvirtDomainVCPUStealTime: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain", "vcpu_steal_time_ns_total"),
			"all vcpu used system time, in ns.",
			append(domainLabels, "index"),
			nil,
		),

		LibvirtDomainNicRxBytes: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain", "nic_rx_bytes_total"),
			"all rx bytes total",
			append(domainLabels, "nic"),
			nil,
		),

		LibvirtDomainNicRxPackets: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain", "nic_rx_packets_total"),
			"all rx packets total",
			append(domainLabels, "nic"),
			nil,
		),

		LibvirtDomainNicTxBytes: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain", "nic_tx_bytes_total"),
			"all tx bytes toal",
			append(domainLabels, "nic"),
			nil,
		),

		LibvirtDomainNicTxPackets: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain", "nic_tx_packets_total"),
			"all tx packets total",
			append(domainLabels, "nic"),
			nil,
		),

		LibvirtPool: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "pool", "metrics"),
			"Whether scraping pool metrics was successful.",
			poolLabels,
			nil,
		),
		LibvirtPoolUP: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "pool", "up"),
			"Whether domain is up.",
			poolLabels,
			nil,
		),
		LibvirtPoolCapacity: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "pool", "capacity"),
			"pool capacity",
			poolLabels,
			nil,
		),
		LibvirtPoolAllocated: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "pool", "allocated"),
			"pool allocated",
			poolLabels,
			nil,
		),
	}
}
