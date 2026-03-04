package view

import (
	"tbunny/internal/cluster"
	"tbunny/internal/model"
)

type ClusterAwareResourceTableView[R Resource] struct {
	*ResourceTableView[R]

	cluster *cluster.Cluster
}

func NewClusterAwareResourceTableView[R Resource](name string, strategy UpdateStrategy) *ClusterAwareResourceTableView[R] {
	v := ClusterAwareResourceTableView[R]{
		ResourceTableView: NewResourceTableView[R](name, strategy),
	}

	return &v
}

func (v *ClusterAwareResourceTableView[R]) Cluster() *cluster.Cluster {
	return v.cluster
}

func (v *ClusterAwareResourceTableView[R]) Init(app model.App) (err error) {
	err = v.ResourceTableView.Init(app)
	if err != nil {
		return err
	}

	v.cluster = cluster.Current()

	return nil
}

func (v *ClusterAwareResourceTableView[R]) Start() {
	if v.cluster == nil || !v.cluster.IsAvailable() {
		v.Strategy().Pause()
	}

	v.ResourceTableView.Start()

	if v.cluster != nil {
		v.cluster.AddActiveVirtualHostListener(v)
		v.cluster.AddVirtualHostsListener(v)
		v.cluster.AddConnectionListener(v)
	}
}

func (v *ClusterAwareResourceTableView[R]) Stop() {
	if v.cluster != nil {
		v.cluster.RemoveActiveVirtualHostListener(v)
		v.cluster.RemoveVirtualHostsListener(v)
		v.cluster.RemoveConnectionListener(v)
	}

	v.ResourceTableView.Stop()
}

func (v *ClusterAwareResourceTableView[R]) ClusterActiveVirtualHostChanged(*cluster.Cluster) {
	v.RequestUpdate(FullUpdate)
}

func (v *ClusterAwareResourceTableView[R]) ClusterVirtualHostsChanged(*cluster.Cluster) {}

func (v *ClusterAwareResourceTableView[R]) ClusterConnectionLost(*cluster.Cluster) {
	v.Strategy().Pause()
	v.RefreshActions()
}

func (v *ClusterAwareResourceTableView[R]) ClusterConnectionRestored(*cluster.Cluster) {
	v.Strategy().Resume()
	v.RefreshActions()
}
