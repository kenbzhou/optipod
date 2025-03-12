package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"io/ioutil"
)

// Extender request structure from Kubernetes
type ExtenderArgs struct {
	Nodes       *NodeList   `json:"nodes"`
	Pod         *Pod        `json:"pod"`
	NodeNames   *[]string   `json:"nodenames,omitempty"`
}

// Node list
type NodeList struct {
	Items []Node `json:"items"`
}

// Node information
type Node struct {
	Metadata NodeMetadata `json:"metadata"`
}

// Node metadata
type NodeMetadata struct {
	Name string `json:"name"`
}

// Pod information
type Pod struct {
	Metadata PodMetadata `json:"metadata"`
}

// Pod metadata
type PodMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Priority structure for responses
type HostPriority struct {
	Host  string `json:"host"`
	Score int    `json:"score"`
}

// Response structure for the extender
type ExtenderFilterResult struct {
	NodeNames  []string `json:"nodenames,omitempty"`
	NodeCounts int      `json:"nodescount,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// PriorityResponse for returning scores
type PriorityResponse struct {
	HostPriorities []HostPriority `json:"hostPriorities"`
}

// fetchMetrics queries Prometheus for metrics
func fetchMetrics(nodeName string) (float64, error) {
	url := fmt.Sprintf("http://orchestrator-service.default.svc.cluster.local:9090/api/v1/query?query=mem_bytes_allocated{node_id=\"%s\"}", nodeName)
	client := http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %v", err)
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
		return 0, fmt.Errorf("failed to parse JSON: %v", err)
	}
	
	if len(result.Data.Result) == 0 {
		return 0, fmt.Errorf("no data found for node %s", nodeName)
	}

	// The value is typically [timestamp, value_string]
	if len(result.Data.Result[0].Value) < 2 {
		return 0, fmt.Errorf("invalid response format")
	}
	
	valueStr, ok := result.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("value is not a string")
	}
	
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse value: %v", err)
	}

	return value, nil
}

// prioritize nodes based on metrics
func prioritize(w http.ResponseWriter, r *http.Request) {
	var extenderArgs ExtenderArgs
	var hostPriorities []HostPriority

	// Parse the request body
	if err := json.NewDecoder(r.Body).Decode(&extenderArgs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if extenderArgs.Nodes == nil || len(extenderArgs.Nodes.Items) == 0 {
		http.Error(w, "empty node list", http.StatusBadRequest)
		return
	}

	log.Printf("Prioritizing for pod %s/%s", 
		extenderArgs.Pod.Metadata.Namespace, 
		extenderArgs.Pod.Metadata.Name)

	// Score each node
	for _, node := range extenderArgs.Nodes.Items {
		nodeName := node.Metadata.Name
		memValue, err := fetchMetrics(nodeName)
		
		var score int
		if err != nil {
			log.Printf("Warning: Error fetching metrics for node %s: %v", nodeName, err)
			score = 50 // Neutral score on error
		} else {
			// Convert to MB and score inversely (lower mem usage = higher score)
			memValueMB := memValue / 1024 / 1024
			score = 100 - int(memValueMB)
			if score < 0 {
				score = 0
			}
			if score > 100 {
				score = 100
			}
		}
		
		hostPriorities = append(hostPriorities, HostPriority{
			Host:  nodeName,
			Score: score,
		})
		
		log.Printf("Node %s scored %d", nodeName, score)
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PriorityResponse{
		HostPriorities: hostPriorities,
	})
}

func main() {
	http.HandleFunc("/prioritize", prioritize)
	log.Println("Starting scheduler extender on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}