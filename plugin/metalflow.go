package main

import (
	"github.com/hashicorp/go-plugin"

	"github.com/pipego/plugin-fetch/proto"
)

type MetalFlow struct{}

func (n *MetalFlow) Fetch(host string) proto.Result {
	var result proto.Result

	// TODO

	return result
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
