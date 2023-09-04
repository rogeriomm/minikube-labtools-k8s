package main

import (
	"context"
	"flag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os/exec"
	"path/filepath"
	"strings"
)

type k8s struct {
	core          v1.CoreV1Interface
	clientSet     *kubernetes.Clientset
	config        string
	serverVersion string // K8S server version
	ctx           string
}

var kubeConfig *string

/*
https://github.com/kubernetes/client-go/issues/192
*/
func buildConfigFromFlags(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

func buildConfigOverrideFlags(context string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

func (k *k8s) createNamespace(namespace string) error {
	_, err := k.core.Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})

	if err != nil {
		nsName := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}

		_, err = k.core.Namespaces().Create(context.TODO(), nsName, metav1.CreateOptions{})
	}

	return err
}

func (k *k8s) connect() {
	var config *rest.Config
	var err error

	if k.core != nil {
		return
	}

	if kubeConfig == nil {
		if home := homedir.HomeDir(); home != "" {
			kubeConfig = flag.String("kubeConfig", filepath.Join(home, ".kube", "config"),
				"(optional) absolute path to the kubeConfig file")
		} else {
			kubeConfig = flag.String("kubeConfig", "", "absolute path to the kubeConfig file")
		}
		//flag.Parse()
	}

	k.config = *kubeConfig

	config, err = buildConfigFromFlags(k.ctx, *kubeConfig)

	if err != nil {
		sugar.Fatal(err, config)
	}

	// Get server version
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)

	if err != nil {
		sugar.Fatal(" error in discoveryClient %v", err)
	}

	information, err := discoveryClient.ServerVersion()
	if err != nil {
		sugar.Fatal("Error while fetching server version information", err)
	}

	k.serverVersion = strings.ReplaceAll(information.GitVersion, "v", "")

	k.clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		sugar.Fatal(err, k.clientSet)
	}

	k.core = k.clientSet.CoreV1()
}

func (k *k8s) getNodeIngress() string {
	pod := k.core.Pods("ingress-nginx")
	opts := metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=ingress-nginx",
	}

	p, err := pod.List(context.TODO(), opts)
	if err != nil {
		sugar.Fatal(err)
	}

	if len(p.Items) > 0 {
		return p.Items[0].Spec.NodeName
	}

	return ""
}

func (k *k8s) getIngressMinikube() string {
	node := k.getNodeIngress()

	if node == "" {
		sugar.Fatal("Ingress pod doesn't exist")
	}

	n, err := k.core.Nodes().Get(context.TODO(), node, metav1.GetOptions{})
	if err != nil {
		sugar.Fatal(err)
	}

	ip := n.Status.Addresses[0].Address

	return ip
}

func (k *k8s) getSvcIp(namespace string, name string) string {
	svc, err := k.core.Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		sugar.Fatal(err)
	}

	return svc.Spec.ClusterIP
}

func (k *k8s) asdfLocal() {
	out, err := exec.Command("asdf", "local", "kubectl", k.serverVersion).CombinedOutput()

	if err != nil {
		sugar.Fatal(err, out)
	}
}

func (k *k8s) asdfGlobal() {
	out, err := exec.Command("asdf", "global", "kubectl", k.serverVersion).CombinedOutput()

	if err != nil {
		sugar.Fatal(err, out)
	}
}

func (k *k8s) useContext() {
	k.asdfLocal()

	out, err := exec.Command("kubectl", "config", "use-context", k.ctx).CombinedOutput()

	if err != nil {
		sugar.Fatal(err, out)
	}
}

func (k *k8s) ctl(arg ...string) error {

	out, err := exec.Command("kubectl", arg...).CombinedOutput()

	if err != nil {
		sugar.Error(string(out), err)
	}

	return err
}
