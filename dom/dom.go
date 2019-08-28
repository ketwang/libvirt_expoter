package dom

import (
	"io"
	"bufio"
	"errors"
	"strings"

	"gopkg.in/yaml.v2"
	"libvirt_exporter/cmdutil"
)

var (
	ErrDomNotFound     = errors.New("dom instance not found")
	ErrDomStateUnclear = errors.New("dom instance state is unclear")
)

const (
	STATE_RUNNING = "running"
	STATE_SHUTOFF = "shut off"
	STATE_PAUSED  = "paused"
	STATE_OTHER   = "other"
)

type DomMeta struct {
	Name        string `yaml:"Name"`
	UUID        string `yaml:"UUID"`
	Mem         string `yaml:"Max memory"`
	CPU         string `yaml:"CPU(s)"`
	State       string `yaml:"State"`
	Annotations map[string]string
}

type Dom interface {
	String() string
	GetOverallState() (*DomMeta, error)
	GetStats() (string, error)
	GetDoms() ([]Dom, error)
	GetDomByUUID(uuid string) (Dom, error)
}

type DomObj struct {
	Domain string
}

var _ Dom = &DomObj{}

func (d *DomObj) String() string {
	return d.Domain
}

func (d *DomObj) GetOverallState() (*DomMeta, error) {
	args := []string{"dominfo", d.Domain}
	output, errput, err := cmdutil.Command("virsh", args)
	if err != nil {
		return nil, err
	}

	if len(errput) != 0 {
		return nil, errors.New(string(errput))
	}

	dm := &DomMeta{
		Annotations: make(map[string]string, 0),
	}

	// rm line `id: -`
	strOutput := strings.Join(strings.Split(string(output), "\n")[1:], "\n")

	if err := yaml.Unmarshal([]byte(strOutput), dm); err != nil {
		return nil, err
	}

	args = []string{"domstats", d.Domain}
	output, errput, err = cmdutil.Command("virsh", args)
	if err != nil {
		return nil, err
	}

	if len(errput) != 0 {
		return nil, errors.New(string(errput))
	}

	items := strings.Fields(string(output))
	for _, item := range items {
		kv := strings.Split(item, "=")
		if len(kv) == 2 {
			dm.Annotations[kv[0]] = kv[1]
		}
	}

	return dm, nil
}

func (d *DomObj) GetStats() (string, error) {
	args := []string{"domstate", d.Domain}

	output, errput, err := cmdutil.Command("virsh", args)
	if err != nil {
		return STATE_OTHER, err
	}

	if len(errput) != 0 {
		return STATE_OTHER, errors.New(string(errput))
	}

	outputStr := strings.TrimSpace(string(output))

	switch outputStr {
	case STATE_RUNNING, STATE_SHUTOFF, STATE_PAUSED:
		return outputStr, nil
	default:
		return STATE_OTHER, ErrDomStateUnclear
	}
}

func (d *DomObj) GetDoms() ([]Dom, error) {
	args := []string{"list", "--all", "--uuid"}
	outout, errput, err := cmdutil.Command("virsh", args)
	if err != nil {
		return nil, err
	}

	if len(errput) != 0 {
		return nil, errors.New(string(errput))
	}

	doms := make([]Dom, 0)
	lineReader := bufio.NewReader(strings.NewReader(string(outout)))
	for {
		line, _, err := lineReader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		lineStr := strings.TrimSpace(string(line))
		if lineStr != "" {
			dom := &DomObj{
				Domain: lineStr,
			}
			doms = append(doms, dom)
		}
	}

	return doms, nil
}

func (d *DomObj) GetDomByUUID(uuid string) (Dom, error) {
	doms, err := d.GetDoms()
	if err != nil {
		return nil, err
	}

	for _, dom := range doms {
		if uuid == dom.String() {
			return dom, nil
		}
	}

	return nil, ErrDomNotFound
}
