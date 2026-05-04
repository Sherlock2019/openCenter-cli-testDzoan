package talos

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"
	"github.com/opencenter-cloud/opencenter-cli/internal/core/paths"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
	machineryconfig "github.com/siderolabs/talos/pkg/machinery/config"
	"github.com/siderolabs/talos/pkg/machinery/config/encoder"
	"github.com/siderolabs/talos/pkg/machinery/config/generate"
	"github.com/siderolabs/talos/pkg/machinery/config/generate/secrets"
	"github.com/siderolabs/talos/pkg/machinery/config/machine"
	v1alpha1 "github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
)

type RuntimeOption func(*Runtime)

type Runtime struct {
	cfg           *v2.Config
	clusterPaths  *paths.ClusterPaths
	artifactPaths ArtifactPaths
	clientFactory ClientFactory

	inventory      *Inventory
	secretsBundle  *secrets.Bundle
	talosConfig    *clientconfig.Config
	client         Client
	machineConfigs []MachineConfig
}

type MachineConfig struct {
	Node Node
	Data []byte
}

func WithClientFactory(factory ClientFactory) RuntimeOption {
	return func(r *Runtime) {
		r.clientFactory = factory
	}
}

func NewRuntime(cfg *v2.Config, clusterPaths *paths.ClusterPaths, opts ...RuntimeOption) (*Runtime, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Talos runtime requires config")
	}
	if clusterPaths == nil {
		return nil, fmt.Errorf("Talos runtime requires cluster paths")
	}
	runtime := &Runtime{
		cfg:           cfg,
		clusterPaths:  clusterPaths,
		artifactPaths: ResolveArtifactPaths(clusterPaths, cfg.ClusterName()),
		clientFactory: NewMachineryClient,
	}
	for _, opt := range opts {
		opt(runtime)
	}
	return runtime, nil
}

func (r *Runtime) ArtifactPaths() ArtifactPaths {
	return r.artifactPaths
}

func (r *Runtime) ReadInventory(ctx context.Context) error {
	inventory, err := LoadInventory(r.artifactPaths.InventoryPath)
	if err != nil {
		return err
	}
	r.inventory = inventory
	return nil
}

func (r *Runtime) GenerateSecrets(ctx context.Context) error {
	bundle, err := r.ensureMachineSecrets()
	if err != nil {
		return err
	}
	r.secretsBundle = bundle

	talosConfig, err := r.buildTalosConfig(bundle)
	if err != nil {
		return err
	}
	r.talosConfig = talosConfig
	return nil
}

func (r *Runtime) ApplyMachineConfigs(ctx context.Context) error {
	if err := r.ensureInventory(); err != nil {
		return err
	}
	if err := r.GenerateSecrets(ctx); err != nil {
		return err
	}
	configs, err := r.buildMachineConfigs()
	if err != nil {
		return err
	}
	client, err := r.ensureClient(ctx)
	if err != nil {
		return err
	}
	for _, config := range configs {
		endpoint := talosEndpoint(config.Node.TalosAPIIP, r.inventory.Cluster.TalosAPIPort)
		if err := client.ApplyMachineConfig(ctx, endpoint, config.Data); err != nil {
			return fmt.Errorf("apply Talos machine config to %s (%s): %w", config.Node.Name, endpoint, err)
		}
	}
	r.machineConfigs = configs
	return nil
}

func (r *Runtime) BootstrapControlPlane(ctx context.Context) error {
	if err := r.ensureInventory(); err != nil {
		return err
	}
	client, err := r.ensureClient(ctx)
	if err != nil {
		return err
	}
	node, err := r.inventory.FirstControlPlane()
	if err != nil {
		return err
	}
	endpoint := talosEndpoint(node.TalosAPIIP, r.inventory.Cluster.TalosAPIPort)
	if err := client.Bootstrap(ctx, endpoint); err != nil {
		return fmt.Errorf("bootstrap Talos control plane %s (%s): %w", node.Name, endpoint, err)
	}
	return nil
}

func (r *Runtime) ExportTalosConfig(ctx context.Context) error {
	if err := r.GenerateSecrets(ctx); err != nil {
		return err
	}
	return WriteTalosConfig(r.artifactPaths.TalosConfigPath, r.talosConfig)
}

func (r *Runtime) ExportKubeconfig(ctx context.Context) error {
	if err := r.ensureInventory(); err != nil {
		return err
	}
	client, err := r.ensureClient(ctx)
	if err != nil {
		return err
	}
	node, err := r.inventory.FirstControlPlane()
	if err != nil {
		return err
	}
	endpoint := talosEndpoint(node.TalosAPIIP, r.inventory.Cluster.TalosAPIPort)
	kubeconfig, err := client.Kubeconfig(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("export Talos kubeconfig from %s (%s): %w", node.Name, endpoint, err)
	}
	if err := os.MkdirAll(filepath.Dir(r.artifactPaths.KubeconfigPath), 0o700); err != nil {
		return fmt.Errorf("creating kubeconfig directory: %w", err)
	}
	if err := os.WriteFile(r.artifactPaths.KubeconfigPath, kubeconfig, 0o600); err != nil {
		return fmt.Errorf("writing kubeconfig %s: %w", r.artifactPaths.KubeconfigPath, err)
	}
	return nil
}

func (r *Runtime) WaitReady(ctx context.Context) error {
	if err := r.ensureInventory(); err != nil {
		return err
	}
	client, err := r.ensureClient(ctx)
	if err != nil {
		return err
	}
	nodes := r.inventory.AllNodeEndpoints()
	if err := client.Health(ctx, nodes); err != nil {
		return fmt.Errorf("waiting for Talos API readiness: %w", err)
	}
	return nil
}

func (r *Runtime) ensureInventory() error {
	if r.inventory != nil {
		return nil
	}
	return r.ReadInventory(context.Background())
}

func (r *Runtime) ensureClient(ctx context.Context) (Client, error) {
	if r.client != nil {
		return r.client, nil
	}
	if err := r.GenerateSecrets(ctx); err != nil {
		return nil, err
	}
	client, err := r.clientFactory(ctx, r.talosConfig, r.inventory.ControlPlaneEndpoints())
	if err != nil {
		return nil, fmt.Errorf("create Talos client: %w", err)
	}
	r.client = client
	return client, nil
}

func (r *Runtime) ensureMachineSecrets() (*secrets.Bundle, error) {
	if r.secretsBundle != nil {
		return r.secretsBundle, nil
	}
	if _, err := os.Stat(r.artifactPaths.MachineSecretsPath); err == nil {
		return LoadMachineSecrets(r.artifactPaths.MachineSecretsPath)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("checking machine secrets %s: %w", r.artifactPaths.MachineSecretsPath, err)
	}

	contract, err := talosVersionContract(r.cfg)
	if err != nil {
		return nil, err
	}
	bundle, err := secrets.NewBundle(secrets.NewFixedClock(time.Now()), contract)
	if err != nil {
		return nil, fmt.Errorf("generate Talos machine secrets: %w", err)
	}
	if err := WriteMachineSecrets(r.artifactPaths.MachineSecretsPath, bundle); err != nil {
		return nil, err
	}
	return bundle, nil
}

func (r *Runtime) buildTalosConfig(bundle *secrets.Bundle) (*clientconfig.Config, error) {
	if err := r.ensureInventory(); err != nil {
		return nil, err
	}
	input, err := r.newGenerateInput(Node{}, bundle)
	if err != nil {
		return nil, err
	}
	talosConfig, err := input.Talosconfig()
	if err != nil {
		return nil, fmt.Errorf("generate talosconfig: %w", err)
	}
	return talosConfig, nil
}

func (r *Runtime) buildMachineConfigs() ([]MachineConfig, error) {
	if err := r.ensureInventory(); err != nil {
		return nil, err
	}
	bundle, err := r.ensureMachineSecrets()
	if err != nil {
		return nil, err
	}

	configs := make([]MachineConfig, 0, len(r.inventory.AllNodes()))
	for idx, node := range r.inventory.ControlPlane {
		node.Role = RoleControlPlane
		machineType := machine.TypeControlPlane
		if idx == 0 {
			machineType = machine.TypeInit
		}
		data, err := r.buildMachineConfig(node, machineType, bundle)
		if err != nil {
			return nil, err
		}
		configs = append(configs, MachineConfig{Node: node, Data: data})
	}
	for _, node := range r.inventory.Workers {
		node.Role = RoleWorker
		data, err := r.buildMachineConfig(node, machine.TypeWorker, bundle)
		if err != nil {
			return nil, err
		}
		configs = append(configs, MachineConfig{Node: node, Data: data})
	}
	return configs, nil
}

func (r *Runtime) buildMachineConfig(node Node, machineType machine.Type, bundle *secrets.Bundle) ([]byte, error) {
	input, err := r.newGenerateInput(node, bundle)
	if err != nil {
		return nil, err
	}
	provider, err := input.Config(machineType)
	if err != nil {
		return nil, fmt.Errorf("generate Talos config for %s (%s): %w", node.Name, node.Role, err)
	}
	data, err := provider.EncodeBytes(encoder.WithComments(encoder.CommentsDisabled))
	if err != nil {
		return nil, fmt.Errorf("encode Talos config for %s (%s): %w", node.Name, node.Role, err)
	}

	patchNames := nilSafePatchNames(r.cfg)
	patches, err := LoadNamedPatches(r.artifactPaths.PatchesDir, patchNames, r.inventory, node)
	if err != nil {
		return nil, err
	}
	data, err = ApplyNamedPatches(data, patches, node)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Runtime) newGenerateInput(node Node, bundle *secrets.Bundle) (*generate.Input, error) {
	talosCfg := r.cfg.Deployment.Talos
	if talosCfg == nil {
		return nil, fmt.Errorf("deployment.talos is required")
	}
	contract, err := talosVersionContract(r.cfg)
	if err != nil {
		return nil, err
	}
	endpoint := strings.TrimSpace(talosCfg.Endpoint)
	if endpoint == "" && r.inventory != nil {
		endpoint = strings.TrimSpace(r.inventory.Cluster.Endpoint)
	}
	if endpoint == "" {
		return nil, fmt.Errorf("deployment.talos.endpoint or Talos inventory cluster.endpoint is required")
	}

	installDisk := strings.TrimSpace(node.InstallDisk)
	if installDisk == "" {
		installDisk = strings.TrimSpace(talosCfg.Install.Disk)
	}
	opts := []generate.Option{
		generate.WithVersionContract(contract),
		generate.WithSecretsBundle(bundle),
		generate.WithEndpointList(r.inventory.ControlPlaneEndpoints()),
		generate.WithInstallDisk(installDisk),
		generate.WithInstallImage(strings.TrimSpace(talosCfg.Install.Image)),
		generate.WithClusterCNIConfig(&v1alpha1.CNIConfig{CNIName: "none"}),
	}
	if len(node.CertSANs) > 0 {
		opts = append(opts, generate.WithAdditionalSubjectAltNames(node.CertSANs))
	}

	input, err := generate.NewInput(r.cfg.ClusterName(), endpoint, strings.TrimSpace(talosCfg.KubernetesVersion), opts...)
	if err != nil {
		return nil, fmt.Errorf("prepare Talos generator input: %w", err)
	}
	input.PodNet = []string{strings.TrimSpace(talosCfg.Network.PodSubnet)}
	input.ServiceNet = []string{strings.TrimSpace(talosCfg.Network.ServiceSubnet)}
	return input, nil
}

func talosVersionContract(cfg *v2.Config) (*machineryconfig.VersionContract, error) {
	if cfg == nil || cfg.Deployment.Talos == nil {
		return nil, fmt.Errorf("deployment.talos.version is required")
	}
	version := strings.TrimSpace(cfg.Deployment.Talos.Version)
	if version == "" {
		return nil, fmt.Errorf("deployment.talos.version is required")
	}
	contract, err := machineryconfig.ParseContractFromVersion(version)
	if err != nil {
		return nil, fmt.Errorf("parse deployment.talos.version %q: %w", version, err)
	}
	return contract, nil
}

func nilSafePatchNames(cfg *v2.Config) []string {
	if cfg == nil || cfg.Deployment.Talos == nil {
		return nil
	}
	return append([]string(nil), cfg.Deployment.Talos.Patches.Static...)
}
