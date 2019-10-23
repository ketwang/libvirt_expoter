package pool

import (
	"encoding/xml"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"strings"

	"libvirt-exporter/cmdutil"
)

const (
	StateRunning = "running"
	StateShutoff = "shut off"
	StatePaused  = "paused"
	StateOther   = "other"
)

type PoolType struct {
	XMLName xml.Name `xml:"pool"`
	Type    string   `xml:"type,attr"`
	Source  struct {
		Meta string `xml:",innerxml"`
	} `xml:"source"`
}

type PoolMeta struct {
	Name       string `yaml:"Name"`
	UUID       string `yaml:"UUID"`
	State      string `yaml:"State"`
	Persistent string `yaml:"Persistent"`
	Autostart  string `yaml:"Autostart"`
	Capacity   string `yaml:"Capacity"`
	Allocation string `yaml:"Allocation"`
	Available  string `yaml:"Available"`
	Type       *PoolType
}

type Pool interface {
	String() string
	GetOverallState() (*PoolMeta, error)
	GetPools() ([]Pool, error)
}

type PoolObj struct {
	Name string
}

var _ Pool = &PoolObj{}

func (p *PoolObj) String() string {
	return p.Name
}

func (p *PoolObj) GetPools() ([]Pool, error) {
	pools := make([]Pool, 0)

	args := []string{"pool-list", "--all", "--name"}

	output, errput, err := cmdutil.Command("virsh", args)
	if err != nil {
		return nil, err
	}

	if len(errput) != 0 {
		return nil, errors.New(string(errput))
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pools = append(pools, &PoolObj{
			Name: line,
		})
	}

	return pools, nil
}

func (p *PoolObj) GetOverallState() (*PoolMeta, error) {
	poolName := p.Name
	args := []string{"pool-info", poolName}

	output, errput, err := cmdutil.Command("virsh", args)
	if err != nil {
		return nil, err
	}

	if len(errput) != 0 {
		return nil, errors.New(string(errput))
	}

	poolMeta := &PoolMeta{}
	if err := yaml.Unmarshal(output, poolMeta); err != nil {
		return nil, err
	}

	args = []string{"pool-dumpxml", poolName}
	output, errput, err = cmdutil.Command("virsh", args)
	if err != nil {
		return nil, fmt.Errorf("exec failed, cmd='%s' err=%s",
			strings.Join(args, " "),
			err.Error())
	}

	if len(errput) != 0 {
		return nil, fmt.Errorf("get pool info failed, pool=%s err=%s", poolName, err.Error())
	}

	type_ := &PoolType{}
	if err := xml.Unmarshal(output, type_); err != nil {
		return nil, err
	}

	poolMeta.Type = type_

	return poolMeta, nil
}
