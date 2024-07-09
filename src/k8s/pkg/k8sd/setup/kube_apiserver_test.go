package setup_test

import (
	"fmt"
	"net"
	"os"
	"path"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

var apiserverTLSCipherSuites = "TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_RSA_WITH_3DES_EDE_CBC_SHA,TLS_RSA_WITH_AES_128_CBC_SHA,TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_CBC_SHA,TLS_RSA_WITH_AES_256_GCM_SHA384"

func setKubeAPIServerMock(s *mock.Snap, dir string) {
	s.Mock = mock.Mock{
		UID:                   os.Getuid(),
		GID:                   os.Getgid(),
		KubernetesConfigDir:   path.Join(dir, "kubernetes"),
		KubernetesPKIDir:      path.Join(dir, "kubernetes-pki"),
		ServiceArgumentsDir:   path.Join(dir, "args"),
		ServiceExtraConfigDir: path.Join(dir, "args/conf.d"),
		K8sDqliteStateDir:     path.Join(dir, "k8s-dqlite"),
	}
}

func TestKubeAPIServer(t *testing.T) {
	t.Run("ArgsWithProxy", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeAPIServerMock)

		// Call the KubeAPIServer setup function with mock arguments
		g.Expect(setup.KubeAPIServer(s, net.ParseIP("192.168.0.1"), "10.0.0.0/24", "https://auth-webhook.url", true, types.Datastore{Type: utils.Pointer("k8s-dqlite")}, "Node,RBAC", nil)).To(BeNil())

		// Ensure the kube-apiserver arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--advertise-address", expectedVal: "192.168.0.1"},
			{key: "--anonymous-auth", expectedVal: "false"},
			{key: "--allow-privileged", expectedVal: "true"},
			{key: "--authentication-token-webhook-config-file", expectedVal: path.Join(s.Mock.ServiceExtraConfigDir, "auth-token-webhook.conf")},
			{key: "--authorization-mode", expectedVal: "Node,RBAC"},
			{key: "--client-ca-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "client-ca.crt")},
			{key: "--enable-admission-plugins", expectedVal: "NodeRestriction"},
			{key: "--kubelet-certificate-authority", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--kubelet-client-certificate", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "apiserver-kubelet-client.crt")},
			{key: "--kubelet-client-key", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "apiserver-kubelet-client.key")},
			{key: "--kubelet-preferred-address-types", expectedVal: "InternalIP,Hostname,InternalDNS,ExternalDNS,ExternalIP"},
			{key: "--profiling", expectedVal: "false"},
			{key: "--secure-port", expectedVal: "6443"},
			{key: "--service-account-issuer", expectedVal: "https://kubernetes.default.svc"},
			{key: "--service-account-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "serviceaccount.key")},
			{key: "--service-account-signing-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "serviceaccount.key")},
			{key: "--service-cluster-ip-range", expectedVal: "10.0.0.0/24"},
			{key: "--tls-cert-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "apiserver.crt")},
			{key: "--tls-cipher-suites", expectedVal: apiserverTLSCipherSuites},
			{key: "--tls-private-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "apiserver.key")},
			{key: "--etcd-servers", expectedVal: fmt.Sprintf("unix://%s", path.Join(s.Mock.K8sDqliteStateDir, "k8s-dqlite.sock"))},
			{key: "--request-timeout", expectedVal: "300s"},
			{key: "--requestheader-client-ca-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "front-proxy-ca.crt")},
			{key: "--requestheader-allowed-names", expectedVal: "front-proxy-client"},
			{key: "--requestheader-extra-headers-prefix", expectedVal: "X-Remote-Extra-"},
			{key: "--requestheader-group-headers", expectedVal: "X-Remote-Group"},
			{key: "--requestheader-username-headers", expectedVal: "X-Remote-User"},
			{key: "--proxy-client-cert-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "front-proxy-client.crt")},
			{key: "--proxy-client-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "front-proxy-client.key")},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-apiserver", tc.key)
				g.Expect(err).To(BeNil())
				g.Expect(val).To(Equal(tc.expectedVal))
			})
		}

		// Ensure the kube-apiserver arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kube-apiserver"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("ArgsNoProxy", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeAPIServerMock)

		// Call the KubeAPIServer setup function with mock arguments
		g.Expect(setup.KubeAPIServer(s, net.ParseIP("192.168.0.1"), "10.0.0.0/24", "https://auth-webhook.url", false, types.Datastore{Type: utils.Pointer("k8s-dqlite")}, "Node,RBAC", nil)).To(BeNil())

		// Ensure the kube-apiserver arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--advertise-address", expectedVal: "192.168.0.1"},
			{key: "--anonymous-auth", expectedVal: "false"},
			{key: "--allow-privileged", expectedVal: "true"},
			{key: "--authentication-token-webhook-config-file", expectedVal: path.Join(s.Mock.ServiceExtraConfigDir, "auth-token-webhook.conf")},
			{key: "--authorization-mode", expectedVal: "Node,RBAC"},
			{key: "--client-ca-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "client-ca.crt")},
			{key: "--enable-admission-plugins", expectedVal: "NodeRestriction"},
			{key: "--kubelet-certificate-authority", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--kubelet-client-certificate", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "apiserver-kubelet-client.crt")},
			{key: "--kubelet-client-key", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "apiserver-kubelet-client.key")},
			{key: "--kubelet-preferred-address-types", expectedVal: "InternalIP,Hostname,InternalDNS,ExternalDNS,ExternalIP"},
			{key: "--profiling", expectedVal: "false"},
			{key: "--request-timeout", expectedVal: "300s"},
			{key: "--secure-port", expectedVal: "6443"},
			{key: "--service-account-issuer", expectedVal: "https://kubernetes.default.svc"},
			{key: "--service-account-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "serviceaccount.key")},
			{key: "--service-account-signing-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "serviceaccount.key")},
			{key: "--service-cluster-ip-range", expectedVal: "10.0.0.0/24"},
			{key: "--tls-cert-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "apiserver.crt")},
			{key: "--tls-cipher-suites", expectedVal: apiserverTLSCipherSuites},
			{key: "--tls-private-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "apiserver.key")},
			{key: "--etcd-servers", expectedVal: fmt.Sprintf("unix://%s", path.Join(s.Mock.K8sDqliteStateDir, "k8s-dqlite.sock"))},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-apiserver", tc.key)
				g.Expect(err).To(BeNil())
				g.Expect(val).To(Equal(tc.expectedVal))
			})
		}

		// Ensure the kube-apiserver arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kube-apiserver"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("WithExtraArgs", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeAPIServerMock)

		extraArgs := map[string]*string{
			"--allow-privileged": nil,
			"--secure-port":      utils.Pointer("1337"),
			"--my-extra-arg":     utils.Pointer("my-extra-val"),
		}
		// Call the KubeAPIServer setup function with mock arguments
		g.Expect(setup.KubeAPIServer(s, net.ParseIP("192.168.0.1"), "10.0.0.0/24", "https://auth-webhook.url", true, types.Datastore{Type: utils.Pointer("k8s-dqlite")}, "Node,RBAC", extraArgs)).To(BeNil())

		// Ensure the kube-apiserver arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--advertise-address", expectedVal: "192.168.0.1"},
			{key: "--anonymous-auth", expectedVal: "false"},
			{key: "--authentication-token-webhook-config-file", expectedVal: path.Join(s.Mock.ServiceExtraConfigDir, "auth-token-webhook.conf")},
			{key: "--authorization-mode", expectedVal: "Node,RBAC"},
			{key: "--client-ca-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "client-ca.crt")},
			{key: "--enable-admission-plugins", expectedVal: "NodeRestriction"},
			{key: "--kubelet-certificate-authority", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--kubelet-client-certificate", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "apiserver-kubelet-client.crt")},
			{key: "--kubelet-client-key", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "apiserver-kubelet-client.key")},
			{key: "--kubelet-preferred-address-types", expectedVal: "InternalIP,Hostname,InternalDNS,ExternalDNS,ExternalIP"},
			{key: "--profiling", expectedVal: "false"},
			{key: "--secure-port", expectedVal: "1337"},
			{key: "--service-account-issuer", expectedVal: "https://kubernetes.default.svc"},
			{key: "--service-account-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "serviceaccount.key")},
			{key: "--service-account-signing-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "serviceaccount.key")},
			{key: "--service-cluster-ip-range", expectedVal: "10.0.0.0/24"},
			{key: "--tls-cert-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "apiserver.crt")},
			{key: "--tls-cipher-suites", expectedVal: apiserverTLSCipherSuites},
			{key: "--tls-private-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "apiserver.key")},
			{key: "--etcd-servers", expectedVal: fmt.Sprintf("unix://%s", path.Join(s.Mock.K8sDqliteStateDir, "k8s-dqlite.sock"))},
			{key: "--request-timeout", expectedVal: "300s"},
			{key: "--requestheader-client-ca-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "front-proxy-ca.crt")},
			{key: "--requestheader-allowed-names", expectedVal: "front-proxy-client"},
			{key: "--requestheader-extra-headers-prefix", expectedVal: "X-Remote-Extra-"},
			{key: "--requestheader-group-headers", expectedVal: "X-Remote-Group"},
			{key: "--requestheader-username-headers", expectedVal: "X-Remote-User"},
			{key: "--proxy-client-cert-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "front-proxy-client.crt")},
			{key: "--proxy-client-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "front-proxy-client.key")},
			{key: "--my-extra-arg", expectedVal: "my-extra-val"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-apiserver", tc.key)
				g.Expect(err).To(BeNil())
				g.Expect(val).To(Equal(tc.expectedVal))
			})
		}
		// Ensure that the allow-privileged argument was deleted
		val, err := snaputil.GetServiceArgument(s, "kube-apiserver", "--allow-privileged")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(val).To(BeZero())

		// Ensure the kube-apiserver arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kube-apiserver"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})
	t.Run("ArgsDualstack", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setKubeAPIServerMock)

		// Setup without proxy to simplify argument list
		g.Expect(setup.KubeAPIServer(s, net.ParseIP("192.168.0.1"), "10.0.0.0/24,fd01::/64", "https://auth-webhook.url", false, types.Datastore{Type: utils.Pointer("external"), ExternalServers: utils.Pointer([]string{"datastoreurl1", "datastoreurl2"})}, "Node,RBAC", nil)).To(BeNil())

		g.Expect(snaputil.GetServiceArgument(s, "kube-apiserver", "--service-cluster-ip-range")).To(Equal("10.0.0.0/24,fd01::/64"))
		_, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kube-apiserver"))
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("ArgsExternalDatastore", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setKubeAPIServerMock)

		// Setup without proxy to simplify argument list
		g.Expect(setup.KubeAPIServer(s, net.ParseIP("192.168.0.1"), "10.0.0.0/24", "https://auth-webhook.url", false, types.Datastore{Type: utils.Pointer("external"), ExternalServers: utils.Pointer([]string{"datastoreurl1", "datastoreurl2"})}, "Node,RBAC", nil)).To(BeNil())

		g.Expect(snaputil.GetServiceArgument(s, "kube-apiserver", "--etcd-servers")).To(Equal("datastoreurl1,datastoreurl2"))
		_, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kube-apiserver"))
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("UnsupportedDatastore", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeAPIServerMock)

		// Attempt to configure kube-apiserver with an unsupported datastore
		err := setup.KubeAPIServer(s, net.ParseIP("192.168.0.1"), "10.0.0.0/24", "https://auth-webhook.url", false, types.Datastore{Type: utils.Pointer("unsupported")}, "Node,RBAC", nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError(ContainSubstring("unsupported datastore")))
	})
}
