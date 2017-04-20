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
)

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

	// get container id: cat /proc/self/cgroup | grep "cpu:/" | sed 's/\([0-9]\):cpu:\/docker\///g'
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

	config := core.ChukonuConfig{Concurrency: 10, RequestTimeout: 5 * time.Minute}
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
			fmt.Println(b)
			request := v1.ContainerInfoRequest{NumStats: 1}
			sInfo, err := d.cadvisor.ContainerInfo(fmt.Sprintf("/docker/%s", d.cid), &request)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(sInfo)
			fuse <- true
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
	l := NewLifeCycle(*cadvisorBaseUrl)
	l.Run("druidtestplan.so")
}
