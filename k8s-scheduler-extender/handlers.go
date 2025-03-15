package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// DUMMY TEMPLATE FOR NOW.

// ExtenderHandler handles the HTTP requests for the scheduler extender
type ExtenderHandler struct{}

func NewExtenderHandler() *ExtenderHandler {
	return &ExtenderHandler{}
}

// ExtenderArgs represents the arguments sent to the extender from the scheduler
type ExtenderArgs struct {
	Pod       v1.Pod       `json:"pod"`
	Nodes     *v1.NodeList `json:"nodes,omitempty"`
	NodeNames *[]string    `json:"nodenames,omitempty"`
}

// ExtenderFilterResult represents the result returned to the scheduler after filtering
type ExtenderFilterResult struct {
	Nodes       *v1.NodeList      `json:"nodes,omitempty"`
	NodeNames   *[]string         `json:"nodenames,omitempty"`
	FailedNodes map[string]string `json:"failedNodes,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// HostPriority represents the priority assigned to a node
type HostPriority struct {
	Host  string `json:"host"`
	Score int64  `json:"score"`
}

// ExtenderPriorityResult represents the result returned to the scheduler after prioritizing
type ExtenderPriorityResult struct {
	HostPriorityList []HostPriority `json:"hostPriorityList"`
	Error            string         `json:"error,omitempty"`
}

// NodeMetrics holds the data from a Prom request
type NodeMetrics struct {
	MemBytesAllocated   float64
	PageFaults          float64
	CtxSwitchesGraceful float64
	CtxSwitchesForced   float64
	FsReadCount         float64
	FsReadSizeKb        float64
	FsWriteCount        float64
	FsWriteSizeKb       float64
}

// Maps Node_ID to NodeMetrics
type MetricsResult map[string]NodeMetrics

type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				NodeID string `json:"node_id"`
			} `json:"metric"`
			Value []interface{} `json:"value"` // [timestamp, value]
		} `json:"result"`
	} `json:"data"`
}

// Captures metric avgs across all nodes
type MetricAverages struct {
	MemBytesAllocated   float64
	PageFaults          float64
	CtxSwitchesGraceful float64
	CtxSwitchesForced   float64
	FsReadCount         float64
	FsReadSizeKb        float64
	FsWriteCount        float64
	FsWriteSizeKb       float64
}

// FetchNodeMetrics fetches metrics for all nodes in the ExtenderArgs
// Returns each node's metrics and their avgs
func FetchNodeMetrics(args ExtenderArgs) (MetricsResult, MetricAverages, error) {
	// Collect node IDs from either Nodes or NodeNames
	nodeIDs := []string{}
	if args.Nodes != nil {
		for _, node := range args.Nodes.Items {
			nodeIDs = append(nodeIDs, node.Name)
		}
	} else if args.NodeNames != nil {
		nodeIDs = *args.NodeNames
	} else {
		return nil, MetricAverages{}, fmt.Errorf("no nodes provided in args")
	}

	if len(nodeIDs) == 0 {
		return nil, MetricAverages{}, fmt.Errorf("empty node list")
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Define metrics to fetch
	metrics := []string{
		"mem_bytes_allocated",
		"page_faults",
		"ctx_switches_graceful",
		"ctx_switches_forced",
		"fs_read_count",
		"fs_read_size_kb",
		"fs_write_count",
		"fs_write_size_kb",
	}

	result := make(MetricsResult)
	for _, nodeID := range nodeIDs {
		result[nodeID] = NodeMetrics{}
	}

	metricSums := MetricAverages{}
	metricCounts := map[string]int{
		"mem_bytes_allocated":   0,
		"page_faults":           0,
		"ctx_switches_graceful": 0,
		"ctx_switches_forced":   0,
		"fs_read_count":         0,
		"fs_read_size_kb":       0,
		"fs_write_count":        0,
		"fs_write_size_kb":      0,
	}

	// Build the node selector for Prometheus query
	nodeSelector := strings.Join(nodeIDs, "|")
	nodeSelector = fmt.Sprintf(`node_id=~"%s"`, nodeSelector)

	// Fetch each metric
	for _, metric := range metrics {
		query := fmt.Sprintf("%s{%s}", metric, nodeSelector)
		encodedQuery := url.QueryEscape(query)
		url := fmt.Sprintf("http://orchestrator-service.default.svc.cluster.local:9090/api/v1/query?query=%s", encodedQuery)
		resp, err := client.Get(url)
		if err != nil {
			klog.Errorf("Error fetching metric %s: %v", metric, err)
			continue
		}

		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			klog.Errorf("Error reading response for metric %s: %v", metric, err)
			continue
		}

		// Parse the response
		var promResponse PrometheusResponse
		if err := json.Unmarshal(body, &promResponse); err != nil {
			klog.Errorf("Error parsing response for metric %s: %v", metric, err)
			continue
		}

		// Check if status is success
		if promResponse.Status != "success" {
			klog.Errorf("Prometheus returned non-success status for metric %s: %s", metric, promResponse.Status)
			continue
		}

		// Process query results
		for _, res := range promResponse.Data.Result {
			nodeID := res.Metric.NodeID
			// Extract val
			var metricValue float64
			if len(res.Value) >= 2 {
				switch v := res.Value[1].(type) {
				case string:
					fmt.Sscanf(v, "%f", &metricValue)
				case float64:
					metricValue = v
				}
			}

			metricCounts[metric]++

			// Populate results
			if nodeMetrics, ok := result[nodeID]; ok {
				switch metric {
				case "mem_bytes_allocated":
					nodeMetrics.MemBytesAllocated = metricValue
					metricSums.MemBytesAllocated += metricValue
				case "page_faults":
					nodeMetrics.PageFaults = metricValue
					metricSums.PageFaults += metricValue
				case "ctx_switches_graceful":
					nodeMetrics.CtxSwitchesGraceful = metricValue
					metricSums.CtxSwitchesGraceful += metricValue
				case "ctx_switches_forced":
					nodeMetrics.CtxSwitchesForced = metricValue
					metricSums.CtxSwitchesForced += metricValue
				case "fs_read_count":
					nodeMetrics.FsReadCount = metricValue
					metricSums.FsReadCount += metricValue
				case "fs_read_size_kb":
					nodeMetrics.FsReadSizeKb = metricValue
					metricSums.FsReadSizeKb += metricValue
				case "fs_write_count":
					nodeMetrics.FsWriteCount = metricValue
					metricSums.FsWriteCount += metricValue
				case "fs_write_size_kb":
					nodeMetrics.FsWriteSizeKb = metricValue
					metricSums.FsWriteSizeKb += metricValue
				}
				result[nodeID] = nodeMetrics
			}
		}
	}

	// Calculate averages
	averages := MetricAverages{}

	if count := metricCounts["mem_bytes_allocated"]; count > 0 {
		averages.MemBytesAllocated = metricSums.MemBytesAllocated / float64(count)
	}
	if count := metricCounts["page_faults"]; count > 0 {
		averages.PageFaults = metricSums.PageFaults / float64(count)
	}
	if count := metricCounts["ctx_switches_graceful"]; count > 0 {
		averages.CtxSwitchesGraceful = metricSums.CtxSwitchesGraceful / float64(count)
	}
	if count := metricCounts["ctx_switches_forced"]; count > 0 {
		averages.CtxSwitchesForced = metricSums.CtxSwitchesForced / float64(count)
	}
	if count := metricCounts["fs_read_count"]; count > 0 {
		averages.FsReadCount = metricSums.FsReadCount / float64(count)
	}
	if count := metricCounts["fs_read_size_kb"]; count > 0 {
		averages.FsReadSizeKb = metricSums.FsReadSizeKb / float64(count)
	}
	if count := metricCounts["fs_write_count"]; count > 0 {
		averages.FsWriteCount = metricSums.FsWriteCount / float64(count)
	}
	if count := metricCounts["fs_write_size_kb"]; count > 0 {
		averages.FsWriteSizeKb = metricSums.FsWriteSizeKb / float64(count)
	}

	return result, averages, nil
}

// Filters nodes based on metrics from the Prometheus orchestrator
func (h *ExtenderHandler) Filter(w http.ResponseWriter, r *http.Request) {
	klog.V(4).Infof("Received filter request")

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	// Parse the request
	var args ExtenderArgs
	if err := json.Unmarshal(body, &args); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unmarshal request: %v", err), http.StatusBadRequest)
		return
	}

	// We don't have any specific behavior we want to implement in filtering. Specific implementation plan emphasizes scoring based on low level
	// metrics, but the native scheduler should do a good enough job of filtering for available candidates.

	// "Pass through"
	filterResult := ExtenderFilterResult{
		Nodes:       args.Nodes,
		NodeNames:   args.NodeNames,
		FailedNodes: make(map[string]string),
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filterResult); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// Normalizer function for different average ratios.
// Takes in a ratio and some cutoff values. MinCutoff should be generally high since we're indicating some minimum bound
// where resource scarcity is essentially nnexistent.
func normalizeRatio(ratio, minCutoff, maxCutoff float64) float64 {
	if ratio <= minCutoff {
		return 0.0
	}
	if ratio >= maxCutoff {
		return 1.0
	}

	// Linear scaling between cutoffs
	return (ratio - minCutoff) / (maxCutoff - minCutoff)
}

func calculateScore(nodeMetrics NodeMetrics, averages MetricAverages) int64 {
	var memoryScore, cpuScore, fsScore float64

	// 1. Memory metrics (40%)
	// Malloc
	var memBytesScore float64
	if averages.MemBytesAllocated > 0 {
		memoryRatio := nodeMetrics.MemBytesAllocated / averages.MemBytesAllocated
		memBytesScore = 100 * (1 - normalizeRatio(memoryRatio, 0.5, 2.0))
	} else {
		memBytesScore = 50
	}

	// Page faults
	var pageFaultScore float64
	if averages.PageFaults > 0 {
		pageFaultRatio := nodeMetrics.PageFaults / averages.PageFaults
		pageFaultScore = 100 * (1 - normalizeRatio(pageFaultRatio, 0.5, 2.0))
	} else {
		pageFaultScore = 50
	}

	// Weighted combination of memory metrics (memory bytes is more important than page faults)
	memoryScore = (memBytesScore * 0.6) + (pageFaultScore * 0.4)

	// 2. CPU metrics (40%)
	// ctx switches - lower is better for both graceful and forced
	var gracefulCtxSwitchScore, forcedCtxSwitchScore float64

	if averages.CtxSwitchesGraceful > 0 {
		gracefulRatio := nodeMetrics.CtxSwitchesGraceful / averages.CtxSwitchesGraceful
		gracefulCtxSwitchScore = 100 * (1 - normalizeRatio(gracefulRatio, 0.5, 2.0))
	} else {
		gracefulCtxSwitchScore = 50
	}

	if averages.CtxSwitchesForced > 0 {
		forcedRatio := nodeMetrics.CtxSwitchesForced / averages.CtxSwitchesForced
		forcedCtxSwitchScore = 100 * (1 - normalizeRatio(forcedRatio, 0.5, 2.0))
	} else {
		forcedCtxSwitchScore = 50
	}

	// reflect impact of force switches
	cpuScore = (gracefulCtxSwitchScore * 0.2) + (forcedCtxSwitchScore * 0.8)

	// 3. fs metrics (20%)
	var readCountScore, readSizeScore, writeCountScore, writeSizeScore float64

	if averages.FsReadCount > 0 {
		readCountRatio := nodeMetrics.FsReadCount / averages.FsReadCount
		readCountScore = 100 * (1 - normalizeRatio(readCountRatio, 0.5, 2.0))
	} else {
		readCountScore = 50
	}

	if averages.FsReadSizeKb > 0 {
		readSizeRatio := nodeMetrics.FsReadSizeKb / averages.FsReadSizeKb
		readSizeScore = 100 * (1 - normalizeRatio(readSizeRatio, 0.5, 2.0))
	} else {
		readSizeScore = 50
	}

	if averages.FsWriteCount > 0 {
		writeCountRatio := nodeMetrics.FsWriteCount / averages.FsWriteCount
		writeCountScore = 100 * (1 - normalizeRatio(writeCountRatio, 0.5, 2.0))
	} else {
		writeCountScore = 50
	}

	if averages.FsWriteSizeKb > 0 {
		writeSizeRatio := nodeMetrics.FsWriteSizeKb / averages.FsWriteSizeKb
		writeSizeScore = 100 * (1 - normalizeRatio(writeSizeRatio, 0.5, 2.0))
	} else {
		writeSizeScore = 50
	}

	// Writes more impactful than reads
	fsReadScore := (readCountScore * 0.4) + (readSizeScore * 0.6)
	fsWriteScore := (writeCountScore * 0.4) + (writeSizeScore * 0.6)
	fsScore = (fsReadScore * 0.3) + (fsWriteScore * 0.7)

	// 4. Calculate final weighted score
	// Memory: 40%, CPU: 40%, Filesystem: 20%
	finalScore := (memoryScore * 0.4) + (cpuScore * 0.4) + (fsScore * 0.2)

	// Ensure score is within 0-100 range
	if finalScore > 100 {
		finalScore = 100
	}
	if finalScore < 0 {
		finalScore = 0
	}

	return int64(finalScore)
}

// main driver of our scheduler
func (h *ExtenderHandler) Prioritize(w http.ResponseWriter, r *http.Request) {
	klog.V(4).Infof("Received prioritize request")

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	// Parse the request
	var args ExtenderArgs
	if err := json.Unmarshal(body, &args); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unmarshal request: %v", err), http.StatusBadRequest)
		return
	}

	metricMap, avgs, errCode := FetchNodeMetrics(args)
	if errCode != nil {
		http.Error(w, fmt.Sprintf("Failed to map node data to field: %v", errCode), http.StatusBadRequest)
		return
	}

	// Initialize host priority list
	hostPriorityList := make([]HostPriority, 0)

	for nodeId, metrics := range metricMap {
		score := calculateScore(metrics, avgs)
		hostPriorityList = append(hostPriorityList, HostPriority{
			Host:  nodeId,
			Score: score,
		})
	}

	// Create priority result
	priorityResult := ExtenderPriorityResult{
		HostPriorityList: hostPriorityList,
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(priorityResult); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
