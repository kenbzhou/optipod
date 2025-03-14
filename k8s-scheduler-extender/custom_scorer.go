package main

import (
	"context"
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// CustomScorer is a plugin that scores nodes based on external data.
type CustomScorer struct {
	handle framework.Handle
}

var _ framework.ScorePlugin = &CustomScorer{}

// Name returns the name of the plugin.
func (c *CustomScorer) Name() string {
	return "CustomScorer"
}

// Score assigns a score to the node based on external data.
func (c *CustomScorer) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	// Fetch external data (mock for now).
	externalData := fetchExternalData(nodeName)

	// Calculate score based on external data.
	score := calculateScore(externalData)
	fmt.Printf("Node %s scored %d by CustomScorer based on external data: %v\n", nodeName, score, externalData)
	return score, nil
}

// ScoreExtensions returns nil because we don't need to normalize scores.
func (c *CustomScorer) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// NewCustomScorer initializes the plugin.
func NewCustomScorer(_ runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &CustomScorer{handle: handle}, nil
}

// fetchExternalData mocks fetching external data for a node.
func fetchExternalData(nodeName string) map[string]int64 {
	// Mock data: e.g., CPU usage, memory availability, etc.
	return map[string]int64{
		"cpu_usage":    70, // Example: 70% CPU usage
		"memory_avail": 30, // Example: 30% memory available
	}
}

// calculateScore calculates a score based on external data.
func calculateScore(data map[string]int64) int64 {
	// Example scoring logic: prioritize nodes with lower CPU usage and higher memory availability.
	cpuUsage := data["cpu_usage"]
	memoryAvail := data["memory_avail"]
	score := (100 - cpuUsage) + memoryAvail // Higher score is better
	return score
}