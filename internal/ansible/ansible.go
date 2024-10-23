package ansible

import (
	"strings"

	"github.com/etkecc/ansible-ssh/internal/config"
	"github.com/etkecc/ansible-ssh/internal/logger"
	"github.com/etkecc/go-ansible"
)

// inventoryPrefixWorkaround is a workaround for ssh key path defined in the inventory/host_vars/DOMAIN/vars.yml file
// at this moment, we do not have a way to parse it properly, so we just assume that the key is in the same directory as vars.yml file when
// this prefix is used in the key path.
// Consider that a very hacky PoC that should be replaced with a proper implementation in the future.
const inventoryPrefixWorkaround = "{{ playbook_dir }}/../../inventory/host_vars/{{ inventory_hostname }}/"

// GetHost returns a host from the inventory
func GetHost(hostsini, limit string, defaults *config.Defaults) *ansible.Host {
	inv := ansible.ParseInventory("ansible.cfg", hostsini, limit)
	if inv == nil {
		logger.Debug("inventory not found")
		return nil
	}
	host := inv.Hosts[limit]
	if host == nil {
		logger.Debug("host", limit, "not found in inventory")
		return nil
	}
	host = ansible.MergeHost(host, &ansible.Host{
		User:        defaults.User,
		Port:        defaults.Port,
		SSHPass:     defaults.SSHPass,
		BecomePass:  defaults.BecomePass,
		PrivateKeys: defaults.PrivateKeys,
	})

	// replace inventoryPrefixWorkaround with the actual path,
	// details are in the inventoryPrefixWorkaround const description
	for _, invPath := range inv.Paths {
		invPath = strings.TrimSuffix(invPath, "hosts")
		for i, key := range host.PrivateKeys {
			if strings.HasPrefix(key, inventoryPrefixWorkaround) {
				keypath := strings.Replace(key, inventoryPrefixWorkaround, invPath+"host_vars/"+host.Name+"/", 1)
				host.PrivateKeys[i] = keypath
			}
		}
	}

	return host
}
