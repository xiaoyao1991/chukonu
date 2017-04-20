package main

import (
	"os"
	"plugin"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/xiaoyao1991/chukonu/core"
)

type LifeCycle struct {
	hostname string
	client   *api.Client
}

func NewLifeCycle() LifeCycle {
	// Get a new client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}
	hostname, _ := os.Hostname()
	return LifeCycle{
		hostname: hostname,
		client:   client,
	}
}

func (d LifeCycle) Register() error {
	agent := d.client.Agent()
	service := &api.AgentServiceRegistration{
		ID:                d.hostname,
		Name:              "chukonu",
		Tags:              nil,
		Port:              7426,
		Address:           d.hostname,
		EnableTagOverride: false,
		Check:             &api.AgentServiceCheck{DockerContainerID: d.hostname},
		Checks:            nil,
	}
	err := agent.ServiceRegister(service)
	if err != nil {
		return err
	}

	return nil
}

func Run(testplanName string) {
	p, _ := plugin.Open(testplanName)             // open the plugin
	sym, _ := p.Lookup("TestPlan")                // lookup symbol(var or methods) from the plugin
	requestProvider := sym.(core.RequestProvider) // type cast to Caller interface

	config := core.ChukonuConfig{Concurrency: 10, RequestTimeout: 5 * time.Minute}
	var engines []core.Engine = make([]core.Engine, config.Concurrency)
	for i := 0; i < config.Concurrency; i++ {
		engines[i] = requestProvider.UseEngine()(config)
	}
	var pool core.Pool

	pool.Start(engines, requestProvider, requestProvider.MetricsManager(), config)
}

func main() {
	Run("druidtestplan.so")
}
