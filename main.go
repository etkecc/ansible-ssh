package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/adrg/xdg"

	"gitlab.com/etke.cc/int/ansible-ssh/aini"
	"gitlab.com/etke.cc/int/ansible-ssh/config"
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

	inv := getInventory(cfg.Path, cfg.Defaults)
	if inv == nil {
		debug("inventory not found")
		executeSSH(cfg.SSHCommand, nil)
		return
	}
	host := inv.MatchOne(os.Args[1])
	if host == nil {
		debug("host", os.Args[1], "not found in inventory")
		executeSSH(cfg.SSHCommand, nil)
		return
	}
	debug("host", host.Name, "has been found, starting ssh")
	executeSSH(cfg.SSHCommand, host)
}

func executeSSH(sshCmd string, host *aini.Host) {
	var cmd *exec.Cmd
	if host == nil {
		debug("command:", sshCmd, os.Args[1:])
		cmd = exec.Command(sshCmd, os.Args[1:]...)
	} else {
		debug("command:", sshCmd, buildSSHArgs(host))
		cmd = exec.Command(sshCmd, buildSSHArgs(host)...)

		if host.SSHPass != "" {
			logger.Println("ssh password is:", host.SSHPass)
		}

		if host.BecomePass != "" {
			logger.Println("become password is:", host.BecomePass)
		}
	}
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

func buildSSHArgs(host *aini.Host) []string {
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

func getInventory(file string, defaults config.Defaults) *aini.Hosts {
	defaultHost := aini.Host{
		Port:       defaults.Port,
		User:       defaults.User,
		SSHPass:    defaults.SSHPass,
		BecomePass: defaults.BecomePass,
		PrivateKey: defaults.PrivateKey,
	}
	inventory, err := aini.NewFile(file, defaultHost)
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
