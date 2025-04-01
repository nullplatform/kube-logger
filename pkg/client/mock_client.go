package client

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MockClient struct {
	Pods map[string][]corev1.Pod
	Logs map[string]string
}

func NewMockClient() *MockClient {
	return &MockClient{
		Pods: make(map[string][]corev1.Pod),
		Logs: make(map[string]string),
	}
}

func (m *MockClient) GetPods(namespace string, labels map[string]string) ([]corev1.Pod, error) {
	if pods, ok := m.Pods[namespace]; ok {
		return pods, nil
	}
	return []corev1.Pod{}, nil
}

func (m *MockClient) GetLogs(namespace, podName string, options *corev1.PodLogOptions) (string, error) {
	key := namespace + "/" + podName
	if logs, ok := m.Logs[key]; ok {
		return logs, nil
	}
	return "", nil
}

func (m *MockClient) AddPod(namespace string, pod corev1.Pod) {
	if _, ok := m.Pods[namespace]; !ok {
		m.Pods[namespace] = []corev1.Pod{}
	}
	m.Pods[namespace] = append(m.Pods[namespace], pod)
}

func (m *MockClient) AddLogs(namespace, podName, logs string) {
	key := namespace + "/" + podName
	m.Logs[key] = logs
}

func CreateTestPod(name string, labels map[string]string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
			Labels:    labels,
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
}
