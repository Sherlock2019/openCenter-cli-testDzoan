package cluster

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencenter-cloud/opencenter-cli/internal/config"
	configflags "github.com/opencenter-cloud/opencenter-cli/internal/config/flags"
	"github.com/opencenter-cloud/opencenter-cli/internal/core/paths"
	"github.com/opencenter-cloud/opencenter-cli/internal/core/validation"
	"github.com/opencenter-cloud/opencenter-cli/internal/sops"
	"github.com/opencenter-cloud/opencenter-cli/internal/util/crypto"
	"github.com/opencenter-cloud/opencenter-cli/internal/util/errors"
	"github.com/opencenter-cloud/opencenter-cli/internal/util/fs"
	"gopkg.in/yaml.v3"
)

// InitOptions contains options for cluster initialization
type InitOptions struct {
	ClusterName    string
	Organization   string
	Provider       string
	ConfigFile     string
	ConfigMap      map[string]any
	Force          bool
	Strict         bool
	NoKeyGen       bool
	NoSOPSKeyGen   bool
	RegenerateKeys bool
	NoGitInit      bool
	FullSchema     bool
	SchemaVersion  string
	ServerPools    []string
	FlagOverrides  []string
}

// InitResult contains the result of cluster initialization
type InitResult struct {
	ConfigPath     string
	ClusterPaths   *paths.ClusterPaths
	Config         *config.Config
	ConfigMap      map[string]any
	KeysGenerated  bool
	GitInitialized bool
	Message        string
}

// InitService handles cluster initialization business logic
type InitService struct {
	pathResolver     *paths.PathResolver
	validationEngine *validation.ValidationEngine
	configManager    *config.ConfigManager
	configurationMgr *config.ConfigurationManager
	fileSystem       fs.FileSystem
}

// NewInitService creates a new InitService
func NewInitService(
	pathResolver *paths.PathResolver,
	validationEngine *validation.ValidationEngine,
	configManager *config.ConfigManager,
) *InitService {
	return NewInitServiceWithConfigMgr(pathResolver, validationEngine, configManager, nil, nil)
}

// NewInitServiceWithConfigMgr creates a new InitService with optional ConfigurationManager
func NewInitServiceWithConfigMgr(
	pathResolver *paths.PathResolver,
	validationEngine *validation.ValidationEngine,
	configManager *config.ConfigManager,
	configurationMgr *config.ConfigurationManager,
	fileSystem fs.FileSystem,
) *InitService {
	// Create ConfigurationManager if not provided
	if configurationMgr == nil {
		// Create ConfigurationManager with the provided validation engine
		errorHandler := errors.NewDefaultErrorHandlerWithoutMasking()
		fs := fs.NewDefaultFileSystem(errorHandler)

		// Use the pathResolver's base directory
		configurationMgr = config.NewConfigurationManagerWithDeps(
			config.NewConfigIOHandler(fs),
			validationEngine, // Use the validation engine from DI
			config.NewConfigCache(),
			pathResolver,
			fs,
		)
	}

	// Create FileSystem if not provided
	if fileSystem == nil {
		errorHandler := errors.NewDefaultErrorHandlerWithoutMasking()
		fileSystem = fs.NewDefaultFileSystem(errorHandler)
	}

	return &InitService{
		pathResolver:     pathResolver,
		validationEngine: validationEngine,
		configManager:    configManager,
		configurationMgr: configurationMgr,
		fileSystem:       fileSystem,
	}
}

// Initialize performs cluster initialization
func (s *InitService) Initialize(ctx context.Context, opts InitOptions) (*InitResult, error) {
	// Validate cluster name
	if err := s.validateClusterName(ctx, opts.ClusterName); err != nil {
		return nil, fmt.Errorf("invalid cluster name: %w", err)
	}

	// Validate organization
	if opts.Organization == "" {
		opts.Organization = "opencenter"
	}
	if err := s.validateOrganization(ctx, opts.Organization); err != nil {
		return nil, fmt.Errorf("invalid organization: %w", err)
	}

	// Use PathResolver to resolve paths for the cluster
	// For new clusters, this will use the organization-based strategy
	strategy := s.pathResolver.GetStrategies()[0]
	clusterPaths, err := strategy.Resolve(ctx, opts.ClusterName, opts.Organization)
	if err != nil {
		return nil, fmt.Errorf("resolving cluster paths: %w", err)
	}

	// Check if cluster exists and handle force flag
	if err := s.checkExistingCluster(clusterPaths, opts.ClusterName, opts.Organization, opts.Force); err != nil {
		return nil, err
	}

	// Load or create configuration
	cfg, configMap, err := s.loadOrCreateConfig(opts)
	if err != nil {
		return nil, fmt.Errorf("loading/creating config: %w", err)
	}

	// Apply configuration overrides
	if err := s.applyOverrides(cfg, configMap, opts); err != nil {
		return nil, fmt.Errorf("applying overrides: %w", err)
	}

	// Update configuration with resolved paths
	s.updateConfigPaths(cfg, configMap, clusterPaths, opts)

	// Validate configuration if strict mode
	if opts.Strict {
		if err := s.validateConfig(cfg); err != nil {
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
	}

	// Create directory structure
	if err := s.createDirectories(ctx, clusterPaths, opts.Organization); err != nil {
		return nil, fmt.Errorf("creating directories: %w", err)
	}

	// Initialize result
	result := &InitResult{
		ConfigPath:   clusterPaths.ConfigPath,
		ClusterPaths: clusterPaths,
		Config:       cfg,
		ConfigMap:    configMap,
	}

	// Generate keys if requested
	if !opts.NoKeyGen {
		keysGenerated, err := s.generateKeys(clusterPaths, cfg, opts)
		if err != nil {
			return nil, fmt.Errorf("generating keys: %w", err)
		}
		result.KeysGenerated = keysGenerated
	}

	// Save configuration
	if err := s.saveConfig(ctx, cfg, configMap, clusterPaths.ConfigPath, shouldPreserveRawConfig(configMap)); err != nil {
		return nil, fmt.Errorf("saving config: %w", err)
	}

	// Initialize git repository if requested
	if !opts.NoGitInit {
		if err := s.initGitRepo(clusterPaths); err != nil {
			return nil, fmt.Errorf("initializing git repo: %w", err)
		}
		result.GitInitialized = true
	}

	// Build result message
	result.Message = s.buildResultMessage(clusterPaths, opts.Organization, cfg.OpenCenter.GitOps.GitDir, result.KeysGenerated)

	return result, nil
}

// validateClusterName validates the cluster name using the validation engine
func (s *InitService) validateClusterName(ctx context.Context, name string) error {
	result, err := s.validationEngine.Validate(ctx, "cluster-name", name)
	if err != nil {
		return err
	}

	if !result.Valid {
		return fmt.Errorf("validation failed: %v", result.Errors)
	}

	return nil
}

// validateOrganization validates the organization name
func (s *InitService) validateOrganization(ctx context.Context, organization string) error {
	result, err := s.validationEngine.Validate(ctx, "organization-name", organization)
	if err != nil {
		return err
	}

	if !result.Valid {
		return fmt.Errorf("validation failed: %v", result.Errors)
	}

	return nil
}

// checkExistingCluster checks if cluster already exists and handles force flag
func (s *InitService) checkExistingCluster(clusterPaths *paths.ClusterPaths, clusterName, organization string, force bool) error {
	if _, err := os.Stat(clusterPaths.ClusterDir); err == nil {
		if !force {
			return fmt.Errorf("cluster '%s' already exists in organization '%s' at %s, use --force to overwrite", clusterName, organization, clusterPaths.ClusterDir)
		}
		// If force is true, we'll overwrite the config but preserve keys
	}
	return nil
}

// loadOrCreateConfig loads configuration from file or creates default
func (s *InitService) loadOrCreateConfig(opts InitOptions) (*config.Config, map[string]any, error) {
	var cfg config.Config
	var configMap map[string]any

	if opts.ConfigFile != "" {
		// Load from file
		data, err := s.fileSystem.ReadFile(opts.ConfigFile)
		if err != nil {
			return nil, nil, fmt.Errorf("reading config file: %w", err)
		}

		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, nil, fmt.Errorf("parsing config file: %w", err)
		}

		configMap = make(map[string]any)
		if err := yaml.Unmarshal(data, &configMap); err != nil {
			return nil, nil, fmt.Errorf("parsing config file to map: %w", err)
		}
	} else if opts.ConfigMap != nil {
		// Use provided config map
		configMap = opts.ConfigMap
		data, err := yaml.Marshal(configMap)
		if err != nil {
			return nil, nil, fmt.Errorf("marshaling config map: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, nil, fmt.Errorf("parsing config map: %w", err)
		}
	} else {
		// Create default configuration
		var err error
		cfg, configMap, err = s.createDefaultConfig(opts)
		if err != nil {
			return nil, nil, fmt.Errorf("creating default config: %w", err)
		}
	}

	return &cfg, configMap, nil
}

// createDefaultConfig creates a default configuration based on options
func (s *InitService) createDefaultConfig(opts InitOptions) (config.Config, map[string]any, error) {
	var (
		cfg       config.Config
		configMap map[string]any
		err       error
	)

	if opts.FullSchema {
		// Keep full-schema generation on the existing path, then overlay provider defaults below.
		schemaDefaultYAML, err := config.GenerateFullSchemaDefaults(opts.ClusterName)
		if err != nil {
			return config.Config{}, nil, fmt.Errorf("generating schema defaults: %w", err)
		}

		if err := yaml.Unmarshal(schemaDefaultYAML, &cfg); err != nil {
			return config.Config{}, nil, fmt.Errorf("parsing schema defaults to struct: %w", err)
		}

		configMap = make(map[string]any)
		if err := yaml.Unmarshal(schemaDefaultYAML, &configMap); err != nil {
			return config.Config{}, nil, fmt.Errorf("parsing schema defaults to map: %w", err)
		}
	} else {
		cfg, err = config.NewProviderDefault(opts.ClusterName, opts.Provider)
		if err != nil {
			return config.Config{}, nil, fmt.Errorf("building provider defaults: %w", err)
		}
	}

	// Set schema version to v2.0
	cfg.SchemaVersion = "2.0"

	// Set organization
	if opts.Organization != "" {
		cfg.OpenCenter.Meta.Organization = opts.Organization
	} else {
		cfg.OpenCenter.Meta.Organization = "opencenter"
	}

	// Set provider if specified
	if opts.Provider != "" {
		cfg.OpenCenter.Infrastructure.Provider = opts.Provider
	}

	if err := config.ApplyProviderDefaults(&cfg, cfg.OpenCenter.Infrastructure.Provider); err != nil {
		return config.Config{}, nil, fmt.Errorf("applying provider defaults: %w", err)
	}

	// Set storage plugin based on provider type
	provider := cfg.OpenCenter.Infrastructure.Provider
	switch provider {
	case "openstack":
		cfg.OpenCenter.Cluster.Kubernetes.StoragePlugin.Cinder.Enabled = true
		cfg.OpenCenter.Cluster.Kubernetes.StoragePlugin.Vsphere.Enabled = false
	case "vmware":
		cfg.OpenCenter.Cluster.Kubernetes.StoragePlugin.Vsphere.Enabled = true
		cfg.OpenCenter.Cluster.Kubernetes.StoragePlugin.Cinder.Enabled = false
	default:
		// For other providers (kind, baremetal, aws), leave storage plugins disabled by default
		cfg.OpenCenter.Cluster.Kubernetes.StoragePlugin.Cinder.Enabled = false
		cfg.OpenCenter.Cluster.Kubernetes.StoragePlugin.Vsphere.Enabled = false
	}

	// Set initial stage and status
	cfg.OpenCenter.Meta.Stage = config.StageInit
	cfg.OpenCenter.Meta.Status = config.StatusSuccess

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return config.Config{}, nil, fmt.Errorf("marshaling default config: %w", err)
	}

	cfgMap := make(map[string]any)
	if err := yaml.Unmarshal(data, &cfgMap); err != nil {
		return config.Config{}, nil, fmt.Errorf("rebuilding default config map: %w", err)
	}

	if opts.FullSchema {
		deepMergeConfigMap(configMap, cfgMap)
	} else {
		configMap = cfgMap
	}

	return cfg, configMap, nil
}

// applyOverrides applies configuration overrides from options
func (s *InitService) applyOverrides(cfg *config.Config, configMap map[string]any, opts InitOptions) error {
	// Apply organization
	if opts.Organization != "" {
		cfg.OpenCenter.Meta.Organization = opts.Organization
		setNestedConfigValue(configMap, opts.Organization, "opencenter", "meta", "organization")
	}

	// Apply provider
	if opts.Provider != "" {
		cfg.OpenCenter.Infrastructure.Provider = opts.Provider
		setNestedConfigValue(configMap, opts.Provider, "opencenter", "infrastructure", "provider")
	}

	// Apply CLI config defaults
	cliConfig := s.configManager.GetConfig()
	if cfg.OpenCenter.Meta.Region == "" || cfg.OpenCenter.Meta.Region == "sjc3" {
		if cliConfig.Defaults.Region != "" {
			cfg.OpenCenter.Meta.Region = cliConfig.Defaults.Region
			setNestedConfigValue(configMap, cliConfig.Defaults.Region, "opencenter", "meta", "region")
		}
	}
	if cfg.OpenCenter.Meta.Env == "" || cfg.OpenCenter.Meta.Env == "dev" {
		if cliConfig.Defaults.Environment != "" {
			cfg.OpenCenter.Meta.Env = cliConfig.Defaults.Environment
			setNestedConfigValue(configMap, cliConfig.Defaults.Environment, "opencenter", "meta", "env")
		}
	}

	if len(opts.FlagOverrides) > 0 {
		integration, err := configflags.NewCLIIntegration()
		if err != nil {
			return fmt.Errorf("creating flag integration: %w", err)
		}
		if err := integration.ProcessFlags(opts.FlagOverrides, cfg, configMap); err != nil {
			return fmt.Errorf("applying dotted overrides: %w", err)
		}
	}

	return nil
}

// updateConfigPaths updates configuration with resolved paths
func (s *InitService) updateConfigPaths(cfg *config.Config, configMap map[string]any, clusterPaths *paths.ClusterPaths, opts InitOptions) {
	if !hasExplicitConfigValue(configMap, opts, "opencenter", "gitops", "git_dir") {
		cfg.OpenCenter.GitOps.GitDir = clusterPaths.GitOpsDir
		setNestedConfigValue(configMap, clusterPaths.GitOpsDir, "opencenter", "gitops", "git_dir")
	}

	// Update SSH key paths
	env := cfg.OpenCenter.Meta.Env
	region := cfg.OpenCenter.Meta.Region
	if env == "" {
		env = "dev"
	}
	if region == "" {
		region = "local"
	}

	sshKeyBaseName := fmt.Sprintf("%s-%s-%s", opts.ClusterName, env, region)
	sshKeyPath := filepath.Join(clusterPaths.SecretsDir, "ssh", sshKeyBaseName)

	cfg.OpenCenter.GitOps.GitSSHKey = sshKeyPath
	cfg.OpenCenter.GitOps.GitSSHPub = sshKeyPath + ".pub"
	cfg.Secrets.SSHKey.Private = sshKeyPath
	cfg.Secrets.SSHKey.Public = sshKeyPath + ".pub"
	setNestedConfigValue(configMap, sshKeyPath, "opencenter", "gitops", "git_ssh_key")
	setNestedConfigValue(configMap, sshKeyPath+".pub", "opencenter", "gitops", "git_ssh_pub")
	setNestedConfigValue(configMap, sshKeyPath, "secrets", "ssh_key", "private")
	setNestedConfigValue(configMap, sshKeyPath+".pub", "secrets", "ssh_key", "public")

	// Update SOPS key path
	if opts.NoSOPSKeyGen && !hasExplicitConfigValue(configMap, opts, "secrets", "sops_age_key_file") {
		cfg.Secrets.SopsAgeKeyFile = ""
		setNestedConfigValue(configMap, "", "secrets", "sops_age_key_file")
		return
	}

	cfg.Secrets.SopsAgeKeyFile = clusterPaths.SOPSKeyPath
	setNestedConfigValue(configMap, clusterPaths.SOPSKeyPath, "secrets", "sops_age_key_file")
}

// validateConfig validates the configuration
func (s *InitService) validateConfig(cfg *config.Config) error {
	// Use ConfigurationManager for validation if available
	if s.configurationMgr != nil {
		if err := s.configurationMgr.Validate(context.Background(), cfg); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	} else {
		// Fallback: create temporary manager
		tempMgr, err := config.NewConfigurationManager()
		if err != nil {
			return fmt.Errorf("creating configuration manager: %w", err)
		}
		if err := tempMgr.Validate(context.Background(), cfg); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}
	return nil
}

// createDirectories creates the necessary directory structure using PathResolver
func (s *InitService) createDirectories(ctx context.Context, clusterPaths *paths.ClusterPaths, organization string) error {
	// Extract cluster name from the cluster directory path
	clusterName := filepath.Base(clusterPaths.ClusterDir)

	// Use PathResolver to create all necessary directories
	if err := s.pathResolver.CreateClusterDirectories(ctx, clusterName, organization); err != nil {
		return fmt.Errorf("creating cluster directories: %w", err)
	}

	return nil
}

// saveConfig saves the configuration to disk using ConfigurationManager
func (s *InitService) saveConfig(ctx context.Context, cfg *config.Config, configMap map[string]any, configPath string, preserveRaw bool) error {
	// Use ConfigurationManager if available
	if !preserveRaw && s.configurationMgr != nil {
		return s.configurationMgr.Save(ctx, cfg)
	}

	var (
		data []byte
		err  error
	)
	if preserveRaw {
		data, err = yaml.Marshal(configMap)
	} else {
		data, err = yaml.Marshal(cfg)
	}
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := s.fileSystem.WriteFileAtomic(configPath, data, 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// buildResultMessage builds a user-friendly result message
func (s *InitService) buildResultMessage(clusterPaths *paths.ClusterPaths, organization, gitDir string, keysGenerated bool) string {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("Created cluster configuration in organization '%s' at '%s'\n", organization, clusterPaths.ClusterDir))
	msg.WriteString(fmt.Sprintf("GitOps repository root: %s\n", gitDir))
	if keysGenerated {
		msg.WriteString(fmt.Sprintf("SOPS key location: %s\n", clusterPaths.SOPSKeyPath))
	}
	return msg.String()
}

func shouldPreserveRawConfig(configMap map[string]any) bool {
	_, ok := configMap["iac"]
	return ok
}

func deepMergeConfigMap(dst, src map[string]any) {
	for key, value := range src {
		if existing, ok := dst[key]; ok {
			dstMap, dstOK := existing.(map[string]any)
			srcMap, srcOK := value.(map[string]any)
			if dstOK && srcOK {
				deepMergeConfigMap(dstMap, srcMap)
				continue
			}
		}
		dst[key] = value
	}
}

func setNestedConfigValue(configMap map[string]any, value any, parts ...string) {
	if len(parts) == 0 {
		return
	}

	current := configMap
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part].(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[part] = next
		}
		current = next
	}

	current[parts[len(parts)-1]] = value
}

func hasExplicitConfigValue(configMap map[string]any, opts InitOptions, parts ...string) bool {
	dottedPath := strings.Join(parts, ".")
	if hasFlagOverride(opts.FlagOverrides, dottedPath) {
		return true
	}

	if opts.ConfigFile == "" && opts.ConfigMap == nil {
		return false
	}

	current := any(configMap)
	for _, part := range parts {
		next, ok := current.(map[string]any)
		if !ok {
			return false
		}
		value, exists := next[part]
		if !exists {
			return false
		}
		current = value
	}

	return true
}

func hasFlagOverride(overrides []string, key string) bool {
	prefix := "--" + key + "="
	for _, override := range overrides {
		if strings.HasPrefix(override, prefix) {
			return true
		}
	}
	return false
}

// generateKeys generates SSH and SOPS keys for the cluster
func (s *InitService) generateKeys(clusterPaths *paths.ClusterPaths, cfg *config.Config, opts InitOptions) (bool, error) {
	keysGenerated := false

	// Generate SOPS Age key if not disabled
	if !opts.NoSOPSKeyGen {
		sopsKeyExists := false
		if _, err := os.Stat(clusterPaths.SOPSKeyPath); err == nil {
			sopsKeyExists = true
		}

		if opts.RegenerateKeys || !sopsKeyExists {
			if err := s.generateSOPSKey(clusterPaths); err != nil {
				return false, fmt.Errorf("generating SOPS key: %w", err)
			}
			keysGenerated = true
		}

		if cfg.Secrets.SopsAgeKeyFile != "" {
			if _, err := os.Stat(clusterPaths.SOPSConfigPath); os.IsNotExist(err) || opts.RegenerateKeys || keysGenerated {
				if err := s.ensureSOPSConfig(clusterPaths, cfg); err != nil {
					return false, fmt.Errorf("creating SOPS config: %w", err)
				}
			}
		}
	}

	// Generate SSH key pair
	sshKeyExists := false
	if _, err := os.Stat(clusterPaths.SSHKeyPath); err == nil {
		sshKeyExists = true
	}

	if opts.RegenerateKeys || !sshKeyExists {
		if err := s.generateSSHKey(clusterPaths, cfg); err != nil {
			return false, fmt.Errorf("generating SSH key: %w", err)
		}
		keysGenerated = true
	}

	return keysGenerated, nil
}

// generateSOPSKey generates a SOPS Age key for the cluster
func (s *InitService) generateSOPSKey(clusterPaths *paths.ClusterPaths) error {
	// Create the secrets directory with proper permissions
	secretsKeyDir := filepath.Dir(clusterPaths.SOPSKeyPath)
	if err := os.MkdirAll(secretsKeyDir, 0o755); err != nil {
		return fmt.Errorf("creating secrets directory: %w", err)
	}

	// Use the SOPS key manager to generate and save an Age key
	km := sops.NewKeyManager(secretsKeyDir)
	keyPair, err := km.GenerateAgeKey()
	if err != nil {
		return fmt.Errorf("generating Age key pair: %w", err)
	}

	// Extract cluster name from the key path
	// Path format: <org>/secrets/age/<cluster>-key.txt
	keyFileName := filepath.Base(clusterPaths.SOPSKeyPath)
	keyName := strings.TrimSuffix(keyFileName, ".txt")

	// Save the key using the key manager
	if err := km.SaveAgeKey(keyPair, keyName); err != nil {
		return fmt.Errorf("saving Age key pair: %w", err)
	}

	return nil
}

func (s *InitService) ensureSOPSConfig(clusterPaths *paths.ClusterPaths, cfg *config.Config) error {
	publicKey, err := s.readAgePublicKey(cfg.Secrets.SopsAgeKeyFile)
	if err != nil {
		return err
	}

	content := fmt.Sprintf(`creation_rules:
  - path_regex: 'secrets/age/keys/.*-key\.txt$'
    age: >-
      %s
  - path_regex: 'secrets/ssh/(?!.*\.pub$).*'
    age: >-
      %s
  - path_regex: 'applications/overlays/[^/]+/(managed-services|services)/.*/.*\.ya?ml$'
    encrypted_regex: "^(secret)$"
    age: >-
      %s
  - path_regex: '^infrastructure\/clusters\/%s\/(?!(?:venv|kubespray|\.terraform|\.bin)\/)(.*)'
    encrypted_regex: "^(secret)$"
    age: >-
      %s
`, publicKey, publicKey, publicKey, cfg.OpenCenter.Cluster.ClusterName, publicKey)

	return s.fileSystem.WriteFileAtomic(clusterPaths.SOPSConfigPath, []byte(content), 0o600)
}

func (s *InitService) readAgePublicKey(keyPath string) (string, error) {
	if keyPath == "" {
		return "", fmt.Errorf("SOPS age key path cannot be empty")
	}

	data, err := s.fileSystem.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("reading SOPS age key: %w", err)
	}

	privateKey := extractAgePrivateKey(string(data))
	if privateKey == "" {
		return "", fmt.Errorf("no valid age private key found in %s", keyPath)
	}

	keyPair, err := crypto.ParseAgeKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("parsing SOPS age key: %w", err)
	}

	return strings.TrimSpace(keyPair.PublicKey), nil
}

func extractAgePrivateKey(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "AGE-SECRET-KEY-") {
			return trimmed
		}
	}

	return ""
}

// generateSSHKey generates an SSH key pair for the cluster
func (s *InitService) generateSSHKey(clusterPaths *paths.ClusterPaths, cfg *config.Config) error {
	// Create SSH directory if it doesn't exist
	sshDir := filepath.Dir(clusterPaths.SSHKeyPath)
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return fmt.Errorf("creating SSH directory: %w", err)
	}

	// Extract cluster name and organization from paths for the comment
	// Path format: <org>/secrets/ssh/<cluster>-<env>-<region>
	keyFileName := filepath.Base(clusterPaths.SSHKeyPath)
	parts := strings.Split(keyFileName, "-")
	clusterName := parts[0]

	// Get organization from the path
	pathParts := strings.Split(clusterPaths.SSHKeyPath, string(filepath.Separator))
	var organization string
	for i, part := range pathParts {
		if part == "clusters" && i+1 < len(pathParts) {
			organization = pathParts[i+1]
			break
		}
	}

	// Get region from config
	region := cfg.OpenCenter.Meta.Region
	if region == "" {
		region = "local"
	}

	// Create SSH key comment
	sshKeyComment := fmt.Sprintf("%s-%s-%s", organization, clusterName, region)

	// Get cipher from config or default to ed25519
	cipher := cfg.Secrets.SSHKey.Cypher
	if cipher == "" {
		cipher = "ed25519"
	}

	// Generate SSH key pair
	keyPair, err := crypto.GenerateSSHKeyWithComment(cipher, sshKeyComment)
	if err != nil {
		return fmt.Errorf("generating SSH key pair: %w", err)
	}

	// Write private key with restrictive permissions
	if err := s.fileSystem.WriteFileAtomic(clusterPaths.SSHKeyPath, keyPair.PrivateKey, 0o600); err != nil {
		return fmt.Errorf("writing SSH private key: %w", err)
	}

	// Write public key
	pubKeyPath := clusterPaths.SSHKeyPath + ".pub"
	if err := s.fileSystem.WriteFile(pubKeyPath, keyPair.PublicKey, 0o644); err != nil {
		return fmt.Errorf("writing SSH public key: %w", err)
	}

	if shouldPopulateGeneratedSSHAuthorizedKey(cfg.OpenCenter.Cluster.SSHAuthorizedKeys) {
		cfg.OpenCenter.Cluster.SSHAuthorizedKeys = []string{strings.TrimSpace(string(keyPair.PublicKey))}
	}

	return nil
}

func shouldPopulateGeneratedSSHAuthorizedKey(keys []string) bool {
	if len(keys) == 0 {
		return true
	}

	for _, key := range keys {
		trimmed := strings.TrimSpace(key)
		if trimmed != "" && trimmed != config.DefaultSSHAuthorizedKeyPlaceholder {
			return false
		}
	}

	return true
}

// initGitRepo initializes a git repository for the cluster
func (s *InitService) initGitRepo(clusterPaths *paths.ClusterPaths) error {
	// Create the GitOps directory if it doesn't exist
	if err := os.MkdirAll(clusterPaths.GitOpsDir, 0o755); err != nil {
		return fmt.Errorf("creating GitOps directory: %w", err)
	}

	// Check if git repository already exists
	gitDir := filepath.Join(clusterPaths.GitOpsDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Git repository already exists, skip initialization
		return nil
	}

	// Initialize git repository
	// Note: This is a placeholder implementation
	// The actual git initialization will be done by the command layer
	// which has access to the cobra command for output
	return nil
}
