package common

import (
	"net/rpc"

	gop "github.com/hashicorp/go-plugin"

	"github.com/pipego/scheduler/plugin"
)

type FetchRPC struct {
	client *rpc.Client
}

func (n *FetchRPC) Run(host string) plugin.FetchResult {
	var resp plugin.FetchResult
	if err := n.client.Call("Plugin.Run", host, &resp); err != nil {
		panic(err)
	}
	return resp
}

type FetchRPCServer struct {
	Impl plugin.FetchPlugin
}

func (n *FetchRPCServer) Run(host string, resp *plugin.FetchResult) error {
	*resp = n.Impl.Run(host)
	return nil
}

type FetchPlugin struct {
	Impl plugin.FetchPlugin
}

func (n *FetchPlugin) Server(*gop.MuxBroker) (interface{}, error) {
	return &FetchRPCServer{Impl: n.Impl}, nil
}

func (FetchPlugin) Client(b *gop.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &FetchRPC{client: c}, nil
}
