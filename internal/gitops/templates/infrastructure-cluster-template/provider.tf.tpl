terraform {
  backend "s3" {
    bucket       = "1342585-prosys"
    key          = "prosys-dev/tfstate/terraform.tfstate"
    region       = "us-west-2"
    use_lockfile = true
    encrypt      = true
  }
}
