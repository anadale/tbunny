package dialogs

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"tbunny/internal/cluster"
	"tbunny/internal/ui"

	"github.com/rivo/tview"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	k8sConnectionContextFieldLabel   = "Context:"
	k8sConnectionNamespaceFieldLabel = "Namespace:"
	k8sInstanceNameFieldLabel        = "Instance name:"
)

var (
	k8sConfigLoader sync.Once
	k8sConfig       *api.Config
)

func kubernetesIsAvailable() bool {
	k8sConfigLoader.Do(func() {
		kubeConfigFile := filepath.Join(os.Getenv("HOME"), ".kube", "config")

		k8sConfig, _ = clientcmd.LoadFromFile(kubeConfigFile)
	})

	return k8sConfig != nil && len(k8sConfig.Contexts) > 0
}

func createKubernetesConnectionFields(f *ui.ModalForm) int {
	availableContexts := slices.Collect(maps.Keys(k8sConfig.Contexts))

	contextField := tview.NewDropDown().
		SetLabel(k8sConnectionContextFieldLabel).
		SetFieldWidth(30).
		SetOptions(availableContexts, nil).
		SetCurrentOption(0)
	namespaceField := tview.NewInputField().
		SetLabel(k8sConnectionNamespaceFieldLabel).
		SetFieldWidth(30).
		SetText("rabbitmq")
	instanceNameField := tview.NewInputField().
		SetLabel(k8sInstanceNameFieldLabel).
		SetFieldWidth(30).
		SetText("rabbitmq")

	f.AddFormItem(contextField)
	f.AddFormItem(namespaceField)
	f.AddFormItem(instanceNameField)

	return 3
}

func collectKubernetesConnectionParameters(f *ui.ModalForm) (*cluster.K8sConnectionParameters, bool) {
	contextField := f.GetFormItemByLabel(k8sConnectionContextFieldLabel).(*tview.DropDown)
	namespaceField := f.GetFormItemByLabel(k8sConnectionNamespaceFieldLabel).(*tview.InputField)
	instanceNameField := f.GetFormItemByLabel(k8sInstanceNameFieldLabel).(*tview.InputField)

	_, context := contextField.GetCurrentOption()
	namespace := namespaceField.GetText()
	instanceName := instanceNameField.GetText()

	if namespace == "" {
		f.SetFocus(f.GetFormItemIndex(k8sConnectionNamespaceFieldLabel))
		return nil, false
	}

	if instanceName == "" {
		f.SetFocus(f.GetFormItemIndex(k8sInstanceNameFieldLabel))
		return nil, false
	}

	return &cluster.K8sConnectionParameters{
		Context:   context,
		Namespace: namespace,
		Name:      instanceName,
	}, true
}
