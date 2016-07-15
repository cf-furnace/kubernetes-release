package dns_test

import (
	"net/http"

	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/unversioned/remotecommand"
	remotecommandserver "k8s.io/kubernetes/pkg/kubelet/server/remotecommand"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/sets"
)

var _ = Describe("kube-dns", func() {
	It("runs in the system namesapce", func() {
		requirement, err := labels.NewRequirement("k8s-app", labels.EqualsOperator, sets.NewString("kube-dns"))
		Expect(err).NotTo(HaveOccurred())

		dnsPods, err := client.Pods("kube-system").List(api.ListOptions{
			LabelSelector: labels.NewSelector().Add(*requirement),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(dnsPods.Items).NotTo(BeEmpty())

		Expect(dnsPods.Items[0].Status.Phase).To(Equal(v1.PodRunning))
		for _, containerStatus := range dnsPods.Items[0].Status.ContainerStatuses {
			Expect(containerStatus.State.Running).NotTo(BeNil())
		}
	})

	It("is associated with a replication controller", func() {
		rc, err := client.ReplicationControllers("kube-system").Get("kube-dns-v11")
		Expect(err).NotTo(HaveOccurred())

		Expect(rc.Spec.Replicas).NotTo(BeNil())
		Expect(*rc.Spec.Replicas).NotTo(Equal(0))
		Expect(*rc.Spec.Replicas).To(Equal(rc.Status.Replicas))
	})

	Describe("name resolution", func() {
		var namespace *v1.Namespace
		var busyboxPod *v1.Pod

		BeforeEach(func() {
			guid, err := uuid.NewV4()
			Expect(err).NotTo(HaveOccurred())

			namespace, err = client.Namespaces().Create(&v1.Namespace{
				ObjectMeta: v1.ObjectMeta{Name: guid.String()},
			})
			Expect(err).NotTo(HaveOccurred())

			busybox := &v1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "busybox",
					Namespace: namespace.Name,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:            "busybox",
						Image:           "busybox",
						Command:         []string{"sleep", "3600"},
						ImagePullPolicy: v1.PullIfNotPresent,
					}},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			}

			busyboxPod, err = client.Pods(namespace.Name).Create(busybox)
			Expect(err).NotTo(HaveOccurred())
			Eventually(podIsRunning(client, busyboxPod)).Should(BeTrue())
		})

		AfterEach(func() {
			err := client.Namespaces().Delete(namespace.Name, nil)
			Expect(err).NotTo(HaveOccurred())

			err = client.Pods(namespace.Name).Delete(busyboxPod.Name, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("resolves kubernetes.default", func() {
			req := client.GetRESTClient().Post().
				Resource("pods").
				Name(busyboxPod.Name).
				Namespace(busyboxPod.Namespace).
				SubResource("exec").
				Param("container", "busybox")
			req.VersionedParams(&v1.PodExecOptions{
				Container: "busybox",
				Command:   []string{"nslookup", "kubernetes.default"},
				Stdin:     false,
				Stdout:    true,
				Stderr:    true,
				TTY:       false,
			}, api.ParameterCodec)

			executor, err := remotecommand.NewExecutor(clientConfig, http.MethodPost, req.URL())
			Expect(err).NotTo(HaveOccurred())

			err = executor.Stream(remotecommandserver.SupportedStreamingProtocols, nil, GinkgoWriter, GinkgoWriter, false)
			Expect(err).NotTo(HaveOccurred())
		})

		It("resolves consul service urls", func() {
			req := client.GetRESTClient().Post().
				Resource("pods").
				Name(busyboxPod.Name).
				Namespace(busyboxPod.Namespace).
				SubResource("exec").
				Param("container", "busybox")
			req.VersionedParams(&v1.PodExecOptions{
				Container: "busybox",
				Command:   []string{"nslookup", "kube-apiserver.service.cf.internal"},
				Stdin:     false,
				Stdout:    true,
				Stderr:    true,
				TTY:       false,
			}, api.ParameterCodec)

			executor, err := remotecommand.NewExecutor(clientConfig, http.MethodPost, req.URL())
			Expect(err).NotTo(HaveOccurred())

			err = executor.Stream(remotecommandserver.SupportedStreamingProtocols, nil, GinkgoWriter, GinkgoWriter, false)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
