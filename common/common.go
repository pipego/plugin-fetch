package common

import (
	"net/rpc"

	gop "github.com/hashicorp/go-plugin"

	"github.com/pipego/scheduler/plugin"
)

type Fetch interface {
	Fetch(string) Result
}

type Result struct {
	AllocatableResource plugin.Resource
	RequestedResource   plugin.Resource
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

func (n *FetchPlugin) Server(*gop.MuxBroker) (interface{}, error) {
	return &FetchRPCServer{Impl: n.Impl}, nil
}

func (FetchPlugin) Client(b *gop.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &FetchRPC{client: c}, nil
}
