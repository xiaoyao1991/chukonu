package main

import (
	"flag"
	"fmt"
	"os/exec"
	"plugin"
	"time"

	"github.com/google/cadvisor/client"
	"github.com/google/cadvisor/info/v1"
	"github.com/hashicorp/consul/api"
	"github.com/xiaoyao1991/chukonu/core"

	"net/http"
	_ "net/http/pprof"
)

const CriticalMemThreshold = 0.8
const CriticalCpuThreshold = 0.95

type LifeCycle struct {
	cid      string
	consul   *api.Client
	cadvisor *client.Client
}

func NewLifeCycle(cadvisorBaseUrl string) LifeCycle {
	// Get a new client
	consul, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		//panic(err)
	}
	cadvisor, err := client.NewClient(cadvisorBaseUrl)
	if err != nil {
		panic(err)
	}

	cmd := "cat /proc/self/cgroup | grep 'cpu:/' | sed 's/\\([0-9]\\):cpu:\\/docker\\///g'"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		panic(err)
	}
	cid := string(out)

	return LifeCycle{
		cid:      cid,
		consul:   consul,
		cadvisor: cadvisor,
	}
}

func (d LifeCycle) Register() error {
	agent := d.consul.Agent()
	service := &api.AgentServiceRegistration{
		ID:                d.cid,
		Name:              "chukonu",
		Tags:              nil,
		Port:              7426,
		Address:           d.cid,
		EnableTagOverride: false,
		Check:             &api.AgentServiceCheck{DockerContainerID: d.cid},
		Checks:            nil,
	}
	err := agent.ServiceRegister(service)
	if err != nil {
		return err
	}

	return nil
}

func (d LifeCycle) Run(testplanName string) {
	p, _ := plugin.Open(testplanName)
	sym, _ := p.Lookup("TestPlan")
	requestProvider := sym.(core.RequestProvider)

	config := core.ChukonuConfig{Concurrency: 1000, RequestTimeout: 5 * time.Minute}
	var engines []core.Engine = make([]core.Engine, config.Concurrency)
	for i := 0; i < config.Concurrency; i++ {
		engines[i] = requestProvider.UseEngine()(config)
	}
	var pool core.Pool

	fuse := make(chan bool, 1)
	ack := make(chan bool, 1)
	fuse <- true
	go func(fuse chan bool, ack chan bool) {
		for b := range ack {
			if !b {
				return
			}
			request := v1.ContainerInfoRequest{NumStats: 2}
			sInfo, err := d.cadvisor.ContainerInfo(fmt.Sprintf("/docker/%s", d.cid), &request)
			if err != nil {
				fmt.Println(err)
				// TODO:
			}

			if len(sInfo.Stats) != 2 {
				fuse <- true
			} else {
				// cpu limit
				cpuLimit := float64(sInfo.Spec.Cpu.Quota) / float64(sInfo.Spec.Cpu.Period)
				memoryLimit := float64(sInfo.Spec.Memory.Limit)

				currStat := sInfo.Stats[1]
				prevStat := sInfo.Stats[0]

				// Cpu
				intervalNs := currStat.Timestamp.UnixNano() - prevStat.Timestamp.UnixNano()
				deltaCpuTotalUsage := currStat.Cpu.Usage.Total - prevStat.Cpu.Usage.Total
				cpuUsagePercent := float64(deltaCpuTotalUsage) / float64(intervalNs) / cpuLimit

				// memory
				memUsagePercent := float64(currStat.Memory.Usage) / memoryLimit

				fmt.Printf("\tCPU Usage: %f\n\tMem Usage: %f\n", cpuUsagePercent, memUsagePercent)
				if cpuUsagePercent <= CriticalCpuThreshold && memUsagePercent <= CriticalMemThreshold {
					fuse <- true
				} else {
					break
				}
			}
		}
		close(fuse)
	}(fuse, ack)

	pool.Start(engines, requestProvider, requestProvider.MetricsManager(), config, fuse, ack)
}

var cadvisorBaseUrl = flag.String("cadvisor", "http://localhost:8080/", "base url for cadvisor")

func init() {
	flag.Parse()
}

func main() {
	// TODO: remove when production
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
	l := NewLifeCycle(*cadvisorBaseUrl)
	l.Run("druidtestplan.so")
}
