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
	args *proto.Args
	name string
	path string
}

var (
	configs = []config{
		// Plugin: MetalFlow
		{
			args: &proto.Args{},
			name: "MetalFlow",
			path: "./plugin/fetch-metalflow",
		},
	}
)

func main() {
	for _, item := range configs {
		status, _ := helper(item.path, item.name, item.args)
		if status.Error == "" {
			fmt.Println(item.name + ": pass")
		} else {
			fmt.Println(status.Error)
		}
	}
}

func helper(path, name string, args *proto.Args) (proto.Status, error) {
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
	status := n.Fetch(args)

	return status, nil
}
