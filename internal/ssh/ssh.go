package ssh

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/etkecc/ansible-ssh/internal/logger"
	"github.com/etkecc/go-ansible"
)

var legitExitCode = map[int]bool{
	0:   true, // normal exit
	130: true, // Ctrl+C
}

// Run executes the ssh command
func Run(sshCmd string, host *ansible.Host, strict bool, environ []string) {
	cmd := buildCMD(sshCmd, host, strict)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	env := append(os.Environ(), environ...)
	cmd.Env = env

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
		logger.Debug("command:", sshCmd, sshArgs)
		return exec.Command(sshCmd, sshArgs...)
	}

	logger.Debug("command:", sshCmd, buildArgs(sshArgs, osArgs, host))

	if host.SSHPass != "" {
		logger.Println("ssh password is:", host.SSHPass)
	}

	if host.BecomePass != "" && host.User != "root" {
		logger.Println("become password is:", host.BecomePass)
	}
	return exec.Command(sshCmd, buildArgs(sshArgs, osArgs, host)...) //nolint:gosec // that's intended
}

func buildArgs(sshArgs, osArgs []string, host *ansible.Host) []string {
	if host == nil {
		return nil
	}
	if sshArgs == nil {
		sshArgs = make([]string, 0)
	}

	if len(host.PrivateKeys) > 0 {
		for _, key := range host.PrivateKeys {
			sshArgs = append(sshArgs, "-i", key)
		}
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
