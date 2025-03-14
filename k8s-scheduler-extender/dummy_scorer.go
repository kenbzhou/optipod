package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// DummyScorer is a plugin that assigns a random score to each node.
type DummyScorer struct{}

var _ framework.ScorePlugin = &DummyScorer{}

// Name returns the name of the plugin.
func (d *DummyScorer) Name() string {
	return "DummyScorer"
}

// Score assigns a random score to the node.
func (d *DummyScorer) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	// Assign a random score between 0 and 100.
	rand.Seed(time.Now().UnixNano())
	score := rand.Int63n(100)
	fmt.Printf("Node %s scored %d by DummyScorer\n", nodeName, score)
	return score, nil
}

// ScoreExtensions returns nil because we don't need to normalize scores.
func (d *DummyScorer) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// NewDummyScorer initializes the plugin.
func NewDummyScorer(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &DummyScorer{}, nil
}