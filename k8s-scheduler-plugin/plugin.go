package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// CustomScorer implements the ScorePlugin interface
type CustomScorer struct {
	handle framework.Handle
}

var _ framework.ScorePlugin = &CustomScorer{}

// Name returns the name of the plugin
func (c *CustomScorer) Name() string {
	return "CustomScorer"
}

// assigns a score to nodes based on Prometheus data
func (c *CustomScorer) Score(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeName string) (int64, *framework.Status) {
	klog.V(2).InfoS("Scoring node based on memory metrics", "node", nodeName, "pod", pod.Name)
	
	metrics, err := fetchMetrics(nodeName)
	if err != nil {
		klog.ErrorS(err, "Failed to fetch metrics", "node", nodeName)
		return 50, framework.NewStatus(framework.Success) // Return a neutral score on error
	}

	// TODO: incorporate all metrics in score, right now its only prioritizes nodes with lower memory usage
	memUsage := metrics["mem_bytes_allocated"]
	memUsageInMB := memUsage / 1024 / 1024 // B -> MB
	
	// Calculate score - lower memory usage is better
	score := 100 - int64(memUsageInMB)
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	klog.V(2).InfoS("Final score for node", 
		"node", nodeName, 
		"score", score, 
		"memoryUsage(MB)", memUsageInMB)
	
	return score, framework.NewStatus(framework.Success)
}

// returns the score extension object
func (c *CustomScorer) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// fetchMetrics queries Prometheus for the given node's memory metrics
// TODO: debug this, it's not working as expected
func fetchMetrics(nodeName string) (map[string]float64, error) {
	url := fmt.Sprintf("http://orchestrator-service.default.svc.cluster.local:9090/api/v1/query?query=mem_bytes_allocated{node_id=\"%s\"}", nodeName)
	client := http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Value []interface{} `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}
	
	if len(result.Data.Result) == 0 {
		return nil, fmt.Errorf("no data found for node %s", nodeName)
	}

	// TODO: is [timestamp, value_string] the format here? 
	if len(result.Data.Result[0].Value) < 2 {
		return nil, fmt.Errorf("invalid response format")
	}
	
	valueStr, ok := result.Data.Result[0].Value[1].(string)
	if !ok {
		return nil, fmt.Errorf("value is not a string")
	}
	
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse value: %v", err)
	}

	return map[string]float64{"mem_bytes_allocated": value}, nil
}

// New initializes a new plugin and returns it
func New(configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
	klog.V(2).InfoS("Creating CustomScorer plugin")
	return &CustomScorer{
		handle: f,
	}, nil
}