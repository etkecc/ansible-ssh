package parser

import (
	"log"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type LineType int

const (
	TypeIgnore        LineType = iota // Line to ignore
	TypeVar           LineType = iota // Line contains var (key=value pair)
	TypeHost          LineType = iota // Line contains host (name key1=value1 key2=value2 ...)
	TypeGroup         LineType = iota // Line contains group ([group])
	TypeGroupVars     LineType = iota // Line contains group vars ([group:vars])
	TypeGroupChild    LineType = iota // Line contains group child (group_child)
	TypeGroupChildren LineType = iota // Line contains group children ([group_children])
)

var groupReplacer = strings.NewReplacer("[", "", "]", "", ":children", "", ":vars", "")

func parseType(line string) LineType {
	line = strings.TrimSpace(line)
	if line == "" {
		return TypeIgnore
	}
	if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || line == "" {
		return TypeIgnore
	}

	if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
		if strings.Contains(line, ":children") {
			return TypeGroupChildren
		}
		if strings.Contains(line, ":vars") {
			return TypeGroupVars
		}
		return TypeGroup
	}

	words := strings.Fields(line)
	if len(words) <= 3 { // "key=value", "key =value", "key = value"
		if slices.Contains(words, "=") {
			log.Println(words)
			return TypeVar
		}
	}

	if len(words) == 1 {
		return TypeGroupChild
	}

	return TypeHost
}

func parseGroup(line string) string {
	return groupReplacer.Replace(line)
}

func parseVar(line string) (key, value string) {
	parts := strings.Split(line, "=")
	if len(parts) != 2 {
		return "", ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func parseHost(line string) *Host {
	parts := strings.Fields(line)
	hostname := parts[0]
	var port int
	if (strings.Contains(hostname, "[") &&
		strings.Contains(hostname, "]") &&
		strings.Contains(hostname, ":") &&
		(strings.LastIndex(hostname, "]") < strings.LastIndex(hostname, ":"))) ||
		(!strings.Contains(hostname, "]") && strings.Contains(hostname, ":")) {

		splithost := strings.Split(hostname, ":")
		if i, err := strconv.Atoi(splithost[1]); err == nil && i != 0 {
			port = i
		}
		hostname = splithost[0]
	}
	params := parts[1:]

	host := parseParams(params)
	host.Name = hostname
	if host.Port == 0 {
		host.Port = port
	}
	return host
}

func parseParams(params []string) *Host {
	vars := &Host{}
	for _, p := range params {
		parts := strings.Split(p, "=")
		if len(parts) < 2 {
			continue
		}
		switch strings.TrimSpace(parts[0]) {
		case "ansible_host":
			vars.Host = parts[1]
		case "ansible_port", "ansible_ssh_port":
			vars.Port, _ = strconv.Atoi(parts[1]) //nolint:errcheck // should not be a big problem
		case "ansible_user":
			vars.User = parts[1]
		case "ansible_ssh_pass":
			vars.SSHPass = parts[1]
		case "ansible_ssh_private_key_file":
			vars.PrivateKey = parts[1]
		case "ansible_become_password":
			vars.BecomePass = parts[1]
		}
	}

	return vars
}
