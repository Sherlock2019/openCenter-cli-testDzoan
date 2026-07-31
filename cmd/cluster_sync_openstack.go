package cmd

import (
	"fmt"
	"os"
	"strings"

	cloudopenstack "github.com/opencenter-cloud/opencenter-cli/internal/cloud/openstack"
	openstacksync "github.com/opencenter-cloud/opencenter-cli/internal/cluster/sync/openstack"
	"github.com/opencenter-cloud/opencenter-cli/internal/ui"
	"github.com/spf13/cobra"
)

func newClusterSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize cluster configuration from external systems",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newClusterSyncOpenStackCmd())
	return cmd
}

func newClusterSyncOpenStackCmd() *cobra.Command {
	var (
		cloud           string
		cloudsYAML      string
		subnetID        string
		services        string
		rotateCreds     bool
		noScopeCreds    bool
		matchFlavors    bool
		matchVolumeType bool
	)

	cmd := &cobra.Command{
		Use:   "openstack <cluster>",
		Short: "Synchronize a cluster with an OpenStack clouds.yaml profile",
		Long: `Discover OpenStack values from a clouds.yaml profile and reconcile them into
the cluster configuration. Storage service wiring creates Swift containers and
credentials only when the configuration does not already contain them, unless
--rotate-creds is supplied.`,
		Example: `  # Preview discovery and configuration changes
  opencenter cluster sync openstack acme/prod --os-cloud production --dry-run

  # Reconcile core settings and selected storage services
  opencenter cluster sync openstack acme/prod --os-cloud production \
    --services loki=swift,tempo=s3 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := getGlobalOptions(cmd)
			if strings.TrimSpace(cloud) == "" {
				cloud = strings.TrimSpace(os.Getenv("OS_CLOUD"))
			}
			if cloud == "" {
				return fmt.Errorf("--os-cloud is required (or set OS_CLOUD)")
			}
			if strings.TrimSpace(cloudsYAML) == "" {
				cloudsYAML = cloudopenstack.DefaultCloudsYAMLPath()
			}
			modes, err := openstacksync.ParseServiceModes(services)
			if err != nil {
				return err
			}
			_, _, _, clusterPaths, err := loadNativeV2ConfigWithIdentifier(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("load cluster configuration: %w", err)
			}
			if !opts.DryRun && !opts.Yes {
				prompter := ui.GetPrompter(os.Stdin, cmd.OutOrStdout(), os.Getenv("OPENCENTER_TEST_MODE") != "")
				confirmed, err := prompter.Confirm(cmd.Context(), "This will update the cluster configuration and may create OpenStack containers or credentials. Continue?")
				if err != nil {
					return fmt.Errorf("confirmation prompt failed: %w", err)
				}
				if !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "OpenStack sync cancelled.")
					return nil
				}
			}

			deps, err := cloudopenstack.NewClusterSyncDependencies(cloudsYAML, cloud)
			if err != nil {
				return err
			}
			service, err := openstacksync.NewService(deps)
			if err != nil {
				return err
			}
			result, err := service.Sync(cmd.Context(), openstacksync.Options{
				ConfigPath:       clusterPaths.ConfigPath,
				OSCloud:          cloud,
				SubnetID:         subnetID,
				ServiceModes:     modes,
				ServicesExplicit: cmd.Flags().Changed("services"),
				DryRun:           opts.DryRun,
				RotateCreds:      rotateCreds,
				NoScopeCreds:     noScopeCreds,
				MatchFlavors:     matchFlavors,
				MatchVolumeType:  matchVolumeType,
			})
			if err != nil {
				return err
			}
			return writeOpenStackSyncOutput(cmd, opts.Output, result, opts.DryRun)
		},
	}

	cmd.Flags().StringVar(&cloud, "os-cloud", "", "clouds.yaml profile to use (defaults to OS_CLOUD)")
	cmd.Flags().StringVar(&cloudsYAML, "clouds-yaml", "", "path to clouds.yaml (defaults to OS_CLIENT_CONFIG_FILE or ~/.config/openstack/clouds.yaml)")
	cmd.Flags().StringVar(&subnetID, "subnet-id", "", "internal subnet ID to use when discovery finds multiple subnets")
	cmd.Flags().StringVar(&services, "services", "", "comma-separated storage mappings, for example loki=swift,tempo=s3")
	cmd.Flags().BoolVar(&rotateCreds, "rotate-creds", false, "create replacement service credentials")
	cmd.Flags().BoolVar(&noScopeCreds, "no-scope-creds", false, "allow unscoped Swift application credentials")
	cmd.Flags().BoolVar(&matchFlavors, "match-flavors", false, "select fitting OpenStack flavors for cluster roles")
	cmd.Flags().BoolVar(&matchVolumeType, "match-volume-type", false, "select an available OpenStack volume type")
	return cmd
}

func writeOpenStackSyncOutput(cmd *cobra.Command, format OutputFormat, result *openstacksync.Result, dryRun bool) error {
	if format != OutputText {
		return writeStructuredOutput(cmd, format, result)
	}
	verb := "Synchronized"
	if dryRun {
		verb = "Previewed"
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s OpenStack profile %q for %s.\n", verb, result.OSCloud, result.ConfigPath)
	for _, path := range result.CoreChangedPaths {
		fmt.Fprintf(cmd.OutOrStdout(), "  changed: %s\n", path)
	}
	for _, action := range result.CredentialActions {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", action)
	}
	for _, warning := range result.Warnings {
		fmt.Fprintf(cmd.OutOrStdout(), "  warning: %s\n", warning)
	}
	return nil
}
