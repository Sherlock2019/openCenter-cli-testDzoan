allowVolumeExpansion: true
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{ (index .OpenCenter.Services "vsphere-csi").RetainDataStoreName | default "default" }}-retain
parameters:
  datastoreurl: ds:///vmfs/volumes/{{ (index .OpenCenter.Services "vsphere-csi").RetainDataStoreUUID }}/
provisioner: csi.vsphere.vmware.com
reclaimPolicy: Retain
volumeBindingMode: Immediate