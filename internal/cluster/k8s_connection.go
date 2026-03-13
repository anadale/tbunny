package cluster

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"tbunny/internal/config"
	"tbunny/internal/sl"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type k8sConnection struct {
	parameters *K8sConnectionParameters
	restConfig *rest.Config
	clientSet  *kubernetes.Clientset

	uri       string
	listeners []connectionListener

	mx     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

type portForwardSession struct {
	uri      string
	stopChan chan struct{}
	done     <-chan error
}

const (
	// initialBackoff defines the initial duration to wait before retrying an operation.
	initialBackoff = 2 * time.Second

	// maximumBackoff defines the maximum duration to wait before retrying an operation.
	maximumBackoff = 30 * time.Second
)

func newK8sConnection(ctx context.Context, params *K8sConnectionParameters) (*k8sConnection, error) {
	slog.Info(fmt.Sprintf("Creating k8s connection for context %s, namespace %s, instance %s", params.Context, params.Namespace, params.Name))

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

	connCtx, cancel := context.WithCancel(context.Background())
	conn := &k8sConnection{
		parameters: params,
		restConfig: restConfig,
		clientSet:  clientSet,
		ctx:        connCtx,
		cancel:     cancel,
	}

	// Block until the first port-forward session is ready
	firstReady := make(chan error, 1)
	go conn.keepAlive(firstReady)

	select {
	case err = <-firstReady:
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to start port-forward: %w", err)
		}
		return conn, nil
	case <-ctx.Done():
		cancel()
		return nil, ctx.Err()
	}
}

func (c *k8sConnection) Uri() string {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.uri
}

func (c *k8sConnection) Close() {
	slog.Info("Closing k8s connection")
	c.cancel()
}

func (c *k8sConnection) AddListener(l connectionListener) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.listeners = append(c.listeners, l)
}

func (c *k8sConnection) notifyConnectionUriChanged(uri string) {
	c.mx.RLock()
	ls := make([]connectionListener, len(c.listeners))
	copy(ls, c.listeners)
	c.mx.RUnlock()

	for _, l := range ls {
		l.ConnectionUriChanged(uri)
	}
}

// findPod resolves the name of the first (alphabetically) pod for the RabbitMQ instance.
// Called on every connection attempt so that pod restarts are handled transparently.
func (c *k8sConnection) findPod() (string, error) {
	labelSelector := map[string]string{
		"app.kubernetes.io/instance": "rabbitmq",
		"app.kubernetes.io/name":     c.parameters.Name,
	}

	pods, err := c.clientSet.CoreV1().
		Pods(c.parameters.Namespace).
		List(c.ctx, metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector).String(),
		})
	if err != nil {
		return "", fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pods found for RabbitMQ instance %s in namespace %s", c.parameters.Name, c.parameters.Namespace)
	}

	// Sorting pods by name to hit rabbitmq-0
	sort.Slice(pods.Items, func(i, j int) bool {
		return pods.Items[i].Name < pods.Items[j].Name
	})

	return pods.Items[0].Name, nil
}

// startSession resolves the current pod, then creates a port-forward session, and waits until it is ready.
// The caller is responsible for closing session.stopChan when done.
func (c *k8sConnection) startSession() (*portForwardSession, error) {
	podName, err := c.findPod()
	if err != nil {
		return nil, err
	}

	transport, upgrader, err := spdy.RoundTripperFor(c.restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create spdy round tripper: %w", err)
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", c.parameters.Namespace, podName)
	serverURL, err := url.Parse(c.restConfig.Host + path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server URL: %w", err)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", serverURL)

	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	// io.Discard suppresses port forward's own stdout/stderr output.
	// Port 0 lets the OS assign a free local port on each reconnection.
	forwarder, err := portforward.New(dialer, []string{"0:15672"}, stopChan, readyChan, io.Discard, io.Discard)
	if err != nil {
		close(stopChan)
		return nil, fmt.Errorf("failed to create port-forward: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- forwarder.ForwardPorts()
	}()

	select {
	case err = <-done:
		// ForwardPorts exited before readyChan fired — startup failed
		return nil, fmt.Errorf("port-forward failed to start: %w", err)
	case <-readyChan:
	case <-c.ctx.Done():
		close(stopChan)
		return nil, fmt.Errorf("connection cancelled")
	case <-time.After(config.Current().ConnectionTimeout):
		close(stopChan)
		return nil, fmt.Errorf("port-forward start timed out")
	}

	forwardedPorts, err := forwarder.GetPorts()
	if err != nil {
		close(stopChan)
		return nil, fmt.Errorf("failed to get forwarded ports: %w", err)
	}

	return &portForwardSession{
		uri:      fmt.Sprintf("http://127.0.0.1:%d", forwardedPorts[0].Local),
		stopChan: stopChan,
		done:     done,
	}, nil
}

// keepAlive monitors the active port-forward session and reconnects on failure.
// It signals firstReady once (either with nil on success or an error on initial failure).
func (c *k8sConnection) keepAlive(firstReady chan<- error) {
	first := true
	backoff := initialBackoff

	for {
		session, err := c.startSession()
		if err != nil {
			if first {
				firstReady <- err
				return
			}

			slog.Warn("Failed to restart port-forward, retrying", sl.Error, err, "backoff", backoff)

			if !c.waitBackoff(backoff) {
				return
			}

			if backoff < maximumBackoff {
				backoff *= 2
			}

			continue
		}

		c.mx.Lock()
		c.uri = session.uri
		c.mx.Unlock()

		c.notifyConnectionUriChanged(session.uri)

		backoff = initialBackoff

		if first {
			firstReady <- nil
			first = false
		}

		slog.Info("Port-forward established", "uri", session.uri)

		if !c.monitorSession(session) {
			return
		}
	}
}

// waitBackoff sleeps for d and returns true to continue, false if the connection was canceled.
func (c *k8sConnection) waitBackoff(d time.Duration) bool {
	select {
	case <-c.ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

// monitorSession waits for the session to disconnect or the connection to be closed.
// Returns true if a reconnection should be attempted, false if the connection was closed.
func (c *k8sConnection) monitorSession(s *portForwardSession) bool {
	select {
	case err := <-s.done:
		close(s.stopChan)

		if c.ctx.Err() != nil {
			return false
		}

		slog.Warn("Port-forward disconnected, reconnecting", sl.Error, err)

		return true
	case <-c.ctx.Done():
		close(s.stopChan)
		return false
	}
}
