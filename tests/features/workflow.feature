#tests/features/organization_workflow.feature
#Mapstoorganization-basedworkflow:
#  ./opencenterclusterinitdemo --orgmy-org
#  ./opencenterclusterselectmy-org/demo
#  #minimalnetworkchoice:use_octavia=false ->mustsetvrrp_ip
#  ./opencenterclustervalidate
#  ./opencenterclustersetup --render
#  ./opencenterclusterbootstrap
Feature:Organization-basedminimalnetworkworkflow (VRRP)frominittobootstrap
Background:
Givenanemptydirectory "tmp/conf"
Andanemptydirectory "tmp/repo-demo"
  @workflow @init @select @validate @setup @bootstrap @wip
Scenario:Initializewithorg,select,validateVRRPrequirement,rendersetup,andbootstrap
WhenIrun "opencenterclusterinitdemo --orgmy-org --config-dirtmp/conf --force"
Thentheexitcodeshouldbe0
Andthefile "tmp/conf/clusters/my-org/.demo-config.yaml"shouldexist
WhenIrun "opencenterclusterselectmy-org/demo --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andthefile "tmp/conf/active"shouldmatchregex "^my-org/demo\\s*$"
GivenIupdatetheYAML "tmp/conf/clusters/my-org/.demo-config.yaml"toset:
  """
opencenter:
cluster:
domain:example.com
gitops:
git_dir:tmp/repo-demo
git_url:tmp/remote.git
infrastructure:
provider:openstack
cloud:
openstack:
domain: "Default"
application_credential_id: "12345678-1234-1234-1234-123456789012"
application_credential_secret: "test-app-cred-secret"
floating_network_id: "12345678-1234-1234-1234-123456789012"
secrets:
global:
openstack:
application_credential_id: "12345678-1234-1234-1234-123456789012"
application_credential_secret: "test-app-cred-secret"
networking:
use_octavia:false
vrrp_enabled:true
vrrp_ip: ""
"""
WhenIrun "opencenterclustervalidate --config-dirtmp/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "vrrp_ip"
Andstderrshouldcontain "mustbeset"
GivenIupdatetheYAML "tmp/conf/clusters/my-org/.demo-config.yaml"toset:
  """
networking:
vrrp_ip:10.0.0.10
"""
WhenIrun "opencenterclustervalidate --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "validation"
Andstdoutshouldnotcontain "ERROR"
WhenIrun "opencenterclustersetup --render --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Setupcomplete"
Andthedirectory "tmp/repo-demo"shouldexist
Andthedirectory "tmp/repo-demo"shouldcontainafilematching "gitignore"
Andthedirectory "tmp/repo-demo"shouldcontainadirectory "applications"
Givenabaregitrepositoryexistsat "tmp/remote.git"
WhenIrun "opencenterclusterbootstrap --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andthebarerepo "tmp/remote.git"shouldhavebranch "main"