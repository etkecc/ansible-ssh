package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/adrg/xdg"

	"gitlab.com/etke.cc/int/ansible-ssh/config"
	"gitlab.com/etke.cc/int/ansible-ssh/parsers/ansiblecfg"
	"gitlab.com/etke.cc/int/ansible-ssh/parsers/hostsini"
)

var (
	withDebug bool
	logger    = log.New(os.Stdout, "[ansible.ssh] ", 0)
)

func main() {
	if len(os.Args) < 2 {
		logger.Println("you need to provide at least host name")
		return
	}

	path, err := xdg.SearchConfigFile("ansible-ssh.yml")
	if err != nil {
		logger.Fatal("cannot find the ansible-ssh.yml config file:", err)
	}
	cfg, err := config.Read(path)
	if err != nil {
		logger.Fatal("cannot read the ansible-ssh.yml config file:", err)
	}
	withDebug = cfg.Debug

	acfg, defaults := parseAnsibleConfig(cfg)
	inv := findInventory(cfg.Path, acfg, defaults)
	if inv == nil {
		debug("inventory not found")
		executeSSH(cfg.SSHCommand, nil, cfg.InventoryOnly)
		return
	}
	host := inv.MatchOne(os.Args[1])
	if host == nil {
		debug("host", os.Args[1], "not found in inventory")
		executeSSH(cfg.SSHCommand, nil, cfg.InventoryOnly)
		return
	}
	debug("host", host.Name, "has been found, starting ssh")
	executeSSH(cfg.SSHCommand, host, cfg.InventoryOnly)
}

func parseAnsibleConfig(cfg *config.Config) (*ansiblecfg.AnsibleCfg, *config.Defaults) {
	defaults := &cfg.Defaults
	acfg, err := ansiblecfg.NewFile("./ansible.cfg")
	if err != nil {
		debug("ansible.cfg is not available", err)
		return nil, defaults
	}

	if _, ok := acfg.Config["defaults"]; !ok {
		debug("ansible.cfg doesn't contain [defaults] section")
		return acfg, defaults
	}

	if user := acfg.Config["defaults"]["remote_user"]; user != "" {
		defaults.User = user
	}
	if privkey := acfg.Config["defaults"]["private_key_file"]; privkey != "" {
		defaults.PrivateKey = privkey
	}
	if port := acfg.Config["defaults"]["remote_port"]; port != "" {
		portI, err := strconv.Atoi(port)
		if err == nil {
			defaults.Port = portI
		}
	}

	return acfg, defaults
}

func findInventory(cfgPath string, acfg *ansiblecfg.AnsibleCfg, defaults *config.Defaults) *hostsini.Hosts {
	inv := getInventory(cfgPath, defaults)
	if inv != nil {
		debug("inventory", cfgPath, "is not found")
		return inv
	}

	if acfg == nil {
		return nil
	}

	invcfg := acfg.Config["defaults"]["inventory"]
	if invcfg == "" {
		debug("no inventories found in ansible.cfg")
		return nil
	}
	invpaths := strings.Split(invcfg, ",")
	if len(invpaths) == 0 {
		debug("no inventories found in ansible.cfg")
		return nil
	}

	inv = &hostsini.Hosts{}
	for _, path := range invpaths {
		parsedInv := getInventory(path, defaults)
		if parsedInv == nil {
			debug("inventory", path, "found in ansible.cfg isn't eligible")
			continue
		}
		inv.Merge(parsedInv)
	}

	if len(inv.Hosts) == 0 {
		return nil
	}

	return inv
}

func buildCMD(sshCmd string, host *hostsini.Host, strict bool) *exec.Cmd {
	if host == nil {
		if strict {
			logger.Fatal("host not found within inventory")
		}
		debug("command:", sshCmd, os.Args[1:])
		return exec.Command(sshCmd, os.Args[1:]...)
	}
	debug("command:", sshCmd, buildSSHArgs(host))

	if host.SSHPass != "" {
		logger.Println("ssh password is:", host.SSHPass)
	}

	if host.BecomePass != "" {
		logger.Println("become password is:", host.BecomePass)
	}
	return exec.Command(sshCmd, buildSSHArgs(host)...)
}

func executeSSH(sshCmd string, host *hostsini.Host, strict bool) {
	cmd := buildCMD(sshCmd, host, strict)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Start()
	if err != nil {
		logger.Fatal("cannot start the command:", err)
	}
	err = cmd.Wait()
	if err != nil {
		logger.Fatal("command failed:", err)
	}
}

func buildSSHArgs(host *hostsini.Host) []string {
	if host == nil {
		return nil
	}
	args := make([]string, 0)

	if host.PrivateKey != "" {
		args = append(args, "-i", host.PrivateKey)
	}

	if host.Port != 0 {
		args = append(args, "-p", strconv.Itoa(host.Port))
	}

	if host.User != "" {
		args = append(args, host.User+"@"+host.Host)
	}

	return args
}

func getInventory(file string, defaults *config.Defaults) *hostsini.Hosts {
	defaultHost := hostsini.Host{
		Port:       defaults.Port,
		User:       defaults.User,
		SSHPass:    defaults.SSHPass,
		BecomePass: defaults.BecomePass,
		PrivateKey: defaults.PrivateKey,
	}
	inventory, err := hostsini.NewFile(file, defaultHost)
	if err != nil {
		debug("error parsing inventory", err)
		return nil
	}
	return inventory
}

func debug(args ...any) {
	if !withDebug {
		return
	}
	logger.Println(args...)
}
