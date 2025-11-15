{{- $service := index .OpenCenter.Services "vsphere-csi" }}
{{- $secrets := .Secrets.VSphereCsi }}
# This is a template for the vSphere CSI configuration secret.
# The actual secret should be created in the overlay directory with SOPS encryption.
# This file serves as a reference for the expected structure.
#
# To create the actual secret:
# 1. Copy this file to applications/overlays/{{ .ClusterName }}/services/vsphere-csi/
# 2. Replace the placeholder values with actual vSphere credentials
# 3. Encrypt with SOPS: sops -e -i vsphere-config-secret.yaml
#
# Example vSphere CSI configuration:
# [Global]
# cluster-id = "{{ .ClusterName }}"
#
# [VirtualCenter "vcenter.example.com"]
# insecure-flag = "false"
# user = "administrator@vsphere.local"
# password = "your-password"
# port = "443"
# datacenters = "Datacenter1"
apiVersion: v1
kind: Secret
metadata:
  name: vsphere-config-secret
  namespace: {{ $service.Namespace | default "vmware-system-csi" }}
type: Opaque
stringData:
  csi-vsphere.conf: |
    [Global]
    cluster-id = "{{ .ClusterName }}"
    
    {{- if $secrets.VCenterHost }}
    [VirtualCenter "{{ $secrets.VCenterHost }}"]
    insecure-flag = "{{ $secrets.InsecureFlag | default "false" }}"
    user = "{{ $secrets.Username }}"
    password = "{{ $secrets.Password }}"
    port = "{{ $secrets.Port | default "443" }}"
    datacenters = "{{ $secrets.Datacenters }}"
    {{- else }}
    # VirtualCenter configuration required
    # [VirtualCenter "vcenter.example.com"]
    # insecure-flag = "false"
    # user = "administrator@vsphere.local"
    # password = "your-password"
    # port = "443"
    # datacenters = "Datacenter1"
    {{- end }}
