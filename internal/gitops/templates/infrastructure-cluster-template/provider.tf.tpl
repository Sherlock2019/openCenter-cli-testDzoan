terraform {
  backend "s3" {
    bucket       = "{{ .OpenTofu.Backend.S3.Bucket | default "1342585-prosys" }}"
    key          = "{{ .OpenTofu.Backend.S3.Key | default "prosys-dev/tfstate/terraform.tfstate" }}"
    region       = "{{ .OpenTofu.Backend.S3.Region | default "us-west-2" }}"
    use_lockfile = {{ .OpenTofu.Backend.S3.UseLockfile | default true }}
    encrypt      = {{ .OpenTofu.Backend.S3.Encrypt | default true }}
  }
}
