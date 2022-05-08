package proto

import (
	"math"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

const (
	// ResourceCPU CPU, in cores. (500m = .5 cores)
	ResourceCPU = "cpu"
	// ResourceMemory Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	ResourceMemory = "memory"
	// ResourceStorage Volume size, in bytes (e,g. 5Gi = 5GiB = 5 * 1024 * 1024 * 1024)
	ResourceStorage = "storage"
)

const (
	// DefaultMilliCPURequest defines default milli cpu request number.
	DefaultMilliCPURequest int64 = 100 // 0.1 core
	// DefaultMemoryRequest defines default memory request size.
	DefaultMemoryRequest int64 = 200 * 1024 * 1024 // 200 MB
)

// Resources to consider when scoring.
// The default resource set includes "cpu" and "memory" with an equal weight.
const (
	// DefaultCPUWeight defines default cpu weight (allowed weights go from 1 to 100)
	DefaultCPUWeight int64 = 1
	// DefaultMemoryWeight defines default memory weight (allowed weights go from 1 to 100)
	DefaultMemoryWeight int64 = 1
	// DefaultStorageWeight defines default storage weight (allowed weights go from 1 to 100)
	DefaultStorageWeight int64 = 1
)

const (
	// MaxNodeScore is the maximum score a Score plugin is expected to return.
	MaxNodeScore int64 = 100

	// MinNodeScore is the minimum score a Score plugin is expected to return.
	MinNodeScore int64 = 0

	// MaxTotalScore is the maximum total score.
	MaxTotalScore int64 = math.MaxInt64
)

type Args struct {
	Node Node
	Task Task
}

type Node struct {
	AllocatableResource Resource `json:"allocatableResource"`
	Host                string   `json:"host"`
	Label               Label    `json:"label"`
	Name                string   `json:"name"`
	RequestedResource   Resource `json:"requestedResource"`
	Unschedulable       bool     `json:"unschedulable"`
}

type Task struct {
	Name                   string   `json:"name"`
	NodeName               string   `json:"nodeName"`
	NodeSelector           Selector `json:"nodeSelector"`
	RequestedResource      Resource `json:"requestedResource"`
	ToleratesUnschedulable bool     `json:"toleratesUnschedulable"`
}

type Label map[string]string

type Resource struct {
	MilliCPU int64 `json:"milliCPU"`
	Memory   int64 `json:"memory"`
	Storage  int64 `json:"storage"`
}

type Selector map[string][]string

type Fetch interface {
	Fetch(string) Result
}

type Result struct {
	AllocatableResource Resource
	RequestedResource   Resource
}

type FetchRPC struct {
	client *rpc.Client
}

func (n *FetchRPC) Fetch(host string) Result {
	var resp Result
	if err := n.client.Call("Plugin.Fetch", host, &resp); err != nil {
		panic(err)
	}
	return resp
}

type FetchRPCServer struct {
	Impl Fetch
}

func (n *FetchRPCServer) Fetch(host string, resp *Result) error {
	*resp = n.Impl.Fetch(host)
	return nil
}

type FetchPlugin struct {
	Impl Fetch
}

func (n *FetchPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &FetchRPCServer{Impl: n.Impl}, nil
}

func (FetchPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &FetchRPC{client: c}, nil
}
