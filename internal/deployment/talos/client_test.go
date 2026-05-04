package talos

import (
	"context"
	"errors"
	"strings"
	"testing"

	machineapi "github.com/siderolabs/talos/pkg/machinery/api/machine"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
	"google.golang.org/grpc"
)

type talosClientCall struct {
	mode     talosClientMode
	endpoint string
}

type fakeTalosAPIClient struct{}

func (f *fakeTalosAPIClient) ApplyConfiguration(context.Context, *machineapi.ApplyConfigurationRequest, ...grpc.CallOption) (*machineapi.ApplyConfigurationResponse, error) {
	return &machineapi.ApplyConfigurationResponse{}, nil
}

func (f *fakeTalosAPIClient) Bootstrap(context.Context, *machineapi.BootstrapRequest) error {
	return nil
}

func (f *fakeTalosAPIClient) Kubeconfig(context.Context) ([]byte, error) {
	return []byte("apiVersion: v1\n"), nil
}

func (f *fakeTalosAPIClient) Version(context.Context, ...grpc.CallOption) (*machineapi.VersionResponse, error) {
	return &machineapi.VersionResponse{}, nil
}

func (f *fakeTalosAPIClient) Close() error {
	return nil
}

func TestMachineryClientApplyMachineConfigUsesMaintenanceDirectEndpoint(t *testing.T) {
	var calls []talosClientCall
	client := &machineryClient{
		newClient: recordingTalosAPIClientFactory(&calls),
	}

	err := client.ApplyMachineConfig(context.Background(), "198.51.100.11:50000", []byte("machine config"))
	if err != nil {
		t.Fatalf("ApplyMachineConfig() error = %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("client calls = %v, want one call", calls)
	}
	if calls[0].mode != talosClientModeMaintenance {
		t.Fatalf("mode = %q, want %q", calls[0].mode, talosClientModeMaintenance)
	}
	if calls[0].endpoint != "198.51.100.11:50000" {
		t.Fatalf("endpoint = %q, want direct node endpoint", calls[0].endpoint)
	}
}

func TestMachineryClientHealthChecksEachEndpointDirectly(t *testing.T) {
	var calls []talosClientCall
	client := &machineryClient{
		config:    &clientconfig.Config{},
		newClient: recordingTalosAPIClientFactory(&calls),
	}

	err := client.Health(context.Background(), []string{"198.51.100.11:50000", "198.51.100.21:50000"})
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("client calls = %v, want two direct calls", calls)
	}
	for idx, call := range calls {
		if call.mode != talosClientModeAuthenticated {
			t.Fatalf("call %d mode = %q, want %q", idx, call.mode, talosClientModeAuthenticated)
		}
	}
	if got := calls[0].endpoint + "," + calls[1].endpoint; got != "198.51.100.11:50000,198.51.100.21:50000" {
		t.Fatalf("endpoints = %q, want direct node endpoints", got)
	}
}

func TestTalosOperationErrorIncludesEndpointAndRemediation(t *testing.T) {
	err := wrapTalosOperationError("health check", "demo-cp-1", "198.51.100.11:50000", errors.New("dial tcp 198.51.100.11:50000: connect: connection refused"))
	if err == nil {
		t.Fatal("expected wrapped error")
	}
	for _, want := range []string{
		"health check",
		"demo-cp-1",
		"198.51.100.11:50000",
		"OpenStack security group",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not contain %q", err.Error(), want)
		}
	}
}

func recordingTalosAPIClientFactory(calls *[]talosClientCall) talosAPIClientFactory {
	return func(_ context.Context, mode talosClientMode, endpoint string, _ *clientconfig.Config) (talosAPIClient, error) {
		*calls = append(*calls, talosClientCall{mode: mode, endpoint: endpoint})
		return &fakeTalosAPIClient{}, nil
	}
}
