Feature:opencenterclusterbasics
Background:
Givenanemptydirectory "<<tmp>>/conf"
Scenario:Initializeaclusterwithdefaults
WhenIrun "opencenterclusterinitdemo --config-dir <<tmp>>/conf"
Thenafile "<<tmp>>/conf/demo.yaml"shouldexist
Andthefile "<<tmp>>/conf/demo.yaml"shouldcontain "cluster_name:demo"
Scenario:Selectthecluster
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
  """
WhenIrun "opencenterclusterselectdemo --config-dir <<tmp>>/conf"
Thenthefile "<<tmp>>/conf/.active"shouldmatchregex "^demo$"
Scenario:Showcurrentcluster
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
  """
AndIrun "opencenterclusterselectdemo --config-dir <<tmp>>/conf"
WhenIrun "opencenterclustercurrent --config-dir <<tmp>>/conf"
Thenstdoutshouldcontain "demo"
Scenario:Listclusters
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
  """
Andafile "<<tmp>>/conf/blue.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:blue
  """
Andafile "<<tmp>>/conf/green.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:green
  """
WhenIrun "opencenterclusterlist --config-dir <<tmp>>/conf"
Thenstdoutshouldcontain:
  """
blue
demo
green
  """
Scenario:ListclustersasJSON
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
  """
Andafile "<<tmp>>/conf/blue.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:blue
  """
Andafile "<<tmp>>/conf/green.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:green
  """
WhenIrun "opencenterclusterlist --json --config-dir <<tmp>>/conf"
Thenstdoutshouldcontain '["blue","demo","green"]'
Scenario:Infoforacluster
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
  """
AndIrun "opencenterclusterselectdemo --config-dir <<tmp>>/conf"
WhenIrun "opencenterclusterinfo --config-dir <<tmp>>/conf"
Thenstdoutshouldcontain "cluster_name:demo"
Scenario:Validateconstraints
WhenIrun "opencenterclusterinitdemo --config-dir <<tmp>>/conf --no-keygen"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclusterinfodemo --validate --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Validationsuccessful."
Scenario:Validateconstraintsfailure
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
gitops:
git_dir: ""
  """
WhenIrun "opencenterclusterinfodemo --validate --config-dir <<tmp>>/conf"
Thenexitcodeshouldbe1
Andstderrshouldcontain "Config.OpenCenter.GitOps.GitDir:fieldisrequired"
Scenario:Preflight
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
  """
AndIrun "opencenterclusterselectdemo --config-dir <<tmp>>/conf"
WhenIrun "opencenterclusterpreflight --config-dir <<tmp>>/conf"
Thenstdoutshouldcontain "Preflightcomplete."
  @hangs @wip
  @wip
Scenario:Bootstrappushesanewcommittoaremoterepository
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
gitops:
git_dir: "<<tmp>>/opencenter-demo"
  """
Andabaregitrepositoryexistsat "<<tmp>>/remote.git"
AndIupdatetheYAML "<<tmp>>/conf/demo.yaml"toset:
  """
opencenter:
gitops:
git_url: "git@localhost:newuser/gitops-repo.git"
  """
AndIrun "opencenterclusterrenderdemo --config-dir <<tmp>>/conf"
WhenIrun "opencenterclusterbootstrapdemo --force --config-dir <<tmp>>/conf"
Thenthecommandshouldsucceed
Andtheremotegitrepositoryshouldcontaina "Bootstrapcommit"
  @hangs
Scenario:Setupwithprovisioning
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
gitops:
git_dir: "<<tmp>>/opencenter-demo"
opentofu:
enabled:true
  """
WhenIrun "opencenterclusterrenderdemo --config-dir <<tmp>>/conf"
Thenafile "<<tmp>>/opencenter-demo/infrastructure/clusters/demo/main.tf"shouldexist
Andafile "<<tmp>>/opencenter-demo/infrastructure/clusters/demo/provider.tf"shouldexist
  @skip @priority7
Scenario:Destroyacluster
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
gitops:
git_dir: "<<tmp>>/opencenter-demo"
  """
WhenIrun "opencenterclusterdestroydemo --force --config-dir <<tmp>>/conf"
Thenthecommandshouldsucceed
Andafile "<<tmp>>/conf/demo.yaml"shouldnotexist
Andadirectory "<<tmp>>/opencenter-demo"shouldnotexist
