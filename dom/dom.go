package dom

import (
	"bufio"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
	"libvirt_exporter/cmdutil"
)

const (
	qemuXMLPath = "/var/run/libvirt/qemu/"
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

type Nic struct {
	Interface string
	Type      string
	Source    string
	Model     string
	MAC       string
}

type VCPUCollection struct {
	XMLName xml.Name `xml:"domstatus"`
	PID     string   `xml:"pid,attr"`
	VCPUs   []VCPU   `xml:"vcpus>vcpu"`
}

type VCPU struct {
	XMLName xml.Name `xml:"vcpu"`
	Index   string   `xml:"id,attr"`
	PID     string   `xml:"pid,attr"`
}

type DomMeta struct {
	Name  string `yaml:"Name"`
	UUID  string `yaml:"UUID"`
	Mem   string `yaml:"Max memory"`
	CPU   string `yaml:"CPU(s)"`
	State string `yaml:"State"`
	Nics  []Nic
	VCpus VCPUCollection
}

type Dom interface {
	Name() string
	GetOverallState() (*DomMeta, error)
	GetDoms() ([]Dom, error)
}

func NewDom() (Dom, error) {
	return &DomObj{}, nil
}

type DomObj struct {
	Domain string
}

var _ Dom = &DomObj{}

func (d *DomObj) Name() string {
	return d.Domain
}

func (d *DomObj) GetOverallState() (*DomMeta, error) {
	dm := &DomMeta{
		Nics: make([]Nic, 0),
	}

	args := []string{"dominfo", d.Domain}
	output, errput, err := cmdutil.Command("virsh", args)
	if err != nil {
		return nil, err
	}

	if len(errput) != 0 {
		return nil, errors.New(string(errput))
	}

	// rm line `id: -`
	strOutput := strings.Join(strings.Split(string(output), "\n")[1:], "\n")

	if err := yaml.Unmarshal([]byte(strOutput), dm); err != nil {
		return nil, err
	}

	if dm.State != STATE_RUNNING {
		return dm, nil
	}

	// get nic list
	args = []string{"domiflist", d.Domain}
	output, errput, err = cmdutil.Command("virsh", args)
	if err != nil {
		return nil, err
	}

	if len(errput) != 0 {
		return nil, errors.New(string(errput))
	}

	for _, line := range strings.Split(string(output), "\n")[2:] {
		items := strings.Fields(line)
		if len(items) != 5 {
			continue
		}

		dm.Nics = append(dm.Nics, Nic{
			Interface: items[0],
			Type:      items[1],
			Source:    items[2],
			Model:     items[3],
			MAC:       items[4],
		})
	}

	// get vcpu info
	path := qemuXMLPath + dm.Name + ".xml"
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(content, &dm.VCpus); err != nil {
		return nil, err
	}

	return dm, nil
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
