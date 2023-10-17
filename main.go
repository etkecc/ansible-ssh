package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/adrg/xdg"
	"gitlab.com/etke.cc/go/ansible"

	"gitlab.com/etke.cc/tools/ansible-ssh/config"
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

	inv := ansible.ParseInventory("ansible.cfg", cfg.Path, os.Args[1])
	if inv == nil {
		debug("inventory not found")
		executeSSH(cfg.SSHCommand, nil, cfg.InventoryOnly)
		return
	}
	host := inv.Hosts[os.Args[1]]
	if host == nil {
		debug("host", os.Args[1], "not found in inventory")
		executeSSH(cfg.SSHCommand, nil, cfg.InventoryOnly)
		return
	}
	host = ansible.MergeHost(host, &ansible.Host{
		User:       cfg.Defaults.User,
		Port:       cfg.Defaults.Port,
		SSHPass:    cfg.Defaults.SSHPass,
		BecomePass: cfg.Defaults.BecomePass,
		PrivateKey: cfg.Defaults.PrivateKey,
	})
	debug("host", host.Name, "has been found, starting ssh")
	executeSSH(cfg.SSHCommand, host, cfg.InventoryOnly)
}

func buildCMD(sshCmd string, host *ansible.Host, strict bool) *exec.Cmd {
	if host == nil {
		if strict {
			logger.Fatal("host not found within inventory")
		}
		debug("command:", sshCmd, os.Args[1:])
		return exec.Command(sshCmd, os.Args[1:]...) //nolint:gosec // that's intended
	}
	debug("command:", sshCmd, buildSSHArgs(host))

	if host.SSHPass != "" {
		logger.Println("ssh password is:", host.SSHPass)
	}

	if host.BecomePass != "" {
		logger.Println("become password is:", host.BecomePass)
	}
	return exec.Command(sshCmd, buildSSHArgs(host)...) //nolint:gosec // that's intended
}

func executeSSH(sshCmd string, host *ansible.Host, strict bool) {
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

func buildSSHArgs(host *ansible.Host) []string {
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

func debug(args ...any) {
	if !withDebug {
		return
	}
	logger.Println(args...)
}
