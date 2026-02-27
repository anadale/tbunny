package view

import (
	"tbunny/internal/cluster"
	"tbunny/internal/model"
)

type ClusterAwareRefreshableView[U UiComponent] struct {
	*RefreshableView[U]

	cluster *cluster.Cluster
}

func NewClusterAwareRefreshableView[U UiComponent](name string, ui U, strategy UpdateStrategy) *ClusterAwareRefreshableView[U] {
	v := ClusterAwareRefreshableView[U]{
		RefreshableView: NewRefreshableView[U](name, ui, strategy),
	}

	return &v
}

func (v *ClusterAwareRefreshableView[U]) Cluster() *cluster.Cluster {
	return v.cluster
}

func (v *ClusterAwareRefreshableView[U]) Init(app model.App) (err error) {
	err = v.RefreshableView.Init(app)
	if err != nil {
		return err
	}

	v.cluster = cluster.Current()
	if v.cluster == nil {
		panic("cluster expected not to be nil")
	}

	return nil
}

func (v *ClusterAwareRefreshableView[U]) Start() {
	if !v.cluster.IsAvailable() {
		v.strategy.Pause()
	}

	v.RefreshableView.Start()
	v.cluster.AddListener(v)
}

func (v *ClusterAwareRefreshableView[U]) Stop() {
	v.cluster.RemoveListener(v)
	v.RefreshableView.Stop()
}

func (v *ClusterAwareRefreshableView[U]) ClusterActiveVirtualHostChanged(*cluster.Cluster) {
	v.RequestUpdate(FullUpdate)
}

func (v *ClusterAwareRefreshableView[U]) ClusterVirtualHostsChanged(*cluster.Cluster) {
	v.RequestUpdate(PartialUpdate)
}

func (v *ClusterAwareRefreshableView[U]) ClusterConnectionLost(*cluster.Cluster) {
	v.strategy.Pause()
	v.RefreshActions()
}

func (v *ClusterAwareRefreshableView[U]) ClusterConnectionRestored(*cluster.Cluster) {
	v.strategy.Resume()
	v.RefreshActions()
}
