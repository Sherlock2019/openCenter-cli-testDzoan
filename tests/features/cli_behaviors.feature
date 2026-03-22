#tests/features/cli_behaviors.feature
#End-to-endCLIbehaviors:
# -listingclusters
# -selecting (byname &interactive)
# -info (active &named)with --jsonand --validate
# -init (non-interactive)incl. --strictfailures
# -setup (materialization,idempotency,forcedoverwrite)
# -bootstrap (gitinit/commit/remote/push)
# -validationrules (Octavia/VRRP/Designate/counts/flavors/git_dir)
Feature:CLIcoreflowsandvalidations
Background:
Givenanemptydirectory "<<tmp>>/conf"
Andanemptydirectory "<<tmp>>/repo-dev"
Andanemptydirectory "<<tmp>>/repo-prod"
Andafile "<<tmp>>/conf/dev.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:dev
kubernetes:
master_count:1
worker_count:1
subnet_pods: "10.42.0.0/16"
subnet_services: "10.43.0.0/16"
loadbalancer_provider:octavia
gitops:
git_dir: "<<tmp>>/repo-dev"
git_url: ""
infrastructure:
provider:openstack
cloud:
openstack:
region: "regionOne"
  """
Andafile "<<tmp>>/conf/prod.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:prod
kubernetes:
master_count:3
worker_count:6
subnet_pods: "10.42.0.0/16"
subnet_services: "10.43.0.0/16"
loadbalancer_provider:octavia
gitops:
git_dir: "<<tmp>>/repo-prod"
git_url: ""
infrastructure:
provider:openstack
cloud:
openstack:
region: "regionOne"
  """
  @list
Scenario:Listingclustersshowsnameswithout .yaml
WhenIrun "opencenterclusterlist"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "dev"
Andstdoutshouldcontain "prod"
Andstdoutshouldnotcontain ".yaml"
  @list @json
Scenario:ListingclustersasJSON
WhenIrun "opencenterclusterls --json"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain '["dev","prod"]'
  @select @by_name
Scenario:Selectingaclusterbyname
WhenIrun "opencenterclusterselectdev"
Thentheexitcodeshouldbe0
Andthefile "<<tmp>>/conf/.active"shouldmatchregex "^dev$"
  @select @interactive
Scenario:Selectingaclusterinteractively
WhenIruninteractively "opencenterclusterselect"
AndIchoose "prod"fromtheprompt
Thentheexitcodeshouldbe0
Andthefile "<<tmp>>/conf/.active"shouldmatchregex "^prod$"
  @info @active
Scenario:Showinginfofortheactivecluster
GivenIrun "opencenterclusterselectdev"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterinfo"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "dev"
Andstdoutshouldcontain "git_dir: <<tmp>>/repo-dev"
  @info @json
Scenario:ShowinginfoforanamedclusterwithJSONoutput
WhenIrun "opencenterclusterinfoprod --json"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain '"cluster_name": "prod"'
Andstdoutshouldcontain '"git_dir": "<<tmp>>/repo-prod"'
  @info @validate
Scenario:Validatingconfigurationwith --validate
WhenIrun "opencenterclusterinfodev --validate"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Validationsuccessful."
Andstdoutshouldnotcontain "ERROR"
  @init @non_interactive
Scenario:Non-interactiveinitcreatesaminimalskeleton
WhenIrun "opencenterclusterinittest-nonint --force"
Thentheexitcodeshouldbe0
Andafile "<<tmp>>/conf/test-nonint.yaml"shouldexist
  @init @strict @wip
Scenario: Non-interactive init fails with --strict when required values missing
When I run "opencenter cluster init bad --strict"
Then the exit code should not be 0
And stderr should contain "validation failed"
  @setup @materialize
Scenario:SetupmaterializesGitOpstemplateintogit_dir
GivenIrun "opencenterclusterselectdev"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterrender"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/repo-dev"shouldcontainafilematching "README.md"
Andthedirectory "<<tmp>>/repo-dev"shouldcontainadirectory "applications"
  @setup @idempotent
Scenario:Runningsetupagainisidempotent
GivenIrun "opencenterclusterselectdev"
Andtheexitcodeshouldbe0
AndIrun "opencenterclusterrender"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterrender"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Rendercomplete"
  @setup @force
Scenario:Forcedsetupoverwritesexistingfiles
GivenIrun "opencenterclusterselectdev"
Andtheexitcodeshouldbe0
Andafile "<<tmp>>/repo-dev/README.md"withcontent:
  """
manualeditthatshouldbereplaced
  """
WhenIrun "opencenterclusterrender"
Thentheexitcodeshouldbe0
Andthefile "<<tmp>>/repo-dev/README.md"shouldnotcontain "manualeditthatshouldbereplaced"
  @bootstrap @priority5 @wip
Scenario:Bootstrappushesthelocalrepotoaremote
Givenabaregitrepositoryexistsat "<<tmp>>/remote.git"
AndIupdatetheYAML "<<tmp>>/conf/dev.yaml"toset:
  """
opencenter:
gitops:
git_dir: "<<tmp>>/repo-dev"
git_url: "<<tmp>>/remote.git"
  """
AndIrun "opencenterclusterselectdev"
Andtheexitcodeshouldbe0
AndIrun "opencenterclusterrender"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterbootstrap"
Thentheexitcodeshouldbe0
Andthebarerepo "<<tmp>>/remote.git"shouldhavebranch "main"
  @validate @git_dir_missing @priority2
Scenario:opencenter.gitops.git_dirmissing ->erroronsetup
Givenafile "<<tmp>>/conf/no-gitdir.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:no-gitdir
gitops:
git_dir: ""
  """
WhenIrun "opencenterclusterrenderno-gitdir"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "opencenter.gitops.git_dirmustbeset"
