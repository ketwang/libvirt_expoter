package exporter

import (
	"fmt"
	"libvirt_exporter/pool"
	"strconv"

	"libvirt_exporter/dom"
	"github.com/prometheus/client_golang/prometheus"
)

type LibvirtExporter struct {
	LibvirtUpDesc                 *prometheus.Desc
	LibvirtDomainInfo             *prometheus.Desc
	LibvirtDomainInfoMaxMem       *prometheus.Desc
	LibvirtDomainInfoCurrMem      *prometheus.Desc
	LibvirtDomainInfoCpuTime      *prometheus.Desc
	LibvirtDomainInfoCpuUser      *prometheus.Desc
	LibvirtDomainInfoCpuSystem    *prometheus.Desc
	LibvirtDomainInfoVcpuCurrent  *prometheus.Desc
	LibvirtDomainInfoVcpuMaximum  *prometheus.Desc
	LibvirtDomainInfoVcpustate    *prometheus.Desc
	LibvirtDomainInfoVcputime     *prometheus.Desc
	LibvirtDomainInfoNetCount     *prometheus.Desc
	LibvirtDomainInfoNetRxBytes   *prometheus.Desc
	LibvirtDomainInfoNetRxPkts    *prometheus.Desc
	LibvirtDomainInfoNetRxErrs    *prometheus.Desc
	LibvirtDomainInfoNetRxDrop    *prometheus.Desc
	LibvirtDomainInfoNetTxBytes   *prometheus.Desc
	LibvirtDomainInfoNetTxPkts    *prometheus.Desc
	LibvirtDomainInfoNetTxErrs    *prometheus.Desc
	LibvirtDomainInfoNetTxDrop    *prometheus.Desc
	LibvirtDomainInfoBlockCount   *prometheus.Desc
	LibvirtDomainInfoBlockPath    *prometheus.Desc
	LibvirtDomainInfoBlockRdReqs  *prometheus.Desc
	LibvirtDomainInfoBlockRdBytes *prometheus.Desc
	LibvirtDomainInfoBlockRdTimes *prometheus.Desc
	LibvirtDomainInfoBlockWrReqs  *prometheus.Desc
	LibvirtDomainInfoBlockWrBytes *prometheus.Desc
	LibvirtDomainInfoBlockWrTimes *prometheus.Desc
	LibvirtDomainInfoBlockFlReqs  *prometheus.Desc
	LibvirtDomainInfoBlockFlTimes *prometheus.Desc
	LibvirtPoolInfo				  *prometheus.Desc
	LibvirtPoolInfoStatus		  *prometheus.Desc
	LibvirtPoolInfoCapacity       *prometheus.Desc
	LibvirtPoolInfoAllocated      *prometheus.Desc

}

func (e *LibvirtExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.LibvirtUpDesc
	ch <- e.LibvirtDomainInfo
	ch <- e.LibvirtDomainInfoMaxMem
	ch <- e.LibvirtDomainInfoCurrMem
	ch <- e.LibvirtDomainInfoCpuTime
	ch <- e.LibvirtDomainInfoCpuUser
	ch <- e.LibvirtDomainInfoCpuSystem
	ch <- e.LibvirtDomainInfoVcpuCurrent
	ch <- e.LibvirtDomainInfoVcpuMaximum
	ch <- e.LibvirtDomainInfoVcpustate
	ch <- e.LibvirtDomainInfoVcputime
	ch <- e.LibvirtDomainInfoNetCount
	ch <- e.LibvirtDomainInfoNetRxBytes
	ch <- e.LibvirtDomainInfoNetRxPkts
	ch <- e.LibvirtDomainInfoNetRxErrs
	ch <- e.LibvirtDomainInfoNetRxDrop
	ch <- e.LibvirtDomainInfoNetTxBytes
	ch <- e.LibvirtDomainInfoNetTxPkts
	ch <- e.LibvirtDomainInfoNetTxErrs
	ch <- e.LibvirtDomainInfoNetTxDrop
	ch <- e.LibvirtDomainInfoBlockCount
	ch <- e.LibvirtDomainInfoBlockPath
	ch <- e.LibvirtDomainInfoBlockRdReqs
	ch <- e.LibvirtDomainInfoBlockRdBytes
	ch <- e.LibvirtDomainInfoBlockRdTimes
	ch <- e.LibvirtDomainInfoBlockWrReqs
	ch <- e.LibvirtDomainInfoBlockWrBytes
	ch <- e.LibvirtDomainInfoBlockWrTimes
	ch <- e.LibvirtDomainInfoBlockFlReqs
	ch <- e.LibvirtDomainInfoBlockFlTimes
	ch <- e.LibvirtPoolInfo
	ch <- e.LibvirtPoolInfoStatus
	ch <- e.LibvirtPoolInfoCapacity
	ch <- e.LibvirtPoolInfoAllocated
}

func (e *LibvirtExporter) Collect(ch chan<- prometheus.Metric) {
	d := dom.DomObj{}
	rtDoms, err := d.GetDoms()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtUpDesc,
			prometheus.GaugeValue,
			1.0)
	} else {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtUpDesc,
			prometheus.GaugeValue,
			0.0)
		return
	}

	for _, rtDom := range rtDoms {
		domMeta, err := rtDom.GetOverallState()
		if err != nil {
			ch <- prometheus.MustNewConstMetric(
				e.LibvirtDomainInfo,
				prometheus.GaugeValue,
				0.0,
				[]string{domMeta.Name, domMeta.UUID}...,
			)
			return
		} else {
			ch <- prometheus.MustNewConstMetric(
				e.LibvirtDomainInfo,
				prometheus.GaugeValue,
				1.0,
				[]string{domMeta.Name, domMeta.UUID}...,
			)
		}

		e.CollectDomain(domMeta, ch)
	}

	p := pool.PoolObj{}
	pools, err := p.GetPools()
	if err != nil {
		return
	}

	ch <- prometheus.MustNewConstMetric(
		e.LibvirtPoolInfo,
		prometheus.GaugeValue,
		float64(len(pools)),
		)
	for _, p := range pools {
		poolMeta, err := p.GetOverallState()
		if err != nil {
			return
		}

		e.CollectPool(poolMeta, ch)
	}
}

func (e *LibvirtExporter) CollectPool(poolMeta *pool.PoolMeta, ch chan <- prometheus.Metric)  {
	ch <- prometheus.MustNewConstMetric(
		e.LibvirtPoolInfoCapacity,
		prometheus.GaugeValue,
		func() float64 {
			v, err := pool.ConvertToBytes(poolMeta.Capacity)
			if err != nil {
				return 0
			}
			return float64(v)
		}(),
		[]string{poolMeta.Name}...,
		)

	ch <- prometheus.MustNewConstMetric(
		e.LibvirtPoolInfoAllocated,
		prometheus.GaugeValue,
		func() float64 {
			v, err := pool.ConvertToBytes(poolMeta.Allocation)
			if err != nil {
				return 0
			}
			return float64(v)
		}(),
		[]string{poolMeta.Name}...,
		)

	ch <- prometheus.MustNewConstMetric(
		e.LibvirtPoolInfoStatus,
		prometheus.GaugeValue,
		func() float64 {
			if poolMeta.State != "running" {
				return 0
			}
			return 1
		}(),
		[]string{poolMeta.Name}...,
		)
}

func (e *LibvirtExporter) CollectDomain(domMeta *dom.DomMeta, ch chan<- prometheus.Metric) {
	domain := domMeta.Name
	uuid := domMeta.UUID

	if v, ok := domMeta.Annotations["balloon.maximum"]; ok {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtDomainInfoMaxMem,
			prometheus.GaugeValue,
			ConvertToFloat64(v),
			[]string{domain, uuid}...,
		)
	}

	if v, ok := domMeta.Annotations["balloon.current"]; ok {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtDomainInfoCurrMem,
			prometheus.GaugeValue,
			ConvertToFloat64(v),
			[]string{domain, uuid}...,
		)
	}

	if v, ok := domMeta.Annotations["cpu.time"]; ok {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtDomainInfoCpuTime,
			prometheus.GaugeValue,
			ConvertToFloat64(v),
			[]string{domain, uuid}...,
		)
	}

	if v, ok := domMeta.Annotations["cpu.user"]; ok {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtDomainInfoCpuUser,
			prometheus.GaugeValue,
			ConvertToFloat64(v),
			[]string{domain, uuid}...,
		)
	}

	if v, ok := domMeta.Annotations["cpu.system"]; ok {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtDomainInfoCpuSystem,
			prometheus.GaugeValue,
			ConvertToFloat64(v),
			[]string{domain, uuid}...,
		)
	}

	if v, ok := domMeta.Annotations["vcpu.maximum"]; ok {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtDomainInfoVcpuMaximum,
			prometheus.GaugeValue,
			ConvertToFloat64(v),
			[]string{domain, uuid}...,
		)
	}

	if v, ok := domMeta.Annotations["vcpu.current"]; ok {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtDomainInfoVcpuCurrent,
			prometheus.GaugeValue,
			ConvertToFloat64(v),
			[]string{domain, uuid}...,
		)

		for i := int64(0); i < int64(ConvertToFloat64(v)); i++ {
			if v, ok := domMeta.Annotations[fmt.Sprintf("vcpu.%d.state", i)]; ok {
				ch <- prometheus.MustNewConstMetric(
					e.LibvirtDomainInfoVcpustate,
					prometheus.GaugeValue,
					ConvertToFloat64(v),
					[]string{domain, uuid, strconv.FormatInt(i, 10)}...,
				)
			}

			if v, ok := domMeta.Annotations[fmt.Sprintf("vcpu.%d.time", i)]; ok {
				ch <- prometheus.MustNewConstMetric(
					e.LibvirtDomainInfoVcputime,
					prometheus.GaugeValue,
					ConvertToFloat64(v),
					[]string{domain, uuid, strconv.FormatInt(i, 10)}...,
				)
			}

		}
	}

	if v, ok := domMeta.Annotations["net.count"]; ok {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtDomainInfoNetCount,
			prometheus.GaugeValue,
			ConvertToFloat64(v),
			[]string{domain, uuid}...,
		)

		for i := int64(0); i < int64(ConvertToFloat64(v)); i++ {
			if name, ok := domMeta.Annotations[fmt.Sprintf("net.%d.name", i)]; ok {
				if v, ok := domMeta.Annotations[fmt.Sprintf("net.%d.rx.bytes", i)]; ok {
					ch <- prometheus.MustNewConstMetric(
						e.LibvirtDomainInfoNetRxBytes,
						prometheus.GaugeValue,
						ConvertToFloat64(v),
						[]string{domain, uuid, name}...,
					)
				}

				if v, ok := domMeta.Annotations[fmt.Sprintf("net.%d.rx.pkts", i)]; ok {
					ch <- prometheus.MustNewConstMetric(
						e.LibvirtDomainInfoNetRxPkts,
						prometheus.GaugeValue,
						ConvertToFloat64(v),
						[]string{domain, uuid, name}...,
					)
				}

				if v, ok := domMeta.Annotations[fmt.Sprintf("net.%d.rx.errs", i)]; ok {
					ch <- prometheus.MustNewConstMetric(
						e.LibvirtDomainInfoNetRxErrs,
						prometheus.GaugeValue,
						ConvertToFloat64(v),
						[]string{domain, uuid, name}...,
					)
				}

				if v, ok := domMeta.Annotations[fmt.Sprintf("net.%d.rx.drop", i)]; ok {
					ch <- prometheus.MustNewConstMetric(
						e.LibvirtDomainInfoNetRxDrop,
						prometheus.GaugeValue,
						ConvertToFloat64(v),
						[]string{domain, uuid, name}...,
					)
				}

				if v, ok := domMeta.Annotations[fmt.Sprintf("net.%d.tx.bytes", i)]; ok {
					ch <- prometheus.MustNewConstMetric(
						e.LibvirtDomainInfoNetTxBytes,
						prometheus.GaugeValue,
						ConvertToFloat64(v),
						[]string{domain, uuid, name}...,
					)
				}

				if v, ok := domMeta.Annotations[fmt.Sprintf("net.%d.tx.pkts", i)]; ok {
					ch <- prometheus.MustNewConstMetric(
						e.LibvirtDomainInfoNetTxPkts,
						prometheus.GaugeValue,
						ConvertToFloat64(v),
						[]string{domain, uuid, name}...,
					)
				}

				if v, ok := domMeta.Annotations[fmt.Sprintf("net.%d.tx.errs", i)]; ok {
					ch <- prometheus.MustNewConstMetric(
						e.LibvirtDomainInfoNetTxErrs,
						prometheus.GaugeValue,
						ConvertToFloat64(v),
						[]string{domain, uuid, name}...,
					)
				}

				if v, ok := domMeta.Annotations[fmt.Sprintf("net.%d.tx.drop", i)]; ok {
					ch <- prometheus.MustNewConstMetric(
						e.LibvirtDomainInfoNetTxDrop,
						prometheus.GaugeValue,
						ConvertToFloat64(v),
						[]string{domain, uuid, name}...,
					)
				}
			}
		}
	}

	if v, ok := domMeta.Annotations["block.count"]; ok {
		ch <- prometheus.MustNewConstMetric(
			e.LibvirtDomainInfoBlockCount,
			prometheus.GaugeValue,
			ConvertToFloat64(v),
			[]string{domain, uuid}...,
		)

		for i := int64(0); i < int64(ConvertToFloat64(v)); i++ {
			if name, ok := domMeta.Annotations[fmt.Sprintf("block.%d.name", i)]; ok {
				if path, ok := domMeta.Annotations[fmt.Sprintf("block.%d.path", i)]; ok {
					if v, ok := domMeta.Annotations[fmt.Sprintf("block.%d.rd.reqs", i)]; ok {
						prometheus.MustNewConstMetric(
							e.LibvirtDomainInfoBlockRdReqs,
							prometheus.GaugeValue,
							ConvertToFloat64(v),
							[]string{domain, uuid, name, path}...,
						)
					}

					if v, ok := domMeta.Annotations[fmt.Sprintf("block.%d.rd.bytes", i)]; ok {
						prometheus.MustNewConstMetric(
							e.LibvirtDomainInfoBlockRdBytes,
							prometheus.GaugeValue,
							ConvertToFloat64(v),
							[]string{domain, uuid, name, path}...,
						)
					}

					if v, ok := domMeta.Annotations[fmt.Sprintf("block.%d.rd.times", i)]; ok {
						prometheus.MustNewConstMetric(
							e.LibvirtDomainInfoBlockRdTimes,
							prometheus.GaugeValue,
							ConvertToFloat64(v),
							[]string{domain, uuid, name, path}...,
						)
					}

					if v, ok := domMeta.Annotations[fmt.Sprintf("block.%d.wr.reqs", i)]; ok {
						prometheus.MustNewConstMetric(
							e.LibvirtDomainInfoBlockWrReqs,
							prometheus.GaugeValue,
							ConvertToFloat64(v),
							[]string{domain, uuid, name, path}...,
						)
					}

					if v, ok := domMeta.Annotations[fmt.Sprintf("block.%d.wr.bytes", i)]; ok {
						prometheus.MustNewConstMetric(
							e.LibvirtDomainInfoBlockWrReqs,
							prometheus.GaugeValue,
							ConvertToFloat64(v),
							[]string{domain, uuid, name, path}...,
						)
					}

					if v, ok := domMeta.Annotations[fmt.Sprintf("block.%d.wr.times", i)]; ok {
						prometheus.MustNewConstMetric(
							e.LibvirtDomainInfoBlockWrTimes,
							prometheus.GaugeValue,
							ConvertToFloat64(v),
							[]string{domain, uuid, name, path}...,
						)
					}

					if v, ok := domMeta.Annotations[fmt.Sprintf("block.%d.fl.reqs", i)]; ok {
						prometheus.MustNewConstMetric(
							e.LibvirtDomainInfoBlockFlReqs,
							prometheus.GaugeValue,
							ConvertToFloat64(v),
							[]string{domain, uuid, name, path}...,
						)
					}

					if v, ok := domMeta.Annotations[fmt.Sprintf("block.%d.fl.times", i)]; ok {
						prometheus.MustNewConstMetric(
							e.LibvirtDomainInfoBlockFlTimes,
							prometheus.GaugeValue,
							ConvertToFloat64(v),
							[]string{domain, uuid, name, path}...,
						)
					}
				}
			}
		}
	}

}

func ConvertToFloat64(text string) float64 {
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0
	}
	return value
}

func NewLibvirtExporter() *LibvirtExporter {
	domainLabels := []string{"domain", "uuid"}
	poolLabels := []string{"pool"}

	return &LibvirtExporter{
		// libvirt
		LibvirtUpDesc: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "", "up"),
			"Whether scraping libvirt's metrics was successful.",
			nil,
			nil),

		// domain
		LibvirtDomainInfo: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "up"),
			"Whether scraping domain metrics was successful.",
			domainLabels,
			nil,
		),

		// memory
		LibvirtDomainInfoMaxMem: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "maximum_memory_bytes"),
			"Maximum allowed memory of the domain, in bytes.",
			domainLabels,
			nil),
		LibvirtDomainInfoCurrMem: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "current_memory_bytes"),
			"Current allowed memory of the domain, in bytes.",
			domainLabels,
			nil),

		// cpu
		LibvirtDomainInfoCpuTime: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "cpu_time_ns_total"),
			"all vcpu used time, in ns.",
			domainLabels,
			nil),
		LibvirtDomainInfoCpuUser: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "cpu_user_time_ns_total"),
			"all vcpu used user time, in ns.",
			domainLabels,
			nil),
		LibvirtDomainInfoCpuSystem: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "cpu_system_time_ns_total"),
			"all vcpu used system time, in ns.",
			domainLabels,
			nil,
		),
		LibvirtDomainInfoVcpuCurrent: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "vcpu_current_count"),
			"vcpu current count",
			domainLabels,
			nil,
		),
		LibvirtDomainInfoVcpuMaximum: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "vcpu_max_count"),
			"vcpu max count",
			domainLabels,
			nil,
		),
		LibvirtDomainInfoVcpustate: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "vcpu_state"),
			"vcpu state",
			append(domainLabels, "count"),
			nil,
		),
		LibvirtDomainInfoVcputime: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "vcpu_time"),
			"vcpu time",
			append(domainLabels, "count"),
			nil,
		),

		// net
		LibvirtDomainInfoNetCount: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "net_nic_count"),
			"net nic count",
			domainLabels,
			nil,
		),
		LibvirtDomainInfoNetRxBytes: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "net_nic_rx_bytes"),
			"net nic rx bytes",
			append(domainLabels, "name"),
			nil,
		),
		LibvirtDomainInfoNetRxPkts: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "net_nic_rx_pkts"),
			"net nic rx pkts",
			append(domainLabels, "name"),
			nil,
		),
		LibvirtDomainInfoNetRxErrs: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "net_nic_rx_errs"),
			"net nic rx errs",
			append(domainLabels, "name"),
			nil,
		),
		LibvirtDomainInfoNetRxDrop: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "net_nic_rx_drop"),
			"net nic rx drop",
			append(domainLabels, "name"),
			nil,
		),
		LibvirtDomainInfoNetTxBytes: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "net_nic_tx_bytes"),
			"net nic tx bytes",
			append(domainLabels, "name"),
			nil,
		),
		LibvirtDomainInfoNetTxPkts: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "net_nic_tx_pkts"),
			"net nic tx pkts",
			append(domainLabels, "name"),
			nil,
		),
		LibvirtDomainInfoNetTxErrs: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "net_nic_tx_errs"),
			"net nic rx errs",
			append(domainLabels, "name"),
			nil,
		),
		LibvirtDomainInfoNetTxDrop: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "net_nic_tx_drop"),
			"net nic tx drop",
			append(domainLabels, "name"),
			nil,
		),

		// disk
		LibvirtDomainInfoBlockCount: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "blk_count"),
			"block count",
			domainLabels,
			nil,
		),
		LibvirtDomainInfoBlockPath: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "blk_path"),
			"block path",
			append(domainLabels, "name", "path"),
			nil,
		),
		LibvirtDomainInfoBlockRdReqs: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "blk_rd_requests"),
			"block read requests",
			append(domainLabels, "name", "path"),
			nil,
		),
		LibvirtDomainInfoBlockRdBytes: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "blk_rd_bytes"),
			"block read bytes",
			append(domainLabels, "name", "path"),
			nil,
		),
		LibvirtDomainInfoBlockRdTimes: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "blk_rd_times"),
			"block read times",
			append(domainLabels, "name", "path"),
			nil,
		),
		LibvirtDomainInfoBlockWrReqs: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "blk_wr_requests"),
			"block write requests",
			append(domainLabels, "name", "path"),
			nil,
		),
		LibvirtDomainInfoBlockWrBytes: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "blk_wr_bytes"),
			"block write bytes",
			append(domainLabels, "name", "path"),
			nil,
		),
		LibvirtDomainInfoBlockWrTimes: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "blk_wr_times"),
			"block write times",
			append(domainLabels, "name", "path"),
			nil,
		),
		LibvirtDomainInfoBlockFlReqs: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "blk_flush_requests"),
			"block flush requests",
			append(domainLabels, "name", "path"),
			nil,
		),
		LibvirtDomainInfoBlockFlTimes: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "domain_info", "blk_flush_times"),
			"block flush times",
			append(domainLabels, "name", "path"),
			nil,
		),
		LibvirtPoolInfo: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "pool_info", "pool_count"),
			"pool count",
			nil,
			nil,
		),
		LibvirtPoolInfoStatus: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "pool_info", "pool_status"),
			"pool_status",
			poolLabels,
			nil,
		),
		LibvirtPoolInfoCapacity: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "pool_info", "pool_capacity"),
			"pool capacity",
			poolLabels,
			nil,
		),
		LibvirtPoolInfoAllocated: prometheus.NewDesc(
			prometheus.BuildFQName("libvirt", "pool_info", "pool_allocated"),
			"pool allocated",
			poolLabels,
			nil,
		),
	}
}
