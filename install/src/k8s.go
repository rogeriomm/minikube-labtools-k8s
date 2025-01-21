package main

import (
	"context"
	"flag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
)

type k8s struct {
	core v1.CoreV1Interface
}

func (k *k8s) kubecfg() {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

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
