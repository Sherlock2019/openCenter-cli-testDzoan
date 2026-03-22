#Configuration:list,select,andinfoflows
Feature:Configurationselectionandinspection
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
  @config @list
Scenario:Listingclustersshowsfilebasenameswithout .yaml
WhenIrun "opencenterclusterlist --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "dev"
Andstdoutshouldcontain "prod"
Andstdoutshouldnotcontain ".yaml"
  @config @list @json
Scenario:ListingclustersasJSON
WhenIrun "opencenterclusterls --json --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "["
Andstdoutshouldcontain '"dev"'
Andstdoutshouldcontain '"prod"'
  @config @list @missing_dir
Scenario:Ifconfig_dirdoesnotexist,createitandprintnoentries
Giventhedirectory "<<tmp>>/fresh-conf"doesnotexist
WhenIrun "opencenterclusterlist --config-dir <<tmp>>/fresh-conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/fresh-conf"shouldexist
Andstdoutshouldbeempty
  @config @select @by_name
Scenario:Selectingaclusterbynameverifiesfileandwritesactive_pointer
WhenIrun "opencenterclusterselectdev --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andthefile "tmp/conf/.active"shouldmatchregex "^dev$"
Andstdoutshouldcontain "Activeclustersettodev"
  @config @select @interactive
Scenario:Selectingaclusterinteractively
WhenIruninteractively "opencenterclusterselect --config-dirtmp/conf"
AndIchoose "prod"fromtheprompt
Thentheexitcodeshouldbe0
Andthefile "tmp/conf/.active"shouldmatchregex "^prod$"
Andstdoutshouldcontain "Activeclustersettoprod"
  @config @select @missing @priority3
Scenario:Selectinganon-existentclusteryieldsahelpfulerror
WhenIrun "opencenterclusterselectmissing --config-dirtmp/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "clusternotfound:clustermissingnotfoundinanyorganization"
Andstderrshouldcontain "opencenterclusterlist"
  @config @select @header_in_git_dir
Scenario:WhenCWDequalsselectedcluster'sgit_dir,subsequentcommandsshowanactiveheader
GivenIrun "opencenterclusterselectdev --config-dirtmp/conf"
Andtheexitcodeshouldbe0
AndIcdto "tmp/repo-dev"
WhenIrun "opencenterclusterinfo --config-dir ../conf"
Thentheexitcodeshouldbe0
Andthefirstlineofstdoutshouldstartwith "Activecluster:dev"
  @config @info @active
Scenario:Infowithoutargumentreadsactive_pointer
GivenIrun "opencenterclusterselectprod --config-dirtmp/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterinfo --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "cluster_name:prod"
Andstdoutshouldcontain "git_dir:tmp/repo-prod"
  @config @info @named @json
Scenario:Infoforanamedclusterwith --jsonprintsfullparsedconfig
WhenIrun "opencenterclusterinfodev --json --config-dirtmp/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain '"cluster_name": "dev"'
Andstdoutshouldcontain '"git_dir": "tmp/repo-dev"'
  @config @info @unset_active
Scenario:Infowithoutactiveclustersetyieldshelpfulmessage
Giventhefile "tmp/conf/active"doesnotexist
WhenIrun "opencenterclusterinfo --config-dirtmp/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "noactivecluster"
  @config @info @invalid_yaml
Scenario:InvalidYAMLissurfacedasaparseerror
Givenafile "tmp/conf/bad.yaml"withcontent:
  """
  :not:yaml:
  """
WhenIrun "opencenterclusterinfobad --config-dirtmp/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "parse"
Andstderrshouldcontain "yaml"
