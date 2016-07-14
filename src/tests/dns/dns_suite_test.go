package dns_test

import (
	"tests/helpers"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/kubernetes/pkg/api/v1"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"
	v1core "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3/typed/core/v1"
	"k8s.io/kubernetes/pkg/client/restclient"

	"testing"
)

var clientConfig *restclient.Config
var client v1core.CoreInterface

func TestDns(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dns Suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(5 * time.Second)

	config, err := helpers.Load()
	Expect(err).NotTo(HaveOccurred())

	clientConfig = config.ClientConfig()
	clientSet, err := clientset.NewForConfig(clientConfig)
	Expect(err).NotTo(HaveOccurred())

	client = clientSet.Core()
})

func podIsRunning(client v1core.CoreInterface, pod *v1.Pod) func() bool {
	return func() bool {
		if p, err := client.Pods(pod.Namespace).Get(pod.Name); err == nil {
			return p.Status.Phase == v1.PodRunning
		}
		return false
	}
}
