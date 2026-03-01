package cluster

import "fmt"

type DirectConnectionParameters struct {
	Uri string `yaml:"uri" json:"uri"`
}

type directConnection struct {
	parameters *DirectConnectionParameters
}

func newDirectConnection(p *DirectConnectionParameters) connection {
	return &directConnection{p}
}

func (c *directConnection) Uri() string {
	return c.parameters.Uri
}

func (c *directConnection) Close() {}

func (p *DirectConnectionParameters) String() string {
	return fmt.Sprintf("Direct connection to %s", p.Uri)
}
