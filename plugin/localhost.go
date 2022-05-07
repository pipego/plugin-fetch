package main

import (
	"math"
	"strings"

	"github.com/hashicorp/go-plugin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/pipego/plugin-fetch/proto"
)

const (
	DEV  = "/dev/"
	HOME = "/home"
	ROOT = "/"
)

type LocalHost struct{}

func (n *LocalHost) Fetch(host string) proto.Result {
	allocatableMilliCPU, requestedMilliCPU := n.MilliCPU()
	allocatableMemory, requestedMemory := n.Memory()
	allocatableStorage, requestedStorage := n.Storage()

	result := proto.Result{
		AllocatableResource: proto.Resource{
			MilliCPU: allocatableMilliCPU,
			Memory:   allocatableMemory,
			Storage:  allocatableStorage,
		},
		RequestedResource: proto.Resource{
			MilliCPU: requestedMilliCPU,
			Memory:   requestedMemory,
			Storage:  requestedStorage,
		},
	}

	return result
}

func (n *LocalHost) MilliCPU() (alloc int64, request int64) {
	c, err := cpu.Counts(true)
	if err != nil {
		return -1, -1
	}

	c = c * 1000
	if c > math.MaxInt64 {
		return -1, -1
	}

	p, err := cpu.Percent(0, false)
	if err != nil {
		return -1, -1
	}

	used := float64(c) * p[0]
	used = used * 1000
	if used > math.MaxInt64 {
		return -1, -1
	}

	return int64(c), int64(used)
}

func (n *LocalHost) Memory() (alloc int64, request int64) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return -1, -1
	}

	if v.Total > math.MaxInt64 || v.Used > math.MaxInt64 {
		return -1, -1
	}

	return int64(v.Total), int64(v.Used)
}

func (n *LocalHost) Storage() (alloc int64, request int64) {
	helper := func(path string) bool {
		found := false
		p, _ := disk.Partitions(false)
		for _, item := range p {
			if strings.HasPrefix(item.Device, DEV) && item.Mountpoint == path {
				found = true
				break
			}
		}
		return found
	}

	r, err := disk.Usage(ROOT)
	if err != nil {
		return -1, -1
	}

	total := r.Total
	used := r.Used

	if helper(HOME) {
		h, err := disk.Usage(HOME)
		if err != nil {
			return -1, -1
		}
		total = h.Total
		used = h.Used
	}

	if total > math.MaxInt64 || used > math.MaxInt64 {
		return -1, -1
	}

	return int64(total), int64(used)
}

// nolint:typecheck
func main() {
	config := plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "plugin-fetch",
		MagicCookieValue: "plugin-fetch",
	}

	pluginMap := map[string]plugin.Plugin{
		"LocalHost": &proto.FetchPlugin{Impl: &LocalHost{}},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: config,
		Plugins:         pluginMap,
	})
}
