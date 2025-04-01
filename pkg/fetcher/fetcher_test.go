package fetcher

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nullplatform/kube-logger/pkg/api"
	"github.com/nullplatform/kube-logger/pkg/client"
)

func TestFetchLogs(t *testing.T) {
	mockClient := client.NewMockClient()

	pod1 := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "nullplatform",
			Labels: map[string]string{
				"nullplatform":   "true",
				"application_id": "1691688910",
				"scope_id":       "760499159",
				"deployment_id":  "1705961777",
			},
		},
	}

	pod2 := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-2",
			Namespace: "nullplatform",
			Labels: map[string]string{
				"nullplatform":   "true",
				"application_id": "1691688910",
				"scope_id":       "760499159",
				"deployment_id":  "1705961777",
			},
		},
	}

	mockClient.AddPod("nullplatform", pod1)
	mockClient.AddPod("nullplatform", pod2)

	pod1Logs := "2025-04-01T15:44:44.534732559Z Available memory: 7835MB\n" +
		"2025-04-01T15:44:44.548275040Z Available CPU: 2 cores\n" +
		"2025-04-01T15:44:44.548290344Z Instances in cluster: 2\n"

	pod2Logs := "2025-04-01T15:44:45.083053679Z PM2 log: App [application:0] online\n" +
		"2025-04-01T15:44:45.083070225Z PM2 log: App [application:1] starting in -cluster mode-\n" +
		"2025-04-01T15:44:45.229413861Z PM2 log: App [application:1] online\n"

	mockClient.AddLogs("nullplatform", "pod-1", pod1Logs)
	mockClient.AddLogs("nullplatform", "pod-2", pod2Logs)

	fetcher := New(mockClient)

	testCases := []struct {
		name          string
		config        api.Config
		expectedCount int
		expectedFirst string
		expectedToken bool
	}{
		{
			name: "fetch all logs",
			config: api.Config{
				Namespace:     "nullplatform",
				ApplicationID: "1691688910",
				ScopeID:       "760499159",
				DeploymentID:  "1705961777",
				Limit:         10,
			},
			expectedCount: 6,
			expectedFirst: "2025-04-01T15:44:44.534732559Z",
			expectedToken: true,
		},
		{
			name: "limit logs to 2",
			config: api.Config{
				Namespace:     "nullplatform",
				ApplicationID: "1691688910",
				ScopeID:       "760499159",
				DeploymentID:  "1705961777",
				Limit:         2,
			},
			expectedCount: 2,
			expectedFirst: "2025-04-01T15:44:44.534732559Z",
			expectedToken: true,
		},
		{
			name: "filter logs",
			config: api.Config{
				Namespace:     "nullplatform",
				ApplicationID: "1691688910",
				ScopeID:       "760499159",
				DeploymentID:  "1705961777",
				Limit:         10,
				FilterPattern: "CPU",
			},
			expectedCount: 1,
			expectedFirst: "2025-04-01T15:44:44.548275040Z",
			expectedToken: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := fetcher.FetchLogs(tc.config)
			if err != nil {
				t.Fatalf("FetchLogs returned error: %v", err)
			}

			if len(result.Results) != tc.expectedCount {
				t.Errorf("Expected %d logs, got %d", tc.expectedCount, len(result.Results))
			}

			if len(result.Results) > 0 {
				if !contains(result.Results[0].Message, tc.expectedFirst) {
					t.Errorf("First log doesn't match expected. Got: %s", result.Results[0].Message)
				}
			}

			if tc.expectedToken && result.NextPageToken == "" {
				t.Errorf("Expected non-empty next page token, but got empty")
			}
		})
	}
}

func TestProcessLogLines(t *testing.T) {
	testCases := []struct {
		name           string
		logs           string
		filter         string
		podName        string
		expectedCount  int
		expectedLastTS string
	}{
		{
			name:           "process simple logs",
			logs:           "2025-04-01T15:44:44.534Z Line 1\n2025-04-01T15:44:45.123Z Line 2\n",
			filter:         "",
			podName:        "pod-1",
			expectedCount:  2,
			expectedLastTS: "2025-04-01T15:44:45.123Z",
		},
		{
			name:           "filter logs",
			logs:           "2025-04-01T15:44:44.534Z Error line\n2025-04-01T15:44:45.123Z Info line\n",
			filter:         "Error",
			podName:        "pod-1",
			expectedCount:  1,
			expectedLastTS: "2025-04-01T15:44:44.534Z",
		},
		{
			name:           "empty logs",
			logs:           "",
			filter:         "",
			podName:        "pod-1",
			expectedCount:  0,
			expectedLastTS: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logs, lastTS, err := processLogLines(tc.logs, tc.filter, tc.podName)
			if err != nil {
				t.Fatalf("processLogLines returned error: %v", err)
			}

			if len(logs) != tc.expectedCount {
				t.Errorf("Expected %d logs, got %d", tc.expectedCount, len(logs))
			}

			if lastTS != tc.expectedLastTS {
				t.Errorf("Expected last timestamp %s, got %s", tc.expectedLastTS, lastTS)
			}
		})
	}
}

func TestBuildLabels(t *testing.T) {
	testCases := []struct {
		name     string
		config   api.Config
		expected map[string]string
	}{
		{
			name: "all fields",
			config: api.Config{
				ApplicationID: "12345",
				ScopeID:       "67890",
				DeploymentID:  "54321",
			},
			expected: map[string]string{
				"nullplatform":   "true",
				"application_id": "12345",
				"scope_id":       "67890",
				"deployment_id":  "54321",
			},
		},
		{
			name: "partial fields",
			config: api.Config{
				ApplicationID: "12345",
			},
			expected: map[string]string{
				"nullplatform":   "true",
				"application_id": "12345",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			labels := buildLabels(tc.config)
			if !reflect.DeepEqual(labels, tc.expected) {
				t.Errorf("Expected labels %v, got %v", tc.expected, labels)
			}
		})
	}
}

func contains(s, substr string) bool {
	return s != "" && substr != "" && s[0:len(substr)] == substr
}
