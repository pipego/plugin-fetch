package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	gop "github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"

	"github.com/pipego/scheduler/common"
	"github.com/pipego/scheduler/plugin"
)

const (
	URL  = "http://127.0.0.1:4523/mock/954718"
	USER = "user"
	PASS = "pass"
)

type MetalFlow struct {
	host  string
	token string
}

func (n *MetalFlow) Run(host string) plugin.FetchResult {
	var err error
	n.host = host

	n.token, err = n.login()
	if err != nil {
		return plugin.FetchResult{}
	}

	allocatable, requested, err := n.node()
	if err != nil {
		return plugin.FetchResult{}
	}

	return plugin.FetchResult{
		AllocatableResource: allocatable,
		RequestedResource:   requested,
	}
}

func (n *MetalFlow) login() (string, error) {
	token, err := n.jwtToken()
	if err != nil {
		return "", errors.Wrap(err, "failed to fetch jwt token")
	}

	return token, nil
}

func (n *MetalFlow) jwtToken() (string, error) {
	buf := map[string]string{
		"username": USER,
		"password": PASS,
	}

	body, err := json.Marshal(buf)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal\n")
	}

	req, err := http.NewRequest(http.MethodPost, URL+"/api/v1/base/login", bytes.NewBuffer(body))
	if err != nil {
		return "", errors.Wrap(err, "failed to request\n")
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to post\n")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("invalid status\n")
	}

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read\n")
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal(ret, &data); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal\n")
	}

	if int(data["code"].(float64)) != 201 {
		return "", errors.New("incorrect code\n")
	}

	return data["result"].(map[string]interface{})["token"].(string), nil
}

func (n *MetalFlow) node() (alloc common.Resource, request common.Resource, err error) {
	req, err := http.NewRequest(http.MethodGet, URL+"/api/v1/node/list?address="+n.host, nil)
	if err != nil {
		return common.Resource{}, common.Resource{}, errors.Wrap(err, "failed to request\n")
	}

	req.Header.Add("Authorization", "Bearer "+n.token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return common.Resource{}, common.Resource{}, errors.Wrap(err, "failed to get\n")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return common.Resource{}, common.Resource{}, errors.Wrap(err, "invalid status\n")
	}

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return common.Resource{}, common.Resource{}, errors.Wrap(err, "failed to read\n")
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal(ret, &data); err != nil {
		return common.Resource{}, common.Resource{}, errors.Wrap(err, "failed to unmarshal\n")
	}

	if int(data["code"].(float64)) != 201 {
		return common.Resource{}, common.Resource{}, errors.Wrap(err, "incorrect code\n")
	}

	list := data["result"].(map[string]interface{})["list"].([]interface{})
	if len(list) != 1 {
		return common.Resource{}, common.Resource{}, errors.New("incorrect result\n")
	}

	buf := list[0].(map[string]interface{})
	info := buf["information"].(map[string]interface{})

	alloc, err = n.allocHelper(info)
	if err != nil {
		return common.Resource{}, common.Resource{}, errors.New("incorrect alloc\n")
	}

	request, err = n.requestHelper(info)
	if err != nil {
		return common.Resource{}, common.Resource{}, errors.New("incorrect request\n")
	}

	return alloc, request, nil
}

func (n *MetalFlow) allocHelper(info map[string]interface{}) (common.Resource, error) {
	// "4 CPU (2.1% Used)"
	cpuHelper := func(data string) int64 {
		buf := strings.Split(data, " ")
		if len(buf) != 4 {
			return -1
		}
		b, _ := strconv.Atoi(buf[0])
		return int64(b * 1000)
	}

	// "4 GB (2 GB Used)"
	memoryHelper := func(data string) int64 {
		buf := strings.Split(data, " ")
		if len(buf) != 5 {
			return -1
		}
		b, _ := strconv.Atoi(buf[0])
		return int64(b * 1024 * 1024 * 1024)
	}

	// "4.0 GB (2.0 GB Used)"
	storageHelper := func(data string) int64 {
		buf := strings.Split(data, " ")
		if len(buf) != 5 {
			return -1
		}
		b, _ := strconv.ParseFloat(buf[0], 64)
		return int64(b * 1024 * 1024 * 1024)
	}

	return common.Resource{
		MilliCPU: cpuHelper(info["cpu"].(string)),
		Memory:   memoryHelper(info["ram"].(string)),
		Storage:  storageHelper(info["disk"].(string)),
	}, nil
}

func (n *MetalFlow) requestHelper(info map[string]interface{}) (common.Resource, error) {
	// "4 CPU (2.1% Used)"
	cpuHelper := func(data string) int64 {
		buf := strings.Split(data, " ")
		if len(buf) != 4 {
			return -1
		}
		c, _ := strconv.Atoi(buf[0])
		b := strings.TrimPrefix(buf[2], "(")
		b = strings.TrimSuffix(b, "%")
		p, _ := strconv.ParseFloat(b, 64)
		return int64(float64(c) * p * 0.01 * 1000)
	}

	// "4 GB (2 GB Used)"
	memoryHelper := func(data string) int64 {
		buf := strings.Split(data, " ")
		if len(buf) != 5 {
			return -1
		}
		m := strings.TrimPrefix(buf[2], "(")
		b, _ := strconv.Atoi(m)
		return int64(b * 1024 * 1024 * 1024)
	}

	// "4.0 GB (2.0 GB Used)"
	storageHelper := func(data string) int64 {
		buf := strings.Split(data, " ")
		if len(buf) != 5 {
			return -1
		}
		s := strings.TrimPrefix(buf[2], "(")
		b, _ := strconv.ParseFloat(s, 64)
		return int64(b * 1024 * 1024 * 1024)
	}

	return common.Resource{
		MilliCPU: cpuHelper(info["cpu"].(string)),
		Memory:   memoryHelper(info["ram"].(string)),
		Storage:  storageHelper(info["disk"].(string)),
	}, nil
}

// nolint:typecheck
func main() {
	config := gop.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "plugin",
		MagicCookieValue: "plugin",
	}

	pluginMap := map[string]gop.Plugin{
		"MetalFlow": &plugin.Fetch{Impl: &MetalFlow{}},
	}

	gop.Serve(&gop.ServeConfig{
		HandshakeConfig: config,
		Plugins:         pluginMap,
	})
}
