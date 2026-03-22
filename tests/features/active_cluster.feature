#tests/features/active_cluster.feature
#Verifiesactiveclusterbehavior:
#1) `select`writestheselectednametotheactivepointerfile.
#2)Commandsthatrelyontheactiveclusterreadthepointer;errorifunset.
#3)WhenCWD ==selectedcluster'sopencenter.gitops.git_dir,theCLIprefixesoutputwith "Activecluster: <name>".
Feature:Activeclusterrules
Background:
Givenanemptydirectory "<<tmp>>/conf"
Andanemptydirectory "<<tmp>>/repo-demo"
  @active_pointer @select
Scenario:Selectingaclusterwritesitsnametotheactivepointer
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
gitops:
git_dir: <<tmp>>/repo-demo
  """
WhenIrun "opencenterclusterselectdemo"
Thentheexitcodeshouldbe0
Andthefile "<<tmp>>/conf/.active"shouldmatchregex "^demo$"
  @active_pointer @unset @error
Scenario:Commandsthatneedtheactiveclusterfailwhennoneisset
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
gitops:
git_dir: <<tmp>>/repo-demo
  """
Andthefile "<<tmp>>/conf/active"doesnotexist
WhenIrun "opencenterclusterinfo"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "noactivecluster"
  @active_pointer @context_header
Scenario:Wheninthecluster'sgitdirectory,outputstartswithanactive-clusterheader
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
gitops:
git_dir: <<tmp>>/repo-demo
  """
Andthedirectory "<<tmp>>/repo-demo"exists
AndIrun "opencenterclusterselectdemo"
Andtheexitcodeshouldbe0
AndIcdto "<<tmp>>/repo-demo"
WhenIrun "opencenterclusterinfo"
Thentheexitcodeshouldbe0
Andthefirstlineofstdoutshouldstartwith "Activecluster:demo"
  @active_pointer @read
Scenario:Commandsreadtheactivepointerwhennoclusternameisprovided
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
gitops:
git_dir: <<tmp>>/repo-demo
  """
AndIrun "opencenterclusterselectdemo"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterinfo"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "demo"
