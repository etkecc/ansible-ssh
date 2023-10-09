package ansiblecfg

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"strings"
)

const defaultSection = "unknown"

type AnsibleCfg struct {
	input  *bufio.Reader
	Config map[string]map[string]string
}

func NewFile(f string) (*AnsibleCfg, error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return &AnsibleCfg{Config: make(map[string]map[string]string)}, err
	}

	return NewParser(bytes.NewReader(bs)), nil
}

func NewParser(r io.Reader) *AnsibleCfg {
	input := bufio.NewReader(r)
	cfg := &AnsibleCfg{input: input}
	cfg.parse()
	return cfg
}

func (a *AnsibleCfg) parseSection(line string) string {
	section := ""
	if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
		return section
	}
	replacer := strings.NewReplacer("[", "", "]", "")
	section = replacer.Replace(line)

	if _, ok := a.Config[section]; !ok {
		a.Config[section] = make(map[string]string, 0)
	}
	return section
}

func (a *AnsibleCfg) shouldIgnore(line string) bool {
	return strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || line == ""
}

func (a *AnsibleCfg) parseLine(line string) (key, value string) {
	param := strings.Split(line, "=")
	if len(param) < 2 {
		return "", ""
	}
	return strings.TrimSpace(param[0]), strings.TrimSpace(param[1])
}

func (a *AnsibleCfg) parse() {
	a.Config = make(map[string]map[string]string)

	activeSectionName := defaultSection
	scanner := bufio.NewScanner(a.input)

	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " ")
		if a.shouldIgnore(line) {
			continue
		}
		section := a.parseSection(line)
		if section != "" {
			activeSectionName = section
			continue
		}
		key, value := a.parseLine(line)
		if key == "" {
			continue
		}

		if _, ok := a.Config[activeSectionName]; !ok {
			a.Config[activeSectionName] = map[string]string{}
		}
		a.Config[activeSectionName][key] = value
	}
}
