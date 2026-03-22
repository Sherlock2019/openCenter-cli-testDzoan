#tests/features/cluster_commands.feature
#Expectedbehaviorforthe "opencentercluster"commandgroup:
# -Parent "cluster"printshelp &subcommands
# -list/lsscansconfig_dirfor *.yamlandprintsnames (no .yaml); --jsonoutputsJSON
# -select (byname &interactive),writesactive_pointer;headerwhenCWD ==git_dir
# -info (active &named),humansummary; --jsonprintsfullJSON;helpfulerrors
# -init (non-interactive),doesnotoverwriteunless --force;printsnextsteps
# -setup (materializeembeddedtemplatesintogit_dir),idempotent, --forceoverwrites
# -bootstrap (gitinit/commit/remote/push)withactionableerrorsonmissingprereqs
Feature:Clustercommandgroup
Background:
Givenanemptydirectory "tmp/conf"
Andanemptydirectory "tmp/repo-dev"
Andanemptydirectory "tmp/repo-prod"
Andafile "tmp/conf/dev.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:dev
gitops:
git_dir:tmp/repo-dev
git_url: ""
  """
Andafile "tmp/conf/prod.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:prod
gitops:
git_dir:tmp/repo-prod
git_url: ""
  """
  @help @priority6
Scenario: "opencentercluster"printshelpwithallsubcommands
WhenIrun "opencentercluster --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "list"
Andstdoutshouldcontain "select"
Andstdoutshouldcontain "info"
Andstdoutshouldcontain "init"
Andstdoutshouldcontain "render"
Andstdoutshouldcontain "bootstrap"
  @init @by_name
Scenario:init <cluster-name>createsaYAMLwithdefaults;doesnotoverwriteunless --force
Giventhefile "tmp/conf/newone.yaml"doesnotexist
WhenIrun "opencenterclusterinitnewone --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andthefile "tmp/conf/newone.yaml"shouldexist
Andstdoutshouldcontain "Created"
WhenIrun "opencenterclusterinitnewone --config-dirtmp/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "exists"
WhenIrun "opencenterclusterinitnewone --force --config-dirtmp/conf"
Thentheexitcodeshouldbe0
