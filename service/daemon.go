package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	cadvisor "github.com/google/cadvisor/client"
	"github.com/hashicorp/consul/api"
	"github.com/xiaoyao1991/chukonu/core"
	"golang.org/x/net/context"
)

const ChukonuImage = "chukonu"
const BytesInKB = 1024
const BytesInMB = 1024 * 1024
const BytesInGB = 1024 * 1024 * 1024
const numAcceptors = 100

type Daemon struct {
	hostname        string
	consulAddr      string
	consul          *api.Client
	cadvisorBaseUrl string
	cadvisor        *cadvisor.Client
	docker          *docker.Client
}

func NewDaemon(cadvisorBaseUrl string, consulAddr string) Daemon {
	consul, err := api.NewClient(&api.Config{Address: consulAddr})
	if err != nil {
		panic(err)
	}
	cadvisor, err := cadvisor.NewClient(cadvisorBaseUrl)
	if err != nil {
		panic(err)
	}
	docker, err := docker.NewEnvClient()
	if err != nil {
		panic(err)
	}

	hostname, _ := os.Hostname()

	return Daemon{
		hostname:        hostname,
		consulAddr:      consulAddr,
		consul:          consul,
		cadvisorBaseUrl: cadvisorBaseUrl,
		cadvisor:        cadvisor,
		docker:          docker,
	}
}

func (d Daemon) SetupTestPlan(config core.ChukonuConfig) error {
	kv := d.consul.KV()

	concurrencyB := make([]byte, 8)
	iterationsB := make([]byte, 8)
	totalTimeoutB := make([]byte, 8)

	binary.LittleEndian.PutUint64(concurrencyB, uint64(config.Concurrency))
	binary.LittleEndian.PutUint64(iterationsB, uint64(config.Iterations))
	binary.LittleEndian.PutUint64(totalTimeoutB, uint64(config.TotalTimeout.Nanoseconds()))

	kvpair := &api.KVPair{
		Key:   fmt.Sprintf("%s/config/concurrency", config.TenantId),
		Value: concurrencyB,
	}
	_, err := kv.Put(kvpair, nil)
	if err != nil {
		return err
	}

	kvpair = &api.KVPair{
		Key:   fmt.Sprintf("%s/config/iterations", config.TenantId),
		Value: iterationsB,
	}
	_, err = kv.Put(kvpair, nil)
	if err != nil {
		return err
	}

	kvpair = &api.KVPair{
		Key:   fmt.Sprintf("%s/config/timeout", config.TenantId),
		Value: totalTimeoutB,
	}
	_, err = kv.Put(kvpair, nil)
	if err != nil {
		return err
	}

	return nil
}

func (d Daemon) SpawnNewContainer(tenantId string) string {
	ctx := context.Background()

	// TODO: lock consul KV table

	resp, err := d.docker.ContainerCreate(ctx, &container.Config{
		Image: ChukonuImage,
		Cmd:   []string{"-tenant", tenantId, "-cadvisor", d.cadvisorBaseUrl, "-consul", d.consulAddr},
	}, &container.HostConfig{
		Resources: container.Resources{
			// TODO: how do we decide what limit? should we smart calculate?
			Memory:     int64(16 * BytesInMB),
			MemorySwap: int64(20 * BytesInMB),
			CPUQuota:   int64(50000),
			CPUPeriod:  int64(100000),
		},
	}, nil, "")
	if err != nil {
		panic(err)
	}

	if err := d.docker.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	// TODO: unlock consul KV table? or should we unlock it in lifecycle

	return resp.ID
}

func (d Daemon) ShouldSpawnNewContainer(tenantId string) bool {
	kv := d.consul.KV()

	kvs, _, err := kv.List(fmt.Sprintf("%s/workercount", tenantId), nil)
	if err != nil {
		panic(err)
	}

	var totalWorkerCount uint64
	for _, kvp := range kvs {
		totalWorkerCount += binary.LittleEndian.Uint64(kvp.Value)
	}

	result, _, err := kv.Get(fmt.Sprintf("%s/config/concurrency", tenantId), nil)
	if err != nil {
		panic(err)
	}
	concurrency := binary.LittleEndian.Uint64(result.Value)

	// TODO: take node resource into consideration
	return totalWorkerCount < concurrency
}

// TODO: aggregate metrics
func (d Daemon) ReportMetrics() {

}

// TODO: spawn new node when the residing node is out of resource
// func (d Daemon) SpawnNewNode() {
//
// }
// func (d Daemon) ShouldSpawnNewNode() bool {
// 	return false
// }

var cadvisorBaseUrl = flag.String("cadvisor", "http://localhost:8080/", "base url for cadvisor")
var consulAddress = flag.String("consul", "http://localhost:8500/", "consul address")

func init() {
	flag.Parse()
}
func main() {
	//TODO: tmp
	daemon := NewDaemon(*cadvisorBaseUrl, *consulAddress)
	daemon.SetupTestPlan(core.ChukonuConfig{TenantId: "druidtest", Concurrency: 1000, RequestTimeout: 5 * time.Minute})
	tick := time.Tick(5 * time.Second)
	for {
		<-tick
		if daemon.ShouldSpawnNewContainer("druidtest") {
			daemon.SpawnNewContainer("druidtest")
		}
	}
}
