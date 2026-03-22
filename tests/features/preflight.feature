Feature:Providerpreflightchecks
Background:
Givenanemptydirectory "<<tmp>>/conf"
  @preflight
Scenario:Preflightrunsfortheselectedcluster
Givenafile "<<tmp>>/conf/demo.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:demo
  """
AndIrun "opencenterclusterselectdemo --config-dir <<tmp>>/conf"
WhenIrun "opencenterclusterpreflight --config-dir <<tmp>>/conf"
Thenstdoutshouldcontain "Preflightcomplete."
