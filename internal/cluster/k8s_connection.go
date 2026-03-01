package cluster

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type K8sConnectionParameters struct {
	Context   string `yaml:"context" json:"context"`
	Namespace string `yaml:"namespace" json:"namespace"`
	Name      string `yaml:"name" json:"name"`
}

type k8sConnection struct {
	parameters *K8sConnectionParameters

	uri      string
	stopChan chan struct{}
}

func newK8sConnection(ctx context.Context, params *K8sConnectionParameters) (*k8sConnection, error) {
	slog.Info(fmt.Sprintf("Creating k8s connection for context %s, namespace %s, instance %s", params.Context, params.Namespace, params.Name))

	// Loading kubeconfig with a specific context
	kubeConfigFile := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	loadingRules := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: kubeConfigFile,
	}

	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: params.Context,
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client config: %w", err)
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	// Finding a pod for the requested service
	labelSelector := map[string]string{
		"app.kubernetes.io/instance": "rabbitmq",
		"app.kubernetes.io/name":     params.Name,
	}

	pods, err := clientSet.CoreV1().
		Pods(params.Namespace).
		List(ctx, metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector).String(),
		})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for RabbitMQ instance %s in namespace %s", params.Name, params.Namespace)
	}

	// Sorting pods by name to hit rabbitmq-0
	sort.Slice(pods.Items, func(i, j int) bool {
		return pods.Items[i].Name < pods.Items[j].Name
	})

	podName := pods.Items[0].Name

	// Creating port-forward
	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create spdy round tripper: %w", err)
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", params.Namespace, podName)
	serverURL, err := url.Parse(restConfig.Host + path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server URL: %w", err)
	}

	dialer := spdy.NewDialer(
		upgrader,
		&http.Client{Transport: transport},
		"POST",
		serverURL)

	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	ports := []string{"0:15672"}

	forwarder, err := portforward.New(dialer, ports, stopChan, readyChan, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create port-forward: %w", err)
	}

	go func() {
		if err := forwarder.ForwardPorts(); err != nil {
			slog.Error(fmt.Sprintf("Failed to forward ports: %v", err))
		}
	}()

	// Waiting for port-forward to be ready
	select {
	case <-readyChan:
	case <-ctx.Done():
		close(stopChan)
		return nil, fmt.Errorf("failed to start port-forward: %w", ctx.Err())
	case <-time.After(10 * time.Second):
		close(stopChan)
		return nil, fmt.Errorf("failed to start port-forward: timeout")
	}

	// Finding out port number
	forwardedPorts, err := forwarder.GetPorts()
	if err != nil {
		close(stopChan)
		return nil, fmt.Errorf("failed to get forwarded ports: %w", err)
	}

	localPort := forwardedPorts[0].Local

	// Returning k8s connection
	return &k8sConnection{
		parameters: params,
		uri:        fmt.Sprintf("http://127.0.0.1:%d", localPort),
		stopChan:   stopChan,
	}, nil
}

func (c *k8sConnection) Uri() string {
	return c.uri
}

func (c *k8sConnection) Close() {
	slog.Info("Closing k8s connection")

	close(c.stopChan)
}

func (p *K8sConnectionParameters) String() string {
	return fmt.Sprintf("K8s connection, context %s, namespace %s, instance %s", p.Context, p.Namespace, p.Name)
}
