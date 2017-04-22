package main

import (
	"encoding/binary"
	"fmt"
	"os"

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

type Daemon struct {
	hostname        string
	consulAddr      string
	consul          *api.Client
	cadvisorBaseUrl string
	cadvisor        *cadvisor.Client
	docker          *docker.Client
}

func NewDaemon(cadvisorBaseUrl string, consulAddr string) Daemon {
	// Get a new client
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
		Key:   fmt.Sprintf("%s/%s/config/concurrency", config.TenantId, d.hostname),
		Value: concurrencyB,
	}
	_, err := kv.Put(kvpair, nil)
	if err != nil {
		return err
	}

	kvpair = &api.KVPair{
		Key:   fmt.Sprintf("%s/%s/config/iterations", config.TenantId, d.hostname),
		Value: iterationsB,
	}
	_, err = kv.Put(kvpair, nil)
	if err != nil {
		return err
	}

	kvpair = &api.KVPair{
		Key:   fmt.Sprintf("%s/%s/config/timeout", config.TenantId, d.hostname),
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

	resp, err := d.docker.ContainerCreate(ctx, &container.Config{
		Image: ChukonuImage,
		Cmd:   []string{"-tenant", tenantId, "-cadvisor", d.cadvisorBaseUrl, "-consul", d.consulAddr},
	}, &container.HostConfig{
		Resources: container.Resources{
			// TODO: how do we decide what limit? should we smart calculate?
			Memory:     int64(128 * BytesInMB),
			MemorySwap: int64(192 * BytesInMB),
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

	return resp.ID
}

func main() {

}
