package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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

// Filter filters nodes based on metrics from the Prometheus orchestrator
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

	// Initialize the filter result
	filterResult := ExtenderFilterResult{
		Nodes:       args.Nodes,
		NodeNames:   args.NodeNames,
		FailedNodes: make(map[string]string),
	}

	// For simplicity, accept all nodes for now
	// TODO: Implement metrics-based filtering using orchestrator data

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filterResult); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// Prioritize assigns priority to nodes based on metrics from the Prometheus orchestrator
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

	// Initialize host priority list
	hostPriorityList := make([]HostPriority, 0)

	// Process node list if available
	if args.Nodes != nil {
		for _, node := range args.Nodes.Items {
			// TODO: Implement metrics-based scoring using orchestrator data
			// For simplicity, assign a default score of 50 to all nodes
			hostPriorityList = append(hostPriorityList, HostPriority{
				Host:  node.Name,
				Score: 50, // Default score
			})
		}
	}

	// Process node names if available
	if args.NodeNames != nil {
		for _, nodeName := range *args.NodeNames {
			// TODO: Implement metrics-based scoring using orchestrator data
			// For simplicity, assign a default score of 50 to all nodes
			hostPriorityList = append(hostPriorityList, HostPriority{
				Host:  nodeName,
				Score: 50, // Default score
			})
		}
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
