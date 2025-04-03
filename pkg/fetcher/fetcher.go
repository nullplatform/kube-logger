package fetcher

import (
	"fmt"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nullplatform/kube-logger/pkg/api"
	"github.com/nullplatform/kube-logger/pkg/client"
	"github.com/nullplatform/kube-logger/pkg/token"
)

const (
	DefaultContainerName = "application"

	MinLogsPerPod = 10
)

type LogFetcher interface {
	FetchLogs(config api.Config) (*api.Result, error)
}

type logFetcher struct {
	kubeClient client.KubeClient
}

func New(client client.KubeClient) LogFetcher {
	return &logFetcher{
		kubeClient: client,
	}
}

func (f *logFetcher) FetchLogs(config api.Config) (*api.Result, error) {
	labels := buildLabels(config)

	pods, err := f.kubeClient.GetPods(config.Namespace, labels)

	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods) == 0 {
		return &api.Result{
			Results:       []api.LogEntry{},
			NextPageToken: "",
		}, nil
	}

	lastReadTimes, err := token.Decode(config.NextPageToken)

	if err != nil {
		return nil, fmt.Errorf("failed to decode pagination token: %w", err)
	}

	allLogs, newTokenData, err := f.processAllPods(pods, lastReadTimes, config)

	if err != nil {
		return nil, err
	}

	sortedLogs := sortAndLimitLogs(allLogs, config.Limit)

	nextPageToken, err := token.Encode(newTokenData)

	if err != nil {
		return nil, fmt.Errorf("failed to encode pagination token: %w", err)
	}

	return &api.Result{
		Results:       sortedLogs,
		NextPageToken: nextPageToken,
	}, nil
}

func buildLabels(config api.Config) map[string]string {
	labels := map[string]string{
		client.LabelNullplatform: "true",
	}

	if config.ApplicationID != "" {
		labels[client.LabelApplicationID] = config.ApplicationID
	}

	if config.ScopeID != "" {
		labels[client.LabelScopeID] = config.ScopeID
	}

	if config.DeploymentID != "" {
		labels[client.LabelDeploymentID] = config.DeploymentID
	}

	return labels
}

func (f *logFetcher) processAllPods(pods []corev1.Pod, lastReadTimes token.TokenData, config api.Config) ([]api.LogEntry, token.TokenData, error) {
	var allLogs []api.LogEntry
	newTokenData := make(token.TokenData)

	podLimit := calculatePodLimit(config.Limit, len(pods))

	for _, pod := range pods {
		logs, lastTimestamp, err := f.processOnePod(pod, lastReadTimes, config, podLimit)
		if err != nil {
			fmt.Printf("Error fetching logs from pod %s: %v\n", pod.Name, err)
			continue
		}

		allLogs = append(allLogs, logs...)

		if lastTimestamp != "" {
			newTokenData[pod.Name] = lastTimestamp
		}
	}

	return allLogs, newTokenData, nil
}

func (f *logFetcher) processOnePod(pod corev1.Pod, lastReadTimes token.TokenData, config api.Config, podLimit int) ([]api.LogEntry, string, error) {
	sinceTime := determineSinceTime(pod.Name, lastReadTimes, config.StartTime)

	logOptions := createLogOptions(sinceTime, podLimit)

	logsStr, err := f.kubeClient.GetLogs(config.Namespace, pod.Name, logOptions)
	if err != nil {
		return nil, "", err
	}

	return processLogLines(logsStr, config.FilterPattern, pod.Name)
}

func determineSinceTime(podName string, lastReadTimes token.TokenData, configStartTime string) string {
	if lastTime, ok := lastReadTimes[podName]; ok {
		return lastTime
	}

	if configStartTime != "" {
		return configStartTime
	}

	return ""
}

func createLogOptions(sinceTime string, podLimit int) *corev1.PodLogOptions {
	options := &corev1.PodLogOptions{
		Container:  DefaultContainerName,
		Timestamps: true,
		LimitBytes: func(v int64) *int64 { return &v }(int64(podLimit * 1024)),
	}

	if sinceTime != "" {
		parsedTime, err := time.Parse(time.RFC3339Nano, sinceTime)
		if err == nil {
			metaTime := metav1.NewTime(parsedTime)
			options.SinceTime = &metaTime
		}
	}

	return options
}

func processLogLines(logsStr string, filterPattern string, podName string) ([]api.LogEntry, string, error) {
	logLines := strings.Split(logsStr, "\n")

	var logs []api.LogEntry
	var lastTimestamp string

	for _, line := range logLines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		timestamp := parts[0]
		content := parts[1]

		if filterPattern != "" && !strings.Contains(line, filterPattern) {
			continue
		}

		logs = append(logs, api.LogEntry{
			Message:  content,
			Datetime: timestamp,
		})
		lastTimestamp = timestamp
	}

	return logs, lastTimestamp, nil
}

func calculatePodLimit(totalLimit, podCount int) int {
	podLimit := totalLimit / podCount

	if podLimit < MinLogsPerPod {
		podLimit = MinLogsPerPod
	}

	return podLimit
}

func sortAndLimitLogs(logs []api.LogEntry, limit int) []api.LogEntry {
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Datetime < logs[j].Datetime
	})

	if len(logs) > limit {
		logs = logs[:limit]
	}

	return logs
}
