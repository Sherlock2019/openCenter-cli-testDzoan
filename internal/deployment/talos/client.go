package talos

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	machineapi "github.com/siderolabs/talos/pkg/machinery/api/machine"
	talosclient "github.com/siderolabs/talos/pkg/machinery/client"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
	"google.golang.org/grpc"
)

type Client interface {
	ApplyMachineConfig(ctx context.Context, node string, config []byte) error
	Bootstrap(ctx context.Context, node string) error
	Kubeconfig(ctx context.Context, node string) ([]byte, error)
	Health(ctx context.Context, nodes []string) error
}

type ClientFactory func(ctx context.Context, talosConfig *clientconfig.Config, endpoints []string) (Client, error)

type machineryClient struct {
	config    *clientconfig.Config
	endpoints []string
	newClient talosAPIClientFactory
}

func NewMachineryClient(ctx context.Context, talosConfig *clientconfig.Config, endpoints []string) (Client, error) {
	return &machineryClient{config: talosConfig, endpoints: endpoints, newClient: newTalosAPIClient}, nil
}

func (c *machineryClient) ApplyMachineConfig(ctx context.Context, node string, config []byte) error {
	err := c.withMaintenanceClient(ctx, node, func(ctx context.Context, client talosAPIClient) error {
		_, err := client.ApplyConfiguration(ctx, &machineapi.ApplyConfigurationRequest{
			Data: config,
			Mode: machineapi.ApplyConfigurationRequest_AUTO,
		})
		return err
	})
	return wrapTalosOperationError("apply machine config", "", node, err)
}

func (c *machineryClient) Bootstrap(ctx context.Context, node string) error {
	err := c.withAuthenticatedClient(ctx, node, func(ctx context.Context, client talosAPIClient) error {
		return client.Bootstrap(ctx, &machineapi.BootstrapRequest{})
	})
	return wrapTalosOperationError("bootstrap", "", node, err)
}

func (c *machineryClient) Kubeconfig(ctx context.Context, node string) ([]byte, error) {
	var kubeconfig []byte
	err := c.withAuthenticatedClient(ctx, node, func(ctx context.Context, client talosAPIClient) error {
		var err error
		kubeconfig, err = client.Kubeconfig(ctx)
		return err
	})
	return kubeconfig, wrapTalosOperationError("export kubeconfig", "", node, err)
}

func (c *machineryClient) Health(ctx context.Context, nodes []string) error {
	for _, node := range nodes {
		err := c.withAuthenticatedClient(ctx, node, func(ctx context.Context, client talosAPIClient) error {
			_, err := client.Version(ctx)
			return err
		})
		if err != nil {
			return wrapTalosOperationError("health check", "", node, err)
		}
	}
	return nil
}

func (c *machineryClient) withAuthenticatedClient(ctx context.Context, endpoint string, fn func(context.Context, talosAPIClient) error) error {
	return c.withClient(ctx, talosClientModeAuthenticated, endpoint, fn)
}

func (c *machineryClient) withMaintenanceClient(ctx context.Context, endpoint string, fn func(context.Context, talosAPIClient) error) error {
	return c.withClient(ctx, talosClientModeMaintenance, endpoint, fn)
}

func (c *machineryClient) withClient(ctx context.Context, mode talosClientMode, endpoint string, fn func(context.Context, talosAPIClient) error) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return fmt.Errorf("Talos API endpoint is required")
	}
	client, err := c.newClient(ctx, mode, endpoint, c.config)
	if err != nil {
		return err
	}
	defer client.Close() //nolint:errcheck

	return fn(ctx, client)
}

type talosClientMode string

const (
	talosClientModeAuthenticated talosClientMode = "authenticated"
	talosClientModeMaintenance   talosClientMode = "maintenance"
)

type talosAPIClient interface {
	ApplyConfiguration(context.Context, *machineapi.ApplyConfigurationRequest, ...grpc.CallOption) (*machineapi.ApplyConfigurationResponse, error)
	Bootstrap(context.Context, *machineapi.BootstrapRequest) error
	Kubeconfig(context.Context) ([]byte, error)
	Version(context.Context, ...grpc.CallOption) (*machineapi.VersionResponse, error)
	Close() error
}

type talosAPIClientFactory func(context.Context, talosClientMode, string, *clientconfig.Config) (talosAPIClient, error)

func newTalosAPIClient(ctx context.Context, mode talosClientMode, endpoint string, config *clientconfig.Config) (talosAPIClient, error) {
	opts := []talosclient.OptionFunc{talosclient.WithEndpoints(endpoint)}
	switch mode {
	case talosClientModeAuthenticated:
		opts = append([]talosclient.OptionFunc{talosclient.WithConfig(config)}, opts...)
	case talosClientModeMaintenance:
		opts = append(opts, talosclient.WithTLSConfig(&tls.Config{InsecureSkipVerify: true})) //nolint:gosec
	default:
		return nil, fmt.Errorf("unsupported Talos client mode %q", mode)
	}
	return talosclient.New(ctx, opts...)
}

func wrapTalosOperationError(operation, nodeName, endpoint string, err error) error {
	if err == nil {
		return nil
	}
	details := operation
	if strings.TrimSpace(nodeName) != "" {
		details += fmt.Sprintf(" for node %s", nodeName)
	}
	if strings.TrimSpace(endpoint) != "" {
		details += fmt.Sprintf(" at %s", endpoint)
	}
	if remediation := talosNetworkRemediation(err); remediation != "" {
		return fmt.Errorf("%s failed: %w. %s", details, err, remediation)
	}
	return fmt.Errorf("%s failed: %w", details, err)
}

func talosNetworkRemediation(err error) string {
	if err == nil {
		return ""
	}
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout() || os.IsTimeout(err) {
		return "Verify the node management floating IP is reachable on TCP 50000 and deployment.talos.network.management_cidrs includes the operator CIDR."
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "connection refused"):
		return "Verify Talos API is listening on TCP 50000 and the OpenStack security group allows the configured management CIDRs."
	case strings.Contains(message, "no such host"):
		return "Verify the node management endpoint DNS name resolves, or use the management floating IP directly."
	case strings.Contains(message, "i/o timeout"), strings.Contains(message, "timeout"):
		return "Verify the node management floating IP is reachable on TCP 50000 and deployment.talos.network.management_cidrs includes the operator CIDR."
	default:
		return ""
	}
}
