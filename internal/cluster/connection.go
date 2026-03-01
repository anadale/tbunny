package cluster

import (
	"context"
	"fmt"
)

type ConnectionParameters struct {
	Uri      string                      `yaml:"uri,omitempty" json:"uri,omitempty"`
	Direct   *DirectConnectionParameters `yaml:"direct,omitempty" json:"direct,omitempty"`
	K8s      *K8sConnectionParameters    `yaml:"k8s,omitempty" json:"k8s,omitempty"`
	Username string                      `yaml:"username" json:"username"`
	Password string                      `yaml:"password" json:"password"`
}

type connection interface {
	Uri() string

	Close()
}

func (p ConnectionParameters) String() string {
	if p.Direct != nil {
		return p.Direct.String()
	} else if p.K8s != nil {
		return p.K8s.String()
	}

	return ""
}

func (p ConnectionParameters) createConnection(ctx context.Context) (connection, error) {
	if p.Direct != nil {
		return newDirectConnection(p.Direct), nil
	} else if p.K8s != nil {
		return newK8sConnection(ctx, p.K8s)
	}

	return nil, fmt.Errorf("no connection parameters provided")
}

func (p ConnectionParameters) migrate() ConnectionParameters {
	if p.Uri != "" {
		return ConnectionParameters{
			Direct: &DirectConnectionParameters{
				Uri: p.Uri,
			},
			Username: p.Username,
			Password: p.Password,
		}
	}

	return p
}
