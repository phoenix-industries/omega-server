// Package kernel manages the core dependencies and services.
package kernel

import (
	"context"
	"fmt"
)

type Service interface {
	Name() string
	Register(ctx context.Context, env *Env) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Kernel struct {
	env      *Env
	services map[string]Service
}

func NewKernel(env *Env) *Kernel {
	return &Kernel{
		env:      env,
		services: map[string]Service{},
	}
}

func (k *Kernel) Env() *Env {
	return k.env
}

func (k *Kernel) Register(ctx context.Context, services ...Service) error {
	for _, service := range services {
		name := service.Name()
		if _, ok := k.services[name]; ok {
			return fmt.Errorf("service already registered: %s", name)
		}
		if err := service.Register(ctx, k.env); err != nil {
			return fmt.Errorf("failed to register service %s: %w", name, err)
		}
		k.services[name] = service
		k.env.logger.DebugContext(ctx, "service registered", "name", name)
	}
	return nil
}

func (k *Kernel) Start(ctx context.Context) error {
	for name, service := range k.services {
		if err := service.Start(ctx); err != nil {
			return fmt.Errorf("failed to start service %s: %w", name, err)
		}
		k.env.logger.DebugContext(ctx, "service started", "name", name)
	}
	return nil
}

func (k *Kernel) Stop(ctx context.Context) error {
	for _, service := range k.services {
		if err := service.Stop(ctx); err != nil {
			return err
		}
	}
	return nil
}
