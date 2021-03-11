package exporter

import (
	"sync"
	"vmware_exporter/esxi"

	"github.com/rs/zerolog/log"

	"github.com/prometheus/client_golang/prometheus"
)

//Exporter 结构体
type Exporter struct {
	error        prometheus.Gauge
	scrapeErrors *prometheus.CounterVec
}

//NewExporter 构体实例化
func NewExporter() *Exporter {
	return &Exporter{}
}

//Describe 用来生成采集指标的描述信息
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	metricCh := make(chan prometheus.Metric)
	doneCh := make(chan struct{})

	go func() {
		for m := range metricCh {
			ch <- m.Desc()
		}
		close(doneCh)
	}()

	e.Collect(metricCh)
	close(metricCh)
	<-doneCh

}

//Collect 方法，采集数据的入口
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	// 初始化当前esxi的实例
	esxiServers := []map[string]string{
		{
			"ip":       "10.40.92.2",
			"user":     "root",
			"password": "shlz@123456",
		},
		{
			"ip":       "10.40.92.238",
			"user":     "root",
			"password": "shlz@123456",
		},
	}
	var wg sync.WaitGroup

	for _, esxiServer := range esxiServers {
		wg.Add(1)
		go func(ch chan<- prometheus.Metric, ip, user, password string) {
			defer wg.Done()
			log.Info().Msgf("start grab %s monitor data", ip)
			esxiInstance, err1 := esxi.InitEsxiConn(ip, user, password)
			if err1 != nil {
				//无法连接监控（可能是配置信息错误、可能是服务器宕机）
				esxi.ServeConnfail(ch, ip)
				log.Warn().Msgf("%s", err1)
				return
			}

			// 获取内存、cpu监控信息
			if err := esxiInstance.ScrapeHostSystem(ch); err != nil {
				log.Warn().Msgf("%s", err)
				e.scrapeErrors.WithLabelValues("localtime").Inc()
			}
			esxiInstance.Client.Logout(esxiInstance.Context)
		}(ch, esxiServer["ip"], esxiServer["user"], esxiServer["password"])

	}
	wg.Wait()
}
