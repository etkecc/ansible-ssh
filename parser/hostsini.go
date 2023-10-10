package parser

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strings"

	"golang.org/x/exp/slices"
)

const defaultGroup = "ungrouped"

// HostsIni contains all hosts file content
type HostsIni struct {
	cacheGroupVars map[string]*Host    // cached calculated group vars
	cacheGroups    map[string][]string // cached calculated group tree

	Groups    map[string][]*Host           // host-by-group
	GroupVars map[string]map[string]string // group vars
	GroupTree map[string][]string          // raw group tree
	Hosts     map[string]*Host             // hosts-by-name
}

// Host is a parsed host
type Host struct {
	Group      string   // main group
	Groups     []string // all related groups
	Name       string   // host name
	Host       string   // host address
	Port       int      // host port
	User       string   // host user
	SSHPass    string   // host ssh password
	BecomePass string   // host become password
	PrivateKey string   // host ssh private key
}

func NewHostsFile(f string, defaults *Host) (*HostsIni, error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return &HostsIni{}, err
	}

	return NewHostsParser(bs, defaults), nil
}

func NewHostsParser(input []byte, defaults *Host) *HostsIni {
	hosts := &HostsIni{}
	hosts.init()
	hosts.parse(input, defaults)
	return hosts
}

func (h *HostsIni) init() {
	h.Groups = make(map[string][]*Host)
	h.Groups[defaultGroup] = make([]*Host, 0)

	h.GroupTree = make(map[string][]string)
	h.GroupTree[defaultGroup] = make([]string, 0)

	h.GroupVars = make(map[string]map[string]string)
	h.GroupVars[defaultGroup] = make(map[string]string)

	h.Hosts = make(map[string]*Host)

	h.cacheGroupVars = make(map[string]*Host)
	h.cacheGroups = make(map[string][]string)
}

// findAllGroups tries to find all groups related to the group. Experimental
func (h *HostsIni) findAllGroups(groups []string) []string {
	cachekey := strings.Join(groups, ",")
	cached := h.cacheGroups[cachekey]
	if cached != nil {
		return cached
	}

	all := groups
	for name, children := range h.GroupTree {
		if slices.Contains(groups, name) {
			all = append(all, name)
			all = append(all, children...)
			continue
		}
		for _, child := range children {
			if slices.Contains(groups, child) {
				all = append(all, name)
				break
			}
		}
	}
	all = Uniq(all)
	if strings.Join(all, ",") != cachekey {
		all = h.findAllGroups(all)
	}
	h.cacheGroups[cachekey] = all

	return all
}

func (h *HostsIni) initGroup(name string) {
	if _, ok := h.Groups[name]; !ok {
		h.Groups[name] = make([]*Host, 0)
	}
	if _, ok := h.GroupTree[name]; !ok {
		h.GroupTree[name] = make([]string, 0)
	}
	if _, ok := h.GroupVars[name]; !ok {
		h.GroupVars[name] = make(map[string]string)
	}
}

func (h *HostsIni) parse(input []byte, defaults *Host) {
	activeGroupName := defaultGroup
	buff := bytes.NewBuffer(input)
	scanner := bufio.NewScanner(buff)
	for scanner.Scan() {
		line := scanner.Text()
		switch parseType(line) {
		case TypeGroup:
			activeGroupName = parseGroup(line)
			h.initGroup(activeGroupName)
		case TypeGroupVars:
			activeGroupName = parseGroup(line)
			h.initGroup(activeGroupName)
		case TypeGroupChildren:
			activeGroupName = parseGroup(line)
			h.initGroup(activeGroupName)
		case TypeGroupChild:
			group := parseGroup(line)
			h.initGroup(group)
			h.GroupTree[activeGroupName] = append(h.GroupTree[activeGroupName], group)
		case TypeHost:
			host := parseHost(line)
			if host != nil {
				host.Group = activeGroupName
				host.Groups = []string{activeGroupName}
				h.Hosts[host.Name] = host
			}
		case TypeVar:
			k, v := parseVar(line)
			h.GroupVars[activeGroupName][k] = v
		}
	}
	h.finalize(defaults)
}

// groupParams converts group vars map[key]value into []string{"key=value"}
func (h *HostsIni) groupParams(group string) []string {
	vars := h.GroupVars[group]
	if len(vars) == 0 {
		return nil
	}

	params := make([]string, 0, len(vars))
	for k, v := range vars {
		params = append(params, k+"="+v)
	}
	return params
}

// getGroupVars returns merged group vars. Experimental
func (h *HostsIni) getGroupVars(groups []string) *Host {
	cachekey := strings.Join(Uniq(groups), ",")
	cached := h.cacheGroupVars[cachekey]
	if cached != nil {
		return cached
	}

	vars := &Host{}
	for _, group := range groups {
		groupVars := parseParams(h.groupParams(group))
		if groupVars == nil {
			continue
		}
		vars = mergeHost(vars, parseParams(h.groupParams(group)))
	}

	h.cacheGroupVars[cachekey] = vars
	return vars
}

func (h *HostsIni) finalize(defaults *Host) {
	for _, host := range h.Hosts {
		host.Groups = h.findAllGroups(Uniq(host.Groups))
		host = mergeHost(host, h.getGroupVars(host.Groups))
		host = mergeHost(host, defaults)
		h.Hosts[host.Name] = host

		for _, group := range host.Groups {
			h.Groups[group] = append(h.Groups[group], host)
		}
	}
}

// Match a host by name
func (h *HostsIni) Match(m string) *Host {
	return h.Hosts[m]
}

// Merge does append and replace
func (h *HostsIni) Merge(h2 *HostsIni) {
	if h.Groups == nil {
		h.Groups = make(map[string][]*Host)
	}
	if h.Hosts == nil {
		h.Hosts = make(map[string]*Host)
	}

	for group, hosts := range h2.Groups {
		h.Groups[group] = hosts
	}

	for name, host := range h2.Hosts {
		h.Hosts[name] = host
	}
}
