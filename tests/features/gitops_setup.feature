Feature:GitOpsrepositorysetupbehaviors
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
  @gitops @setup @materialize
Scenario:setupmaterializesembeddedtemplatesintogit_dir
GivenIrun "opencenterclusterselectdev --config-dirtmp/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterrender --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andthedirectory "tmp/repo-dev"shouldcontainafilematching "README.md"
Andstdoutshouldcontain "Rendercomplete"
  @gitops @setup @idempotent @priority2
Scenario:setupisidempotentwhenrunrepeatedly
GivenIrun "opencenterclusterselectdev --config-dirtmp/conf"
Andtheexitcodeshouldbe0
AndIrun "opencenterclusterrender --config-dirtmp/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterrender --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Rendercomplete"
  @gitops @setup @force
Scenario:setup --forceoverwritesexistingfiles
GivenIrun "opencenterclusterselectdev --config-dirtmp/conf"
Andtheexitcodeshouldbe0
Andafile "tmp/repo-dev/README.md"withcontent:
  """
localeditsthatshouldbereplaced
  """
WhenIrun "opencenterclusterrender --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andthefile "tmp/repo-dev/README.md"shouldnotcontain "localeditsthatshouldbereplaced"
  @gitops @setup @missing_prereqs @priority2 @wip
Scenario:setuperrorswhennoactiveclusterorgit_dirismissing
Giventhefile "tmp/conf/active"doesnotexist
WhenIrun "opencenterclusterrender --config-dirtmp/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "noactivecluster"
Givenafile "tmp/conf/nogit.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:nogit
  """
WhenIrun "opencenterclusterrendernogit --config-dirtmp/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "opencenter.gitops.git_dirmustbeset"
