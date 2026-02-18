package users

import (
	"fmt"
	"log/slog"
	"strings"
	"tbunny/internal/sl"
	"tbunny/internal/ui"
	"tbunny/internal/utils"
	"tbunny/internal/view"

	"github.com/gdamore/tcell/v2"
	rabbithole "github.com/michaelklishin/rabbit-hole/v3"
)

type View struct {
	view.ClusterAwareResourceView[*Resource]
}

func NewView() *View {
	v := View{
		view.NewClusterAwareResourceTableView[*Resource]("Users", view.NewLiveUpdateStrategy()),
	}

	v.SetResourceProvider(&v)
	v.AddBindingKeysFn(v.bindKeys)

	return &v
}

func (v *View) GetColumns() []ui.TableColumn {
	return []ui.TableColumn{
		{Name: "name", Title: "NAME", Expansion: 2},
		{Name: "tags", Title: "TAGS"},
	}
}

func (v *View) GetResources() ([]*Resource, error) {
	users, err := v.getUsers()
	if err != nil {
		return nil, err
	}

	rows := utils.Map(users, func(u rabbithole.UserInfo) *Resource { return &Resource{u} })

	return rows, nil
}

func (v *View) getUsers() ([]rabbithole.UserInfo, error) {
	c := v.Cluster()

	slog.Debug("Fetching users", sl.Component, v.Name(), sl.Cluster, c.Name())

	users, err := v.Cluster().ListUsers()
	if err != nil {
		slog.Error("Failed to fetch users", sl.Error, err, sl.Component, v.Name(), sl.Cluster, c.Name())
	}

	return users, err
}

func (v *View) CanDeleteResources() bool {
	return true
}

func (v *View) DeleteResource(resource *Resource) error {
	if resource.Name == v.Cluster().Username() {
		slog.Debug("Cannot delete current users", sl.Component, v.Name(), sl.Cluster, v.Cluster().Name(), sl.User, resource.Name)
		v.App().StatusLine().Error(fmt.Sprintf("Cannot delete current users %s", resource.Name))
		return nil
	}

	_, err := v.Cluster().DeleteUser(resource.Name)
	if err != nil {
		slog.Error("Failed to delete users", sl.Error, err, sl.Component, v.Name(), sl.Cluster, v.Cluster().Name(), sl.User, resource.Name)
		return err
	}

	return nil
}

func (v *View) bindKeys(km ui.KeyMap) {
	km.Add(ui.KeyC, ui.NewKeyAction("Create", v.createUserCmd))
	km.Add(ui.KeyE, ui.NewKeyAction("Edit", v.editUserCmd))
	km.Add(ui.KeyP, ui.NewKeyAction("Permissions", v.showPermissionsCmd))
	km.Add(ui.KeyT, ui.NewKeyAction("Topics permissions", v.showTopicsPermissionsCmd))
}

func (v *View) createUserCmd(*tcell.EventKey) *tcell.EventKey {
	ShowCreateUserDialog(v.App(), v.createUser)

	return nil
}

func (v *View) createUser(name, password, tags string) {
	v.App().StatusLine().Info(fmt.Sprintf("Creating users %s", name))

	settings := rabbithole.UserSettings{
		Name:     name,
		Password: password,
		Tags:     rabbithole.UserTags(strings.Split(tags, ",")),
	}

	_, err := v.Cluster().PutUser(name, settings)
	if err != nil {
		slog.Error("Failed to create users", sl.Error, err, sl.Component, v.Name(), sl.Cluster, v.Cluster().Name(), sl.User, name)
		v.App().StatusLine().Error(fmt.Sprintf("Failed to create users %s", name))
		return
	}

	v.App().StatusLine().Info(fmt.Sprintf("User %s created", name))
	v.App().DismissModal()
	v.RequestUpdate(view.PartialUpdate)
}

func (v *View) editUserCmd(*tcell.EventKey) *tcell.EventKey {
	if user, ok := v.GetSelectedResource(); ok {
		ShowEditUserDialog(v.App(), user.Name, strings.Join(user.Tags, ", "), v.editUser)
	}

	return nil
}

func (v *View) editUser(changePassword bool, password, tags string) {
	user, ok := v.GetSelectedResource()
	if !ok {
		v.App().DismissModal()
		return
	}

	name := user.Name
	v.App().StatusLine().Info(fmt.Sprintf("Updating users %s", name))

	settings := rabbithole.UserSettings{
		Name: name,
		Tags: rabbithole.UserTags(strings.Split(tags, ",")),
	}

	if changePassword {
		settings.Password = password
	} else {
		settings.PasswordHash = user.PasswordHash
		settings.HashingAlgorithm = user.HashingAlgorithm
	}

	_, err := v.Cluster().PutUser(name, settings)
	if err != nil {
		slog.Error("Failed to update users", sl.Error, err, sl.Component, v.Name(), sl.Cluster, v.Cluster().Name(), sl.User, name)
		v.App().StatusLine().Error(fmt.Sprintf("Failed to update users %s", name))
		return
	}

	v.App().StatusLine().Info(fmt.Sprintf("User %s updated", name))
	v.App().DismissModal()
	v.RequestUpdate(view.PartialUpdate)
}

func (v *View) showPermissionsCmd(*tcell.EventKey) *tcell.EventKey {
	user, ok := v.GetSelectedResource()
	if !ok {
		return nil
	}

	pv := NewVhostsPermissionsView(user.Name)
	if err := v.App().AddView(pv); err != nil {
		v.App().StatusLine().Error(fmt.Sprintf("Failed to load permissions: %s", err))
	}

	return nil
}

func (v *View) showTopicsPermissionsCmd(*tcell.EventKey) *tcell.EventKey {
	user, ok := v.GetSelectedResource()
	if !ok {
		return nil
	}

	pv := NewTopicsPermissionsView(user.Name)
	if err := v.App().AddView(pv); err != nil {
		v.App().StatusLine().Error(fmt.Sprintf("Failed to load topics permissions: %s", err))
	}

	return nil
}
