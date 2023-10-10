package parser

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strings"
)

const defaultSection = "unknown"

type AnsibleCfg struct {
	Config map[string]map[string]string
}

func NewAnsibleCfgFile(f string) (*AnsibleCfg, error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return &AnsibleCfg{Config: make(map[string]map[string]string)}, err
	}

	return NewAnsibleCfgParser(bs), nil
}

func NewAnsibleCfgParser(input []byte) *AnsibleCfg {
	cfg := &AnsibleCfg{}
	cfg.parse(input)
	return cfg
}

func (a *AnsibleCfg) parse(input []byte) {
	a.Config = make(map[string]map[string]string)

	activeSectionName := defaultSection
	scanner := bufio.NewScanner(bytes.NewBuffer(input))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch parseType(line) {
		case TypeGroup:
			activeSectionName = parseGroup(line)
		case TypeVar:
			key, value := parseVar(line)
			if _, ok := a.Config[activeSectionName]; !ok {
				a.Config[activeSectionName] = map[string]string{}
			}
			a.Config[activeSectionName][key] = value
		}
	}
}
