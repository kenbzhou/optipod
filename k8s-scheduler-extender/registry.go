// K8s registry 

package main

import (
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

// RegisterPlugins registers the custom plugins.
func RegisterPlugins(registry runtime.Registry) error {
	// Register the DummyScorer plugin.
	if err := registry.Register("DummyScorer", func(configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
		return NewDummyScorer(configuration, f)
	}); err != nil {
		return err
	}

	// Register the CustomScorer plugin.
	if err := registry.Register("CustomScorer", func(configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
		return NewCustomScorer(configuration, f)
	}); err != nil {
		return err
	}

	return nil
}