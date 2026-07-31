package sops

import (
	"reflect"
	"testing"
)

func TestOverlayFilesToEncrypt(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     []string
	}{
		{
			name:     "provider without additional secret files",
			provider: "baremetal",
			want: []string{
				"flux-system/gotk-sync.yaml",
				"managed-services/sources/base-repo.yaml",
			},
		},
		{
			name:     "OpenStack",
			provider: "openstack",
			want: []string{
				"flux-system/gotk-sync.yaml",
				"managed-services/sources/base-repo.yaml",
				"secrets/openstack-credentials.yaml",
			},
		},
		{
			name:     "vSphere",
			provider: "vsphere",
			want: []string{
				"flux-system/gotk-sync.yaml",
				"managed-services/sources/base-repo.yaml",
				"secrets/vsphere-credentials.yaml",
				"customer-managed/services/cloud-provider-vsphere/secret.yaml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newSOPSTestConfig("test-cluster", "baremetal", "")
			cfg.OpenCenter.Infrastructure.Provider = tt.provider

			if got := overlayFilesToEncrypt(cfg); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("overlayFilesToEncrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}
