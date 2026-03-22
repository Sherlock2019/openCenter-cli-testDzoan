#tests/features/integrations.feature
#Verifiesscaffolding &docsforTerraform,Pulumi,Secrets;anddescriptiveerrorhandling.
Feature:Integrationsscaffoldinganderrorhandling
Background:
Givenanemptydirectory "tmp/conf"
Andanemptydirectory "tmp/repo-dev"
Andafile "tmp/conf/dev.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:dev
gitops:
git_dir:tmp/repo-dev
git_url: ""
  """
AndIrun "opencenterclusterselectdev --config-dirtmp/conf"
Andtheexitcodeshouldbe0
  @terraform @scaffold @wip
Scenario:SetupincludesTerraformscaffoldundergitops.git_dir/terraformwithdocumentedtasks
WhenIrun "opencenterclustersetup --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andthedirectory "tmp/repo-dev/terraform"shouldexist
Andthedirectory "tmp/repo-dev/terraform"shouldcontainafilematching "README.md|main\\.tf|variables\\.tf"
Andthedirectory "tmp/repo-dev"shouldcontainafilematching "(^|/)Makefile$"
Andthedirectory "tmp/repo-dev"shouldcontainafilematching "(^|/)(\\.?mise\\.toml|mise\\.json)$"
Andthefile "tmp/repo-dev/README.md"shouldcontain "miserunterraform"
Andthedirectory "tmp/repo-dev/docs"shouldcontainafilematching "terraform(\\.md|/index\\.md)$"
  @pulumi @scaffold @wip
Scenario:SetupincludesoptionalPulumiscaffoldandstackconfigurationdocs
WhenIrun "opencenterclustersetup --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andthedirectory "tmp/repo-dev/infra/pulumi"shouldexist
Andthedirectory "tmp/repo-dev/infra/pulumi"shouldcontainafilematching "Pulumi\\.yaml"
Andthedirectory "tmp/repo-dev/infra/pulumi/stacks"shouldexist
Andthedirectory "tmp/repo-dev/infra/pulumi/stacks"shouldcontainafilematching "(dev|default)\\.(ya?ml)$"
Andthedirectory "tmp/repo-dev/infra/pulumi"shouldcontainafilematching "README\\.md"
Andthefile "tmp/repo-dev/README.md"shouldcontain "miserunpulumi"
Andthedirectory "tmp/repo-dev/docs"shouldcontainafilematching "pulumi(\\.md|/index\\.md)$"
  @secrets @sops @sealedsecrets @wip
Scenario:SetupprovidesSOPSandSealedSecretsexamplesandguidance
WhenIrun "opencenterclustersetup --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andthedirectory "tmp/repo-dev/secrets/sops"shouldexist
Andthedirectory "tmp/repo-dev/secrets/sops"shouldcontainafilematching "README\\.md"
Andthedirectory "tmp/repo-dev/secrets/sops"shouldcontainafilematching "(example|sample).*\\.(ya?ml)$"
Andthefile "tmp/repo-dev/secrets/sops/README.md"shouldcontain "age-keygen"
Andthefile "tmp/repo-dev/secrets/sops/README.md"shouldcontain "sops --encrypt"
Andthedirectory "tmp/repo-dev/secrets/sealed-secrets"shouldexist
Andthedirectory "tmp/repo-dev/secrets/sealed-secrets"shouldcontainafilematching "README\\.md"
Andthedirectory "tmp/repo-dev/secrets/sealed-secrets"shouldcontainafilematching "(sealedsecret|example).*\\.(ya?ml)$"
Andthefile "tmp/repo-dev/secrets/sealed-secrets/README.md"shouldcontain "kubeseal"
Andthefile "tmp/repo-dev/secrets/sealed-secrets/README.md"shouldcontain "controller"
  @errors @infra_collision @wip
Scenario:Setupfailsdescriptivelyifanexpecteddirectorypathisoccupiedbyafile (infra)
Givenafile "tmp/repo-dev/infra"withcontent:
  """
Iamafile,notadirectory.
  """
WhenIrun "opencenterclustersetup --config-dirtmp/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "infra"
Andstderrshouldcontain "notadirectory"
  @errors @secrets_collision @wip
Scenario:Setupfailsdescriptivelyif 'secrets'pathisafile
Givenafile "tmp/repo-dev/secrets"withcontent:
  """
Iamafile,notadirectory.
  """
WhenIrun "opencenterclustersetup --config-dirtmp/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "secrets"
Andstderrshouldcontain "notadirectory"
  @errors @unwritable @wip
Scenario:Setupfailswithnon-zerocodeandhelpfulmessagewhengit_dirisnotwritable
GivenIupdatetheYAML "tmp/conf/dev.yaml"toset:
  """
opencenter:
gitops:
git_dir: /root/forbidden-path
  """
WhenIrun "opencenterclustersetup --config-dirtmp/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "git_dir"
Andstderrshouldcontain "permission"
Andstderrshouldcontain "writable"
  @terraform @disabled @wip
Scenario:Terraformscaffoldisomittedwhenterraform.enabledisfalse
GivenIupdatetheYAML "tmp/conf/dev.yaml"toset:
  """
terraform:
enabled:false
  """
WhenIrun "opencenterclustersetup --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andthedirectory "tmp/repo-dev/infra/terraform"shouldnotexist
  @pulumi @enabled_gate @wip
Scenario:Pulumiscaffoldappearsonlywhenpulumi.enabledistrue
GivenIupdatetheYAML "tmp/conf/dev.yaml"toset:
  """
pulumi:
enabled:false
  """
WhenIrun "opencenterclustersetup --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andthedirectory "tmp/repo-dev/infra/pulumi"shouldnotexist
WhenIupdatetheYAML "tmp/conf/dev.yaml"toset:
  """
pulumi:
enabled:true
  """
AndIrun "opencenterclustersetup --force --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andthedirectory "tmp/repo-dev/infra/pulumi"shouldexist
