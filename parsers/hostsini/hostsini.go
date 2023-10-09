package hostsini

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"github.com/flynn/go-shlex"
)

const defaultGroup = "ungrouped"

type Hosts struct {
	input  *bufio.Reader
	Groups map[string][]*Host
	Hosts  map[string]*Host
}

type Host struct {
	Name       string
	Host       string
	Port       int
	User       string
	SSHPass    string
	BecomePass string
	PrivateKey string
}

func NewFile(f string, defaults Host) (*Hosts, error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return &Hosts{}, err
	}

	return NewParser(bytes.NewReader(bs), defaults), nil
}

func NewParser(r io.Reader, defaults Host) *Hosts {
	input := bufio.NewReader(r)
	hosts := &Hosts{input: input, Hosts: make(map[string]*Host)}
	hosts.parse(defaults)
	return hosts
}

func (h *Hosts) parseGroup(line string) string {
	group := ""
	if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
		return group
	}
	replacer := strings.NewReplacer("[", "", "]", "")
	group = replacer.Replace(line)

	if _, ok := h.Groups[group]; !ok {
		h.Groups[group] = make([]*Host, 0)
	}
	return group
}

func (h *Hosts) shouldIgnore(line string) bool {
	return strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || line == ""
}

func (h *Hosts) parse(defaults Host) {
	activeGroupName := defaultGroup
	scanner := bufio.NewScanner(h.input)
	h.Groups = make(map[string][]*Host)
	h.Groups[activeGroupName] = make([]*Host, 0)
	h.Hosts = make(map[string]*Host)

	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " ")
		if h.shouldIgnore(line) {
			continue
		}
		group := h.parseGroup(line)
		if group != "" {
			activeGroupName = group
			continue
		}

		parts, err := shlex.Split(line)
		if err != nil {
			fmt.Println("couldn't tokenize: ", line)
		}
		host := getHost(parts, defaults)
		h.Groups[activeGroupName] = append(h.Groups[activeGroupName], host)
		h.Hosts[host.Name] = host
	}
}

func (h *Hosts) Match(m string) []*Host {
	matchedHosts := make([]*Host, 0, 5)
	for _, hosts := range h.Groups {
		for _, host := range hosts {
			if m, err := path.Match(m, host.Name); err == nil && m {
				matchedHosts = append(matchedHosts, host)
			}
		}
	}
	return matchedHosts
}

func (h *Hosts) MatchOne(m string) *Host {
	return h.Hosts[m]
}

// Merge does append and replace
func (h *Hosts) Merge(h2 *Hosts) {
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

func getHost(parts []string, defaults Host) *Host {
	hostname := parts[0]
	port := defaults.Port
	if (strings.Contains(hostname, "[") &&
		strings.Contains(hostname, "]") &&
		strings.Contains(hostname, ":") &&
		(strings.LastIndex(hostname, "]") < strings.LastIndex(hostname, ":"))) ||
		(!strings.Contains(hostname, "]") && strings.Contains(hostname, ":")) {

		splithost := strings.Split(hostname, ":")
		if i, err := strconv.Atoi(splithost[1]); err == nil {
			port = i
		}
		hostname = splithost[0]
	}
	params := parts[1:]
	host := &Host{
		Name:       hostname,
		Port:       port,
		User:       defaults.User,
		SSHPass:    defaults.SSHPass,
		BecomePass: defaults.BecomePass,
		PrivateKey: defaults.PrivateKey,
	}
	parseParameters(params, host)
	return host
}

func parseParameters(params []string, host *Host) {
	for _, p := range params {
		parts := strings.Split(p, "=")
		if len(parts) < 2 {
			continue
		}
		switch strings.TrimSpace(parts[0]) {
		case "ansible_host":
			host.Host = parts[1]
		case "ansible_port", "ansible_ssh_port":
			host.Port, _ = strconv.Atoi(parts[1]) //nolint:errcheck // should not be a big problem
		case "ansible_user":
			host.User = parts[1]
		case "ansible_ssh_pass":
			host.SSHPass = parts[1]
		case "ansible_ssh_private_key_file":
			host.PrivateKey = parts[1]
		case "ansible_become_password":
			host.BecomePass = parts[1]
		}
	}
}
