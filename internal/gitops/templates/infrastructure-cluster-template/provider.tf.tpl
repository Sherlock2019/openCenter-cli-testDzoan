terraform {
  backend "s3" {
    bucket       = "{{ .OpenTofu.Backend.S3.Bucket | default "opencenter-dev" }}"
    key          = "{{ .OpenTofu.Backend.S3.Key | default "cluster-dev/tfstate/terraform.tfstate" }}"
    region       = "{{ .OpenTofu.Backend.S3.Region | default "us-west-2" }}"
    use_lockfile = true
    encrypt      = true
  }
}
