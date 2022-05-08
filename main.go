package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/pipego/plugin-fetch/proto"
)

type config struct {
	name string
	path string
}

var (
	configs = []config{
		// Plugin: LocalHost
		{
			name: "LocalHost",
			path: "./plugin/fetch-localhost",
		},
		// Plugin: MetalFlow
		{
			name: "MetalFlow",
			path: "./plugin/fetch-metalflow",
		},
	}
)

func main() {
	for _, item := range configs {
		result, _ := helper(item.path, item.name)

		fmt.Println(result.AllocatableResource.MilliCPU, result.AllocatableResource.Memory, result.AllocatableResource.Storage)
		fmt.Println(result.RequestedResource.MilliCPU, result.RequestedResource.Memory, result.RequestedResource.Storage)
	}
}

func helper(path, name string) (proto.Result, error) {
	config := plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "plugin-fetch",
		MagicCookieValue: "plugin-fetch",
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin-fetch",
		Output: os.Stderr,
		Level:  hclog.Error,
	})

	plugins := map[string]plugin.Plugin{
		name: &proto.FetchPlugin{},
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		Cmd:             exec.Command(path),
		HandshakeConfig: config,
		Logger:          logger,
		Plugins:         plugins,
	})
	defer client.Kill()

	rpcClient, _ := client.Client()
	raw, _ := rpcClient.Dispense(name)
	n := raw.(proto.Fetch)
	result := n.Fetch("127.0.0.1")

	return result, nil
}
