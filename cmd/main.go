package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/nullplatform/kube-logger/pkg/api"
	"github.com/nullplatform/kube-logger/pkg/client"
	"github.com/nullplatform/kube-logger/pkg/fetcher"
)

func main() {
	config := parseFlags()

	kubeClient, err := client.NewClient()
	if err != nil {
		exitWithError("Failed to create Kubernetes client", err)
	}

	logFetcher := fetcher.New(kubeClient)
	result, err := logFetcher.FetchLogs(config)
	if err != nil {
		exitWithError("Failed to fetch logs", err)
	}

	err = printResult(result)
	if err != nil {
		exitWithError("Failed to print results", err)
	}

	// Make sure output is flushed
	os.Stdout.Sync()
	return
}

func parseFlags() api.Config {
	var config api.Config

	flag.StringVar(&config.Namespace, "namespace", "", "Kubernetes namespace")
	flag.StringVar(&config.ApplicationID, "application-id", "", "Application ID")
	flag.StringVar(&config.ScopeID, "scope-id", "", "Scope ID")
	flag.StringVar(&config.DeploymentID, "deployment-id", "", "Deployment ID")
	flag.IntVar(&config.Limit, "limit", 100, "Maximum number of log entries")
	flag.StringVar(&config.NextPageToken, "next-page-token", "", "Token for pagination")
	flag.StringVar(&config.FilterPattern, "filter", "", "Filter pattern")
	flag.StringVar(&config.StartTime, "start-time", "", "Start time for logs (ISO format)")

	flag.Parse()

	return config
}

func exitWithError(message string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", message, err)
	os.Exit(1)
}

func printResult(result *api.Result) error {
	output, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal results to JSON: %w", err)
	}

	_, err = os.Stdout.Write(output)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	os.Stdout.Sync()
	return nil
}
