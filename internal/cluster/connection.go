package cluster

type connectionListener interface {
	ConnectionUriChanged(uri string)
}

type connection interface {
	Uri() string
	AddListener(l connectionListener)
	Close()
}
