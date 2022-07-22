package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	gop "github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"

	"github.com/pipego/scheduler/plugin"
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

	handshake = gop.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "plugin",
		MagicCookieValue: "plugin",
	}

	logger = hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stderr,
		Level:  hclog.Error,
	})
)

func main() {
	for _, item := range configs {
		p, _ := filepath.Abs(item.path)
		if result, err := helper(p, item.name); err == nil {
			fmt.Println(result.AllocatableResource.MilliCPU, result.AllocatableResource.Memory, result.AllocatableResource.Storage)
			fmt.Println(result.RequestedResource.MilliCPU, result.RequestedResource.Memory, result.RequestedResource.Storage)
		} else {
			fmt.Println(err.Error())
		}
	}
}

func helper(path, name string) (plugin.FetchResult, error) {
	plugins := map[string]gop.Plugin{
		name: &plugin.Fetch{},
	}

	client := gop.NewClient(&gop.ClientConfig{
		Cmd:             exec.Command(path),
		HandshakeConfig: handshake,
		Logger:          logger,
		Plugins:         plugins,
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		return plugin.FetchResult{}, errors.Wrap(err, "failed to init client")
	}

	raw, err := rpcClient.Dispense(name)
	if err != nil {
		return plugin.FetchResult{}, errors.Wrap(err, "failed to dispense instance")
	}

	n := raw.(plugin.FetchImpl)
	result := n.Run("127.0.0.1")

	return result, nil
}
