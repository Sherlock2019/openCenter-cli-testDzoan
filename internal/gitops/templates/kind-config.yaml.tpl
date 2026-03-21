{{- $kind := .OpenCenter.Infrastructure.Kind -}}
{{- $nodeImage := $kind.NodeImage | default (printf "kindest/node:v%s" $kind.KubernetesVersion) -}}
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  apiServerAddress: "{{ $kind.APIServerAddress }}"
  apiServerPort: {{ $kind.APIServerPort }}
  disableDefaultCNI: {{ $kind.DisableDefaultCNI }}
  podSubnet: "{{ $kind.PodSubnet }}"
  serviceSubnet: "{{ $kind.ServiceSubnet }}"
nodes:
{{- range $_ := until $kind.ControlPlaneCount }}
  - role: control-plane
    image: "{{ $nodeImage }}"
{{- if $kind.ExtraPortMappings }}
    extraPortMappings:
{{- range $mapping := $kind.ExtraPortMappings }}
      - containerPort: {{ $mapping.ContainerPort }}
        hostPort: {{ $mapping.HostPort }}
        listenAddress: "{{ $mapping.ListenAddress | default "0.0.0.0" }}"
        protocol: "{{ $mapping.Protocol | default "TCP" }}"
{{- end }}
{{- end }}
{{- if $kind.ExtraMounts }}
    extraMounts:
{{- range $mount := $kind.ExtraMounts }}
      - hostPath: "{{ $mount.HostPath }}"
        containerPath: "{{ $mount.ContainerPath }}"
        readOnly: {{ $mount.ReadOnly }}
{{- end }}
{{- end }}
{{- end }}
{{- range $_ := until $kind.WorkerCount }}
  - role: worker
    image: "{{ $nodeImage }}"
{{- if $kind.ExtraMounts }}
    extraMounts:
{{- range $mount := $kind.ExtraMounts }}
      - hostPath: "{{ $mount.HostPath }}"
        containerPath: "{{ $mount.ContainerPath }}"
        readOnly: {{ $mount.ReadOnly }}
{{- end }}
{{- end }}
{{- end }}
