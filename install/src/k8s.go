package main

import (
	"context"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"strings"
)

type k8s struct {
	core          v1.CoreV1Interface
	config        string // FIXME remove
	serverVersion string
}

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

func (k *k8s) kubecfg(ctx string) {
	var kubeconfig *string
	var config *rest.Config
	var err error

	if k.core == nil {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
				"(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		//flag.Parse()

		k.config = *kubeconfig

		config, err = buildConfigFromFlags(ctx, *kubeconfig)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		config, err = buildConfigFromFlags(ctx, k.config)
		if err != nil {
			log.Fatal(err)
		}
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)

	if err != nil {
		fmt.Printf(" error in discoveryClient %v", err)
	}

	information, err := discoveryClient.ServerVersion()
	if err != nil {
		fmt.Println("Error while fetching server version information", err)
	}

	k.serverVersion = strings.ReplaceAll(information.GitVersion, "v", "")

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	k.core = clientset.CoreV1()
}

func (k *k8s) getNodeIngress() string {
	pod := k.core.Pods("ingress-nginx")
	opts := metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=ingress-nginx",
	}

	p, err := pod.List(context.TODO(), opts)
	if err != nil {
		log.Fatal(err)
	}

	if len(p.Items) > 0 {
		return p.Items[0].Spec.NodeName
	} else {
		return ""
	}
}

func (k *k8s) getIngressMinikube() string {
	node := k.getNodeIngress()

	if node == "" {
		log.Fatal("Ingress pod doesn't exist")
	}

	n, err := k.core.Nodes().Get(context.TODO(), node, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	ip := n.Status.Addresses[0].Address

	return ip
}

func (k *k8s) getSvcIp(namespace string, name string) string {
	svc, err := k.core.Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		log.Fatal(err)
	}

	return svc.Spec.ClusterIP
}
