package talos

import (
	"context"

	machineapi "github.com/siderolabs/talos/pkg/machinery/api/machine"
	talosclient "github.com/siderolabs/talos/pkg/machinery/client"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
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
}

func NewMachineryClient(ctx context.Context, talosConfig *clientconfig.Config, endpoints []string) (Client, error) {
	return &machineryClient{config: talosConfig, endpoints: endpoints}, nil
}

func (c *machineryClient) ApplyMachineConfig(ctx context.Context, node string, config []byte) error {
	return c.withClient(ctx, node, func(ctx context.Context, client *talosclient.Client) error {
		_, err := client.ApplyConfiguration(ctx, &machineapi.ApplyConfigurationRequest{
			Data: config,
			Mode: machineapi.ApplyConfigurationRequest_AUTO,
		})
		return err
	})
}

func (c *machineryClient) Bootstrap(ctx context.Context, node string) error {
	return c.withClient(ctx, node, func(ctx context.Context, client *talosclient.Client) error {
		return client.Bootstrap(ctx, &machineapi.BootstrapRequest{})
	})
}

func (c *machineryClient) Kubeconfig(ctx context.Context, node string) ([]byte, error) {
	var kubeconfig []byte
	err := c.withClient(ctx, node, func(ctx context.Context, client *talosclient.Client) error {
		var err error
		kubeconfig, err = client.Kubeconfig(ctx)
		return err
	})
	return kubeconfig, err
}

func (c *machineryClient) Health(ctx context.Context, nodes []string) error {
	return c.withClient(ctx, "", func(ctx context.Context, client *talosclient.Client) error {
		for _, node := range nodes {
			if _, err := client.Version(talosclient.WithNodes(ctx, node)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *machineryClient) withClient(ctx context.Context, node string, fn func(context.Context, *talosclient.Client) error) error {
	opts := []talosclient.OptionFunc{talosclient.WithConfig(c.config)}
	if len(c.endpoints) > 0 {
		opts = append(opts, talosclient.WithEndpoints(c.endpoints...))
	}
	client, err := talosclient.New(ctx, opts...)
	if err != nil {
		return err
	}
	defer client.Close() //nolint:errcheck

	if node != "" {
		ctx = talosclient.WithNodes(ctx, node)
	}
	return fn(ctx, client)
}
