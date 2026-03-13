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

type DirectConnectionParameters struct {
	Uri string `yaml:"uri" json:"uri"`
}

type K8sConnectionParameters struct {
	Context   string `yaml:"context" json:"context"`
	Namespace string `yaml:"namespace" json:"namespace"`
	Name      string `yaml:"name" json:"name"`
}

func (p ConnectionParameters) String() string {
	if p.Direct != nil {
		return p.Direct.String()
	} else if p.K8s != nil {
		return p.K8s.String()
	}

	return ""
}

func (p *DirectConnectionParameters) String() string {
	return fmt.Sprintf("Direct connection to %s", p.Uri)
}

func (p *K8sConnectionParameters) String() string {
	return fmt.Sprintf("K8s connection, context %s, namespace %s, instance %s", p.Context, p.Namespace, p.Name)
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
