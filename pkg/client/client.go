package client

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	LabelNullplatform  = "nullplatform"
	LabelApplicationID = "application_id"
	LabelScopeID       = "scope_id"
	LabelDeploymentID  = "deployment_id"
)

type KubeClient interface {
	GetPods(namespace string, labels map[string]string) ([]corev1.Pod, error)
	GetLogs(namespace, podName string, options *corev1.PodLogOptions) (string, error)
}

type DefaultClient struct {
	clientSet kubernetes.Interface
}

func NewClient() (KubeClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
		}
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &DefaultClient{
		clientSet: clientSet,
	}, nil
}

func (c *DefaultClient) GetPods(namespace string, labels map[string]string) ([]corev1.Pod, error) {
	labelSelector := &metav1.LabelSelector{
		MatchLabels: labels,
	}

	selector, err := metav1.LabelSelectorAsSelector(labelSelector)

	if err != nil {
		return nil, err
	}

	pods, err := c.clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if pods == nil {
		return []corev1.Pod{}, nil
	}

	return pods.Items, nil
}

func (c *DefaultClient) GetLogs(namespace, podName string, options *corev1.PodLogOptions) (string, error) {
	req := c.clientSet.CoreV1().Pods(namespace).GetLogs(podName, options)

	podLogs, err := req.DoRaw(context.TODO())

	if err != nil {
		return "", err
	}

	return string(podLogs), nil
}
