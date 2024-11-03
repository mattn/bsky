package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Controller interface {
	// ReconcileNode reconciles the state of the resource.
	ReconcileNode(ctx context.Context, node *yaml.RNode) error
}

// Registry is a registry of all the controllers.
// It is used to map a kind to the controller that can handle that kind.
// We need to avoid circular dependencies.
// So controller registration happens in app.SetupRegistry. The app package can then import
//  1. Any package which defines a controller
//  2. The controller package itself
//
// The controller package shouldn't import any controller packages as that would risk circular dependencies
// Controllers that need to apply other resources can import the registry in order to loop up the controller for the resource
type Registry struct {
	controllers map[schema.GroupVersionKind]Controller
}

func (r *Registry) Register(gvk schema.GroupVersionKind, controller Controller) error {
	log := zapr.NewLogger(zap.L())
	if r.controllers == nil {
		r.controllers = make(map[schema.GroupVersionKind]Controller)
	}

	if _, ok := r.controllers[gvk]; ok {
		return fmt.Errorf("controller already registered for %v", gvk)
	}
	log.Info("Registering controller", "gvk", gvk)
	r.controllers[gvk] = controller
	return nil
}

func (r *Registry) GetController(gvk schema.GroupVersionKind) (Controller, error) {
	controller, ok := r.controllers[gvk]
	if !ok {
		return nil, fmt.Errorf("No controller registered for %v", gvk)
	}
	return controller, nil
}

func (r *Registry) ReconcileNode(ctx context.Context, n *yaml.RNode) error {
	m, err := n.GetMeta()
	if err != nil {
		return errors.Wrapf(err, "Failed to get metadata from node")
	}
	gvk := schema.FromAPIVersionAndKind(m.APIVersion, m.Kind)
	controller, err := r.GetController(gvk)

	if err != nil {
		return err
	}
	return controller.ReconcileNode(ctx, n)
}
