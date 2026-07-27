package gitops

import (
	"strings"
	"testing"

	"github.com/opencenter-cloud/opencenter-cli/internal/config/services"
	v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"
	"github.com/stretchr/testify/require"
)

// The longhorn overlay-files template reads
// (index .OpenCenter.Services "longhorn").Hostname. text/template resolves that
// field against the concrete type stored in the ServiceMap, so the default
// service map must wire longhorn as *services.LonghornConfig. It previously
// stored *services.DefaultServiceConfig, which has no Hostname field, and
// enabling longhorn failed template execution with:
//
//	can't evaluate field Hostname in type *services.DefaultServiceConfig
//
// ServiceMap.UnmarshalYAML types services from the config registry, which maps
// longhorn to LonghornConfig, so the default map disagreeing with the registry
// meant a loaded config and a freshly constructed one behaved differently.

func TestDefaultServiceMapWiresLonghornWithHostname(t *testing.T) {
	cfg, err := v2.NewV2Default("k8s-longhorn", "openstack")
	require.NoError(t, err)

	longhorn, ok := cfg.OpenCenter.Services["longhorn"].(*services.LonghornConfig)
	require.Truef(t, ok, "default service map must wire longhorn as *services.LonghornConfig, got %T. "+
		"The overlay template reads .Hostname, which only exists on LonghornConfig.",
		cfg.OpenCenter.Services["longhorn"])

	require.NotEmpty(t, longhorn.Hostname, "longhorn Hostname default should be set")
	require.Equal(t, "longhorn."+cfg.OpenCenter.Cluster.ClusterFQDN, longhorn.Hostname)
}

func TestLonghornOverlayRendersWhenEnabled(t *testing.T) {
	cfg := enableLonghorn(t, "")

	files, err := longhornOverlayFilesRenderer(cfg)
	require.NoError(t, err, "longhorn overlay renderer must not fail when longhorn is enabled")

	route, ok := files["longhorn-http-route.yaml"]
	require.True(t, ok, "expected longhorn-http-route.yaml")
	require.Contains(t, route, "kind: HTTPRoute")
	require.Contains(t, route, `"longhorn.`+cfg.OpenCenter.Cluster.ClusterFQDN+`"`)
}

func TestLonghornOverlayHonoursCustomHostname(t *testing.T) {
	cfg := enableLonghorn(t, "storage.example.com")

	files, err := longhornOverlayFilesRenderer(cfg)
	require.NoError(t, err)

	require.Contains(t, files["longhorn-http-route.yaml"], `"storage.example.com"`)
}

// TestLonghornPlansCleanly covers the full planning path, which is where the
// original failure surfaced (auto-render service longhorn: overlay-files
// renderer "longhorn": ...).
func TestLonghornPlansCleanly(t *testing.T) {
	cfg := enableLonghorn(t, "")

	actions, err := planClusterAppActions(cfg)
	require.NoError(t, err)

	var found bool
	for _, action := range actions {
		if strings.HasSuffix(action.Output, "longhorn-http-route.yaml") {
			found = true
			break
		}
	}
	require.True(t, found, "planned actions should include the longhorn HTTPRoute overlay file")
}

// enableLonghorn returns a default config with longhorn enabled, optionally
// overriding its hostname.
func enableLonghorn(t *testing.T, hostname string) v2.Config {
	t.Helper()

	cfg, err := v2.NewV2Default("k8s-longhorn", "openstack")
	require.NoError(t, err)

	longhorn, ok := cfg.OpenCenter.Services["longhorn"].(*services.LonghornConfig)
	require.True(t, ok)
	longhorn.Enabled = true
	if hostname != "" {
		longhorn.Hostname = hostname
	}

	return *cfg
}
