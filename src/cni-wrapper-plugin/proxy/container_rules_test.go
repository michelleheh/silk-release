package proxy_test

import (
	"code.cloudfoundry.org/silk/cni/lib/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ContainerRules", func() {
	Describe("Initialize", func() {
		var (
			networkNamespace *fakes.NetNS

			containerRules proxy.ContainerRules
		)

		BeforeEach(func() {
			networkNamespace = &fakes.NetNS{}

			containerRules = proxy.ContainerRules{}
		})

		It("creates iptables rules in the network namespace to redirect overlay traffic to the proxy", func() {
			err := containerRules.Initialize(networkNamespace, overlayNetwork, proxyPort)
			Expect(err).NotTo(HaveOccurred())

			Expect(networkNamespace.DoCallCount()).To(Equal(1))
		})
	})
})
