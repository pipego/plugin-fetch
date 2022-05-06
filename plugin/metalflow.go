package main

import (
	"github.com/hashicorp/go-plugin"

	"github.com/pipego/plugin-fetch/proto"
)

const (
	ErrReasonMetalFlow = "MetalFlow: node(s) didn't return node info"
)

type MetalFlow struct{}

func (n *MetalFlow) Fetch(args *proto.Args) proto.Status {
	var status proto.Status

	// TODO
	status.Error = ErrReasonMetalFlow

	return status
}

// nolint:typecheck
func main() {
	config := plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "plugin-fetch",
		MagicCookieValue: "plugin-fetch",
	}

	pluginMap := map[string]plugin.Plugin{
		"MetalFlow": &proto.FetchPlugin{Impl: &MetalFlow{}},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: config,
		Plugins:         pluginMap,
	})
}
