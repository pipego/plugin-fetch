package main

import (
	"math"
	"strings"

	gop "github.com/hashicorp/go-plugin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/pipego/scheduler/common"
	"github.com/pipego/scheduler/plugin"
)

const (
	// DURATION Duration: 10s = 10*1000ms = 10*1000000000ns
	DURATION = 10 * 1000000000
)

const (
	DEV  = "/dev/"
	HOME = "/home"
	ROOT = "/"
)

type LocalHost struct{}

func (n *LocalHost) Run(_ string) plugin.FetchResult {
	allocatableMilliCPU, requestedMilliCPU := n.MilliCPU()
	allocatableMemory, requestedMemory := n.Memory()
	allocatableStorage, requestedStorage := n.Storage()

	result := plugin.FetchResult{
		AllocatableResource: common.Resource{
			MilliCPU: allocatableMilliCPU,
			Memory:   allocatableMemory,
			Storage:  allocatableStorage,
		},
		RequestedResource: common.Resource{
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

	if c*1000 > math.MaxInt64 {
		return -1, -1
	}

	// FIXME: Got error on MacOS 10.13.6
	p, err := cpu.Percent(DURATION, false)
	if err != nil {
		return -1, -1
	}

	used := float64(c) * p[0] * 0.01
	if used > math.MaxInt64 {
		return -1, -1
	}

	return int64(c * 1000), int64(used * 1000)
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
	config := gop.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "plugin",
		MagicCookieValue: "plugin",
	}

	pluginMap := map[string]gop.Plugin{
		"LocalHost": &plugin.Fetch{Impl: &LocalHost{}},
	}

	gop.Serve(&gop.ServeConfig{
		HandshakeConfig: config,
		Plugins:         pluginMap,
	})
}
