package proxy

import (
	"github.com/containernetworking/plugins/pkg/ns"
)

//go:generate counterfeiter -o fakes/netNS.go --fake-name NetNS . netNS
type netNS interface {
	ns.NetNS
}
