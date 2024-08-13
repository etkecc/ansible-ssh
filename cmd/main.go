package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"github.com/etkecc/go-ansible"

	"github.com/etkecc/ansible-ssh/internal/config"
)

var (
	withDebug     bool
	legitExitCode = map[int]bool{
		0:   true, // normal exit
		130: true, // Ctrl+C
	}
	logger = log.New(os.Stdout, "[ansible.ssh] ", 0)
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

//nolint:nolintlint // please, don't
func buildCMD(sshCmd string, host *ansible.Host, strict bool) *exec.Cmd {
	osArgs := os.Args[1:]
	sshArgs := make([]string, 0)
	parts := strings.Split(sshCmd, " ")
	if len(parts) > 1 {
		sshCmd = parts[0]
		sshArgs = parts[1:]
	}

	if host == nil {
		if strict {
			logger.Fatal("host not found within inventory")
		}
		sshArgs = append(sshArgs, osArgs...)
		debug("command:", sshCmd, sshArgs)
		return exec.Command(sshCmd, sshArgs...) //nolint:gosec // that's intended
	}

	debug("command:", sshCmd, buildSSHArgs(sshArgs, osArgs, host))

	if host.SSHPass != "" {
		logger.Println("ssh password is:", host.SSHPass)
	}

	if host.BecomePass != "" {
		logger.Println("become password is:", host.BecomePass)
	}
	return exec.Command(sshCmd, buildSSHArgs(sshArgs, osArgs, host)...) //nolint:gosec // that's intended
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
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && legitExitCode[exitErr.ExitCode()] {
			return
		}
		logger.Fatal("command failed:", err)
	}
}

func buildSSHArgs(sshArgs, osArgs []string, host *ansible.Host) []string {
	if host == nil {
		return nil
	}
	if sshArgs == nil {
		sshArgs = make([]string, 0)
	}

	if host.PrivateKey != "" {
		sshArgs = append(sshArgs, "-i", host.PrivateKey)
	}

	if host.Port != 0 {
		sshArgs = append(sshArgs, "-p", strconv.Itoa(host.Port))
	}

	if host.User != "" {
		sshArgs = append(sshArgs, host.User+"@"+host.Host)
	}

	if len(osArgs) > 1 {
		sshArgs = append(sshArgs, osArgs[1:]...)
	}

	return sshArgs
}

func debug(args ...any) {
	if !withDebug {
		return
	}
	logger.Println(args...)
}
