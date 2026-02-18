package vhosts

import (
	"fmt"
	"tbunny/internal/cluster"
	"tbunny/internal/ui"
	"tbunny/internal/view"

	"github.com/gdamore/tcell/v2"
)

type Extender[R view.Resource] struct {
	view.ClusterAwareResourceView[R]
}

func NewVHostExtender[R view.Resource](r view.ClusterAwareResourceView[R]) view.ClusterAwareResourceView[R] {
	e := Extender[R]{
		ClusterAwareResourceView: r,
	}

	e.AddBindingKeysFn(e.bindKeys)

	return &e
}

func (e *Extender[R]) Start() {
	e.Cluster().AddListener(e)

	e.SetPath(view.VhostDisplayName(e.Cluster().ActiveVirtualHost()))
	e.ClusterAwareResourceView.Start()
}

func (e *Extender[R]) Stop() {
	e.ClusterAwareResourceView.Stop()

	e.Cluster().RemoveListener(e)
}

func (e *Extender[R]) ClusterActiveVirtualHostChanged(cluster *cluster.Cluster) {
	e.SetPath(view.VhostDisplayName(cluster.ActiveVirtualHost()))
}

func (e *Extender[R]) ClusterVirtualHostsChanged(cluster *cluster.Cluster) {
	e.SetPath(view.VhostDisplayName(cluster.ActiveVirtualHost()))
	e.RefreshActions()
}

func (e *Extender[R]) bindKeys(keyMap ui.KeyMap) {
	c := e.Cluster()

	keyMap.Add(ui.Key0, ui.NewKeyAction("all", e.switchToVirtualHostCmd))

	hostActionsCount := min(len(c.FavoriteVhosts()), 9)
	for i := 0; i < hostActionsCount; i++ {
		vhost := c.FavoriteVhosts()[i]
		keyMap.Add(ui.NumKeys[i+1], ui.NewKeyAction(vhost, e.switchToVirtualHostCmd))
	}
}

func (e *Extender[R]) switchToVirtualHostCmd(key *tcell.EventKey) *tcell.EventKey {
	var vhost string

	if key.Rune() != '0' {
		idx := int(key.Rune() - '1')
		vhost = e.Cluster().FavoriteVhosts()[idx]
	}

	e.App().StatusLine().Info(fmt.Sprintf("Switching to virtual host %s", view.VhostDisplayName(vhost)))
	e.Cluster().SetActiveVirtualHost(vhost)

	return nil
}
