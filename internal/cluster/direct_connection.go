package cluster

type directConnection struct {
	parameters *DirectConnectionParameters
}

func newDirectConnection(p *DirectConnectionParameters) connection {
	return &directConnection{p}
}

func (c *directConnection) Uri() string {
	return c.parameters.Uri
}

func (c *directConnection) AddListener(connectionListener) {}

func (c *directConnection) Close() {}
