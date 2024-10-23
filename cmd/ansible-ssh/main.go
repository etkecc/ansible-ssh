package main

import (
	"os"

	"github.com/adrg/xdg"

	"github.com/etkecc/ansible-ssh/internal/ansible"
	"github.com/etkecc/ansible-ssh/internal/config"
	"github.com/etkecc/ansible-ssh/internal/logger"
	"github.com/etkecc/ansible-ssh/internal/ssh"
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
	logger.Configure(cfg.Debug)

	environ := make([]string, 0)
	for k, v := range cfg.Environ {
		environ = append(environ, k+"="+v)
	}

	host := ansible.GetHost(cfg.Path, os.Args[1], &cfg.Defaults)
	if host == nil {
		ssh.Run(cfg.SSHCommand, nil, cfg.InventoryOnly, environ)
		return
	}

	logger.Debug("host", host.Name, "has been found, starting ssh")
	ssh.Run(cfg.SSHCommand, host, cfg.InventoryOnly, environ)
}
