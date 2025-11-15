{{- $service := index .OpenCenter.Services "vsphere-csi" }}
{{- $version := $service.ImageTag | default "v3.3.0" }}
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {{ $service.Namespace | default "vmware-system-csi" }}
resources:
  - ./vsphere-config-secret.yaml
  - "https://raw.githubusercontent.com/kubernetes-sigs/vsphere-csi-driver/{{ $version }}/manifests/vanilla/vsphere-csi-driver.yaml"

images:
  - name: gcr.io/cloud-provider-vsphere/csi/release/driver
    newName: {{ $service.ImageRepository | default "registry.k8s.io/csi-vsphere" }}/driver
    newTag: {{ $version }}
  - name: gcr.io/cloud-provider-vsphere/csi/release/syncer
    newName: {{ $service.ImageRepository | default "registry.k8s.io/csi-vsphere" }}/syncer
    newTag: {{ $version }}

