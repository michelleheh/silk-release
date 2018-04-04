package legacynet

import (
	"fmt"
	"lib/rules"
	"net"

	multierror "github.com/hashicorp/go-multierror"
)

const prefixNetIn = "netin"

type NetIn struct {
	ChainNamer        chainNamer
	IPTables          rules.IPTablesAdapter
	IngressTag        string
	HostInterfaceName string
}

func (m *NetIn) Initialize(containerHandle string) error {
	chain := m.ChainNamer.Prefix(prefixNetIn, containerHandle)

	args := []fullRule{
		{
			Table:       "nat",
			ParentChain: "PREROUTING",
			Chain:       chain,
			Rules:       []rules.IPTablesRule{},
		},
		{
			Table:       "mangle",
			ParentChain: "PREROUTING",
			Chain:       chain,
			Rules:       []rules.IPTablesRule{},
		},
	}

	return initChains(m.IPTables, args)
}

func (m *NetIn) Cleanup(containerHandle string) error {
	chain := m.ChainNamer.Prefix(prefixNetIn, containerHandle)

	var result error
	err := cleanupChain("nat", "PREROUTING", chain, rules.IPTablesRule{}, m.IPTables)
	if err != nil {
		result = multierror.Append(result, err)
	}

	err = cleanupChain("mangle", "PREROUTING", chain, rules.IPTablesRule{}, m.IPTables)
	if err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func (m *NetIn) AddRule(containerHandle string,
	hostPort, containerPort int, hostIP, containerIP string) error {
	chain := m.ChainNamer.Prefix(prefixNetIn, containerHandle)

	parsedIP := net.ParseIP(hostIP)
	if parsedIP == nil {
		return fmt.Errorf("invalid ip: %s", hostIP)
	}

	parsedIP = net.ParseIP(containerIP)
	if parsedIP == nil {
		return fmt.Errorf("invalid ip: %s", containerIP)
	}

	args := []fullRule{
		{
			Table:       "nat",
			ParentChain: "PREROUTING",
			Chain:       chain,
			Rules: []rules.IPTablesRule{
				rules.NewPortForwardingRule(hostPort, containerPort, hostIP, containerIP),
			},
		},
		{
			Table:       "mangle",
			ParentChain: "PREROUTING",
			Chain:       chain,
			Rules: []rules.IPTablesRule{
				rules.NewIngressMarkRule(m.HostInterfaceName, hostPort, hostIP, m.IngressTag),
			},
		},
	}

	return applyRules(m.IPTables, args)
}

func initChains(iptables rules.IPTablesAdapter, args []fullRule) error {
	for _, arg := range args {
		err := iptables.NewChain(arg.Table, arg.Chain)
		if err != nil {
			return fmt.Errorf("creating chain: %s", err)
		}

		if arg.ParentChain == "INPUT" {
			for _, jumpCondition := range arg.JumpConditions {
				err = iptables.BulkAppend(arg.Table, arg.ParentChain, append(
					jumpCondition, rules.IPTablesRule{"--jump", arg.Chain}...))
				if err != nil {
					return fmt.Errorf("appending rule to INPUT chain: %s", err)
				}
			}
		} else if arg.ParentChain != "" {
			for _, jumpCondition := range arg.JumpConditions {
				err = iptables.BulkInsert(arg.Table, arg.ParentChain, 1, append(
					jumpCondition, rules.IPTablesRule{"--jump", arg.Chain}...))
				if err != nil {
					return fmt.Errorf("inserting rule: %s", err)
				}
			}
		}
	}

	return nil
}

func applyRules(iptables rules.IPTablesAdapter, args []fullRule) error {
	for _, arg := range args {
		err := iptables.BulkAppend(arg.Table, arg.Chain, arg.Rules...)
		if err != nil {
			return fmt.Errorf("appending rule: %s", err)
		}
	}

	return nil
}

func cleanupChain(table, parentChain, chain string, jumpConditions rules.IPTablesRule, iptables rules.IPTablesAdapter) error {
	var result error
	if parentChain != "" {
		if err := iptables.Delete(table, parentChain, append(jumpConditions, rules.IPTablesRule{"--jump", chain}...)); err != nil {
			result = multierror.Append(result, fmt.Errorf("delete rule: %s", err))
		}
	}

	if err := iptables.ClearChain(table, chain); err != nil {
		result = multierror.Append(result, fmt.Errorf("clear chain: %s", err))
	}

	if err := iptables.DeleteChain(table, chain); err != nil {
		result = multierror.Append(result, fmt.Errorf("delete chain: %s", err))
	}
	return result
}
