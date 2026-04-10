package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/opencenter-cloud/opencenter-cli/internal/config"
	"github.com/spf13/cobra"
)

// extractKeysFromSchema recursively extracts all field keys from a schema section
func extractKeysFromSchema(schema map[string]any, prefix string, keys *[]string) {
	if properties, ok := schema["properties"].(map[string]any); ok {
		for key, value := range properties {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}
			*keys = append(*keys, fullKey)

			if valueMap, ok := value.(map[string]any); ok {
				extractKeysFromSchema(valueMap, fullKey, keys)
			}
		}
	}
}

// getIACKeys returns all available iac keys excluding those that exist in opencenter
func getIACKeys() []string {
	schema, err := config.GenerateSchema(false)
	if err != nil {
		return []string{"Error loading schema"}
	}

	var schemaData map[string]any
	if err := json.Unmarshal(schema, &schemaData); err != nil {
		return []string{"Error parsing schema"}
	}

	properties, ok := schemaData["properties"].(map[string]any)
	if !ok {
		return []string{"Error accessing schema properties"}
	}

	// Extract all iac keys
	var iacKeys []string
	if iacSchema, ok := properties["iac"].(map[string]any); ok {
		extractKeysFromSchema(iacSchema, "", &iacKeys)
	}

	// Extract all opencenter keys (flattened, just key names without paths)
	var opencenterKeys []string
	if opencenterSchema, ok := properties["opencenter"].(map[string]any); ok {
		extractKeysFromSchema(opencenterSchema, "", &opencenterKeys)
	}

	// Create a set of opencenter key names (just the final part after last dot)
	opencenterKeySet := make(map[string]bool)
	for _, key := range opencenterKeys {
		parts := strings.Split(key, ".")
		keyName := parts[len(parts)-1]
		opencenterKeySet[keyName] = true
	}

	// Filter iac keys to exclude those that exist in opencenter
	var filteredIACKeys []string
	for _, key := range iacKeys {
		parts := strings.Split(key, ".")
		keyName := parts[len(parts)-1]
		if !opencenterKeySet[keyName] {
			filteredIACKeys = append(filteredIACKeys, "iac."+key)
		}
	}

	sort.Strings(filteredIACKeys)
	return filteredIACKeys
}

// showUpdateHelp displays enhanced help with available IAC schema values
func showUpdateHelp(cmd *cobra.Command) error {
	// Show standard help first
	fmt.Fprint(cmd.OutOrStdout(), cmd.Long)
	if cmd.Long == "" {
		fmt.Fprint(cmd.OutOrStdout(), cmd.Short)
	}
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout())

	fmt.Fprintf(cmd.OutOrStdout(), "Usage:\n  %s\n\n", cmd.UseLine())

	// Show flags
	if cmd.Flags().HasFlags() {
		fmt.Fprintln(cmd.OutOrStdout(), "Flags:")
		fmt.Fprintln(cmd.OutOrStdout(), cmd.Flags().FlagUsages())
	}

	// Show global flags
	if cmd.HasParent() && cmd.Parent().PersistentFlags().HasFlags() {
		fmt.Fprintln(cmd.OutOrStdout(), "Global Flags:")
		fmt.Fprintln(cmd.OutOrStdout(), cmd.Parent().PersistentFlags().FlagUsages())
	}

	// Show available IAC schema values
	fmt.Fprintln(cmd.OutOrStdout(), "Available IAC Configuration Keys:")
	fmt.Fprintln(cmd.OutOrStdout(), "  Use any of the following keys with --<key>=<value> format:")
	fmt.Fprintln(cmd.OutOrStdout())

	iacKeys := getIACKeys()
	for _, key := range iacKeys {
		fmt.Fprintf(cmd.OutOrStdout(), "  --%s=<value>\n", key)
	}

	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "Examples:")
	fmt.Fprintln(cmd.OutOrStdout(), "  opencenter cluster update --iac.main.master_count=5")
	fmt.Fprintln(cmd.OutOrStdout(), "  opencenter cluster update my-cluster --iac.main.worker_count=3")
	fmt.Fprintln(cmd.OutOrStdout(), "  opencenter cluster update --iac.main.kubernetes_version=1.30.4")
	fmt.Fprintf(cmd.OutOrStdout(), "  opencenter cluster update my-kind-cluster --%s\n", kindDisableDefaultCNIFlagName)
	fmt.Fprintf(cmd.OutOrStdout(), "  opencenter cluster update my-kind-cluster --%s=false\n", kindDisableDefaultCNIFlagName)

	return nil
}

// newClusterUpdateCmd updates fields in an existing cluster configuration using
// dynamic dotted flags (e.g., --iac.counts.master=3). If a name is not
// provided, the active cluster is used.
func newClusterUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update fields in an existing cluster configuration",
		Args:  cobra.MaximumNArgs(1),
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			flagOverrides, err := parseUpdateFlagOverrides(cmd)
			if err != nil {
				return err
			}

			// If no cluster name provided and no update flags, show enhanced help
			if len(args) == 0 && len(flagOverrides) == 0 && !cmd.Flags().Changed(kindDisableDefaultCNIFlagName) {
				return showUpdateHelp(cmd)
			}

			// Resolve cluster name from args or active cluster
			name, err := resolveClusterName(args, true)
			if err != nil {
				return err
			}

			cfg, _, _, _, err := loadNativeV2ConfigWithIdentifier(cmd.Context(), name)
			if err != nil {
				return fmt.Errorf("failed to load cluster %s: %w", name, err)
			}

			if enabled, changed, err := kindDisableDefaultCNIValue(cmd, cfg.OpenCenter.Infrastructure.Provider); err != nil {
				return err
			} else if changed {
				flagOverrides = append(flagOverrides, "--"+kindDisableDefaultCNIPath+"="+strconv.FormatBool(enabled))
			}

			for _, override := range flagOverrides {
				nameValue := strings.TrimPrefix(override, "--")
				key, value, ok := strings.Cut(nameValue, "=")
				if !ok {
					return fmt.Errorf("invalid override format: %s", override)
				}
				if err := setField(cfg, key, value); err != nil {
					return fmt.Errorf("error setting config from flag '%s': %w", key, err)
				}
			}

			// Optional strict validation
			strict, _ := cmd.Flags().GetBool("strict")
			if strict {
				if err := validateNativeV2Config(cfg); err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), err)
					return fmt.Errorf("validation failed")
				}
			}

			if err := saveNativeV2Config(cmd.Context(), cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated cluster configuration %s\n", name)
			return nil
		},
	}
	cmd.Flags().Bool("strict", false, "fail if the resulting configuration is not valid")
	cmd.Flags().Bool(kindDisableDefaultCNIFlagName, false, "disable Kind's default CNI so cluster networking is managed by openCenter")
	return cmd
}

func parseUpdateFlagOverrides(cmd *cobra.Command) ([]string, error) {
	rawArgs := rawCommandArgs(cmd)
	overrides := make([]string, 0)

	for i := 0; i < len(rawArgs); i++ {
		arg := rawArgs[i]
		if !strings.HasPrefix(arg, "--") || arg == "--" {
			continue
		}

		nameValue := strings.TrimPrefix(arg, "--")
		name, value, hasValue := strings.Cut(nameValue, "=")

		if flag := lookupCommandFlag(cmd, name); flag != nil {
			if !hasValue && flag.NoOptDefVal == "" && i+1 < len(rawArgs) && !strings.HasPrefix(rawArgs[i+1], "-") {
				i++
			}
			continue
		}

		if !strings.Contains(name, ".") {
			return nil, fmt.Errorf("unknown flag: --%s", name)
		}

		if !hasValue {
			if i+1 >= len(rawArgs) || strings.HasPrefix(rawArgs[i+1], "-") {
				return nil, fmt.Errorf("flag needs an argument: --%s", name)
			}
			value = rawArgs[i+1]
			i++
		}

		overrides = append(overrides, "--"+name+"="+value)
	}

	return overrides, nil
}
