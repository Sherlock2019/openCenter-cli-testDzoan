Feature:Destroyclusterssafely
Background:
Givenanemptydirectory "<<tmp>>/conf"
Andanemptydirectory "<<tmp>>/opencenter-demo"
  @destroy @priority7
Scenario:DestroyremovesconfigandGitOpsdirectory
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
