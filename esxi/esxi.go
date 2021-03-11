package esxi

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	tool "vmware_exporter/tools"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

// const (
// 	Scheme = "https"
// 	Path   = "/sdk"
// )

//EsxiInfo 连接信息
type EsxiInfo struct {
	IP       string
	Context  context.Context
	Client   *govmomi.Client
	password string
	user     string
	mutex    sync.Mutex
}

// InitEsxiConn 初始化esxi连接
func InitEsxiConn(ip, user, password string) (*EsxiInfo, error) {
	instance := &EsxiInfo{
		IP:       ip,
		user:     user,
		password: password,
	}

	u := &url.URL{
		Scheme: "https",
		Host:   instance.IP,
		Path:   "/sdk",
	}
	// ctx, cancel := context.WithCancel(context.Background())
	ctx := context.Background()
	u.User = url.UserPassword(instance.user, instance.password)
	client, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		return instance, err
	}
	instance.Context = ctx
	instance.Client = client
	return instance, nil
}

//ScrapeHostSystem 获取Cpu监控数据
func (e *EsxiInfo) ScrapeHostSystem(ch chan<- prometheus.Metric) error {
	info := []interface{}{}
	// 开始获取cpu/memory 数据
	e.mutex.Lock()
	manager := view.NewManager(e.Client.Client)
	v, err := manager.CreateContainerView(e.Context, e.Client.Client.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	defer v.Destroy(e.Context)
	if err != nil {
		log.Fatal().Msgf("%s", err)
	}
	var hss []mo.HostSystem
	err = v.Retrieve(e.Context, []string{"HostSystem"}, []string{"summary"}, &hss)
	if err != nil {
		log.Fatal().Msgf("%s", err)
	}
	for _, hs := range hss {
		upTime := float64((hs.Summary.QuickStats.Uptime) / 60 / 60)
		esxiStatus := map[string]interface{}{
			// "server":   hs.Summary.Config.Name,
			"esxi_ip": e.IP,
			// "quota":    "status",
			"name": "status",
			"data": tool.StatusConvert(hs.Summary.OverallStatus),
			"desc": "Gauge metric with ESXI server " +
				"status(1:green,2:gray,3:yellow,4:red,5:connection fail,)",
			"datatype": prometheus.GaugeValue,
		}
		esxiUpTime := map[string]interface{}{
			// "server":   hs.Summary.Config.Name,
			"esxi_ip": e.IP,
			// "quota":    "status",
			"name":     "uptime",
			"data":     upTime,
			"desc":     "Gauge metric with uptime info(单位：h)",
			"datatype": prometheus.GaugeValue,
		}

		cpuMhzTotal := map[string]interface{}{
			// "server":   hs.Summary.Config.Name,
			"esxi_ip":  e.IP,
			"quota":    "cpu",
			"name":     "cpuMhz_total",
			"data":     float64(hs.Summary.Hardware.CpuMhz) * float64(hs.Summary.Hardware.NumCpuCores),
			"desc":     "Gauge metric with cpu info(单位：Mhz)",
			"datatype": prometheus.GaugeValue,
		}

		cpuMhzUsage := map[string]interface{}{
			// "server":   hs.Summary.Config.Name,
			"esxi_ip": e.IP,
			// "quota":    "cpu",
			"name":     "cpuMhz_usage",
			"data":     float64(hs.Summary.QuickStats.OverallCpuUsage),
			"desc":     "Gauge metric with cpu info(单位：Mhz)",
			"datatype": prometheus.GaugeValue,
		}

		memoryTotal := map[string]interface{}{
			// "server":   hs.Summary.Config.Name,
			"esxi_ip": e.IP,
			// "quota":    "mem",
			"name":     "memory_total",
			"data":     float64(hs.Summary.Hardware.MemorySize),
			"desc":     "Gauge metric with memory info(单位：Byte)",
			"datatype": prometheus.GaugeValue,
		}

		memoryUsage := map[string]interface{}{
			// "server":   hs.Summary.Config.Name,
			"esxi_ip": e.IP,
			// "quota":    "mem",
			"name":     "memory_usage",
			"data":     float64(hs.Summary.QuickStats.OverallMemoryUsage) * 1024 * 1024,
			"desc":     "Gauge metric with memory info(单位：Byte)",
			"datatype": prometheus.GaugeValue,
		}

		// 计算CPU使用率
		totalCPU := int64(hs.Summary.Hardware.CpuMhz) * int64(hs.Summary.Hardware.NumCpuCores)
		cpuUsageRateValue := float64(hs.Summary.QuickStats.OverallCpuUsage) / float64(totalCPU)
		log.Info().Msgf("%f", cpuUsageRateValue)
		cpuUsageRate := map[string]interface{}{
			// "server":   hs.Summary.Config.Name,
			"esxi_ip": e.IP,
			// "quota":    "mem",
			"name":     "cpu_usage_rate",
			"data":     cpuUsageRateValue,
			"desc":     "Gauge metric with cpu rate",
			"datatype": prometheus.GaugeValue,
		}
		// 计算内存使用率
		memoryUsageRateValue := (float64(hs.Summary.QuickStats.OverallMemoryUsage) * 1024 * 1024) / float64(hs.Summary.Hardware.MemorySize)
		log.Info().Msgf("%f", memoryUsageRateValue)
		memoryUsageRate := map[string]interface{}{
			// "server":   hs.Summary.Config.Name,
			"esxi_ip": e.IP,
			// "quota":    "mem",
			"name":     "memory_usage_rate",
			"data":     memoryUsageRateValue,
			"desc":     "Gauge metric with memory rate",
			"datatype": prometheus.GaugeValue,
		}
		info = append(
			info, cpuMhzTotal, cpuMhzUsage,
			memoryTotal, memoryUsage,
			esxiUpTime, esxiStatus,
			memoryUsageRate, cpuUsageRate)
	}
	// 指标获取逻辑，此处不做具体操作，仅仅赋值进⾏示例
	//指标获取逻辑，此处不做具体操作，仅仅赋值进行示例
	//生成采集的指标名
	for _, quota := range info {
		quotaData := quota.(map[string]interface{})["data"].(float64)
		// quotaArgs := quota.(map[string]interface{})["quota"].(string)
		quotaName := quota.(map[string]interface{})["name"].(string)
		quotaDesc := quota.(map[string]interface{})["desc"].(string)
		// quotaServer := quota.(map[string]interface{})["server"].(string)
		
		quotaIP := quota.(map[string]interface{})["esxi_ip"].(string)
		quotaDataType := quota.(map[string]interface{})["datatype"].(prometheus.ValueType)
		// metricName := prometheus.BuildFQName("esxi", "", quotaPro)
		// //该例子具有disk_name的维度，须在[]string{"disk_name"}
		// newDesc := prometheus.NewDesc(metricName, quotaDesc, []string{"cpu"}, nil)

		newDesc := prometheus.NewDesc(
			fmt.Sprintf("esxi_%s", quotaName), //指标名称
			quotaDesc,                         //帮助信息
			[]string{"esxi_ip"},               //多维度
			nil,                               //设置labels 如prometheus.Labels{"sh": "lz"},
		)

		metricInfo := prometheus.MustNewConstMetric(
			newDesc,
			quotaDataType, //prometheus 数据类型
			quotaData,
			quotaIP,
			// quotaServer,
		)
		ch <- metricInfo
	}
	e.mutex.Unlock()
	return nil
}

//ServeConnfail Esxi连接失败
func ServeConnfail(ch chan<- prometheus.Metric, ip string) {
	/*
		ManagedEntityStatusGray   = ManagedEntityStatus("gray")
		ManagedEntityStatusGreen  = ManagedEntityStatus("green")
		ManagedEntityStatusYellow = ManagedEntityStatus("yellow")
		ManagedEntityStatusRed    = ManagedEntityStatus("red")
	*/

	newDesc := prometheus.NewDesc(
		fmt.Sprintf("esxi_%s", "status"), //指标名称
		"Gauge metric with ESXI server status(1:green,2:gray,3:yellow,4:red,5:connection fail,)", //帮助信息
		[]string{"esxi_ip", "server_name"}, //多维度
		nil,                                //设置labels 如prometheus.Labels{"sh": "lz"},
	)

	metricInfo := prometheus.MustNewConstMetric(
		newDesc,
		prometheus.GaugeValue, //prometheus 数据类型
		float64(5),
		ip,
		"",
	)
	ch <- metricInfo
}
