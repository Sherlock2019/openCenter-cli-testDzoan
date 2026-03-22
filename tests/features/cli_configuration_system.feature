#CLIConfigurationSystemIntegrationTests
#TestsforthecomprehensiveCLIconfigurationmanagementsystemincluding:
# -CLIconfigurationcommands (view,set,get,reset,path)
# -Globalflagsandprecedencesystem
# -Organization-basedpathresolution
# -Enhancedclustercommandswithconfigurationintegration
# -Filesystemoperationsanddirectorycreation
# -Configurationprecedenceacrossalllayers
Feature:CLIConfigurationSystemIntegration
Background:
Givenanemptydirectory "<<tmp>>/conf"
Andanemptydirectory "<<tmp>>/custom-config"
  @config @commands
Scenario:CLIconfigurationcommandsworkwithdefaultconfiguration
WhenIrun "opencenterconfigpath --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "<<tmp>>/conf/config.yaml"
WhenIrun "opencenterconfigview --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "logging:"
Andstdoutshouldcontain "level:warn"
Andstdoutshouldcontain "format:text"
Andstdoutshouldcontain "paths:"
Andstdoutshouldcontain "behavior:"
Andstdoutshouldcontain "defaults:"
WhenIrun "opencenterconfiggetlogging.level --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldmatchregex "^warn$"
  @config @commands @set_get
Scenario:CLIconfigurationsetandgetcommandsworkcorrectly
WhenIrun "opencenterconfigsetlogging.leveldebug --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Configurationupdated:logging.level =debug"
WhenIrun "opencenterconfiggetlogging.level --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldmatchregex "^debug$"
WhenIrun "opencenterconfigsetbehavior.autoConfirmtrue --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Configurationupdated:behavior.autoConfirm =true"
WhenIrun "opencenterconfiggetbehavior.autoConfirm --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldmatchregex "^true$"
WhenIrun "opencenterconfigsetpaths.clustersDir <<tmp>>/custom-clusters --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Configurationupdated:paths.clustersDir = <<tmp>>/custom-clusters"
WhenIrun "opencenterconfiggetpaths.clustersDir --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "<<tmp>>/custom-clusters"
  @config @commands @reset
Scenario:CLIconfigurationresetcommandrestoresdefaults
GivenIrun "opencenterconfigsetlogging.leveldebug --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
AndIrun "opencenterconfigsetbehavior.dryRuntrue --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterconfigreset --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Configurationresettodefaultvalues"
WhenIrun "opencenterconfiggetlogging.level --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldmatchregex "^warn$"
WhenIrun "opencenterconfiggetbehavior.dryRun --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldmatchregex "^false$"
  @config @commands @validation
Scenario:CLIconfigurationcommandsvalidateinputvalues
WhenIrun "opencenterconfigsetlogging.levelinvalid --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "invalidvalue"
WhenIrun "opencenterconfigsetbehavior.autoConfirmmaybe --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "expectedbooleanvalue"
WhenIrun "opencenterconfigsetlogging.file.maxSizeabc --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "expectedintegervalue"
WhenIrun "opencenterconfiggetnonexistent.key --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "failedtogetconfigurationvalue"
  @config @global_flags
Scenario:Globalflagsoverrideconfigurationvalues
GivenIrun "opencenterconfigsetlogging.levelinfo --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterlist --log-leveldebug --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclusterlist --dry-run --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
  @config @global_flags @set_flag
Scenario:Global --setflagoverridesconfigurationvalues
GivenIrun "opencenterconfigsetbehavior.autoConfirmfalse --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterconfiggetbehavior.autoConfirm --setbehavior.autoConfirm=true --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclusterlist --setlogging.level=debug --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
  @config @organization @paths
Scenario:Organization-baseddirectorystructureiscreatedcorrectly
WhenIrun "opencenterclusterinitorg-test --opencenter.meta.organization=test-org --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andadirectory "<<tmp>>/conf/clusters/test-org"shouldexist
Andadirectory "<<tmp>>/conf/clusters/test-org/applications"shouldexist
Andadirectory "<<tmp>>/conf/clusters/test-org/applications/overlays"shouldexist
Andadirectory "<<tmp>>/conf/clusters/test-org/infrastructure"shouldexist
Andadirectory "<<tmp>>/conf/clusters/test-org/infrastructure/clusters"shouldexist
Andadirectory "<<tmp>>/conf/clusters/test-org/infrastructure/clusters/org-test"shouldexist
Andadirectory "<<tmp>>/conf/clusters/test-org/secrets"shouldexist
Andadirectory "<<tmp>>/conf/clusters/test-org/secrets/age"shouldexist
Andadirectory "<<tmp>>/conf/clusters/test-org/secrets/age/keys"shouldexist
Andafile "<<tmp>>/conf/clusters/test-org/.org-test-config.yaml"shouldexist
Andafile "<<tmp>>/conf/clusters/test-org/secrets/age/keys/org-test-key.txt"shouldexist
Andafile "<<tmp>>/conf/clusters/test-org/.sops.yaml"shouldexist
  @config @organization @opencenter
Scenario:Clusternameisusedasorganizationwhennonespecified
WhenIrun "opencenterclusterinitdefault-test"
Thentheexitcodeshouldbe0
Andadirectory "~/.config/opencenter/clusters/opencenter"shouldexist
Andadirectory "~/.config/opencenter/clusters/opencenter/infrastructure/clusters/default-test"shouldexist
Andafile "~/.config/opencenter/clusters/opencenter/.default-test-config.yaml"shouldexist
Andtheclusterconfiguration "default-test"shouldhave "opencenter.meta.organization"setto "opencenter"
  @config @organization @multiple_clusters
Scenario:MultipleclustersinsameorganizationshareGitOpsstructure
WhenIrun "opencenterclusterinitcluster-a --opencenter.meta.organization=shared-org --config-dir <<tmp>>/conf"
AndIrun "opencenterclusterinitcluster-b --opencenter.meta.organization=shared-org --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andadirectory "<<tmp>>/conf/clusters/shared-org/infrastructure/clusters/cluster-a"shouldexist
Andadirectory "<<tmp>>/conf/clusters/shared-org/infrastructure/clusters/cluster-b"shouldexist
Andafile "<<tmp>>/conf/clusters/shared-org/.sops.yaml"shouldexist
Andafile "<<tmp>>/conf/clusters/shared-org/secrets/age/keys/cluster-a-key.txt"shouldexist
Andafile "<<tmp>>/conf/clusters/shared-org/secrets/age/keys/cluster-b-key.txt"shouldexist
  @config @cluster_commands @select
Scenario:Enhancedclusterselectcommandshowsorganizationmetadata
GivenIrun "opencenterclusterinitenhanced-test --opencenter.meta.organization=enhanced-org --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterselectenhanced-test --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Activeclustersettoenhanced-test"
WhenIrun "opencenterclusterinfo --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Activecluster:enhanced-test"
Andstdoutshouldcontain "organization:enhanced-org"
  @config @cluster_commands @list @priority3
Scenario:Clusterlistworkswithorganization-basedstructure
GivenIrun "opencenterclusterinitlist-test-a --opencenter.meta.organization=list-org --config-dir <<tmp>>/conf"
AndIrun "opencenterclusterinitlist-test-b --opencenter.meta.organization=list-org --config-dir <<tmp>>/conf"
AndIrun "opencenterclusterinitlist-test-c --opencenter.meta.organization=other-org --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterlist --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "list-org/list-test-a"
Andstdoutshouldcontain "list-org/list-test-b"
Andstdoutshouldcontain "other-org/list-test-c"
  @config @cluster_commands @info @priority6
Scenario:Clusterinfoshowsorganization-basedpaths
GivenIrun "opencenterclusterinitinfo-test --opencenter.meta.organization=info-org --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterinfoinfo-test --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "cluster_name:info-test"
Andstdoutshouldcontain "organization:info-org"
Andstdoutshouldcontain "git_dir:"
Andstdoutshouldcontain "clusters/info-org"
  @config @filesystem @auto_creation
Scenario:Configurationsystemautomaticallycreatesrequireddirectories
Giventhedirectory "<<tmp>>/fresh-config"doesnotexist
WhenIrun "opencenterconfigview --config-dir <<tmp>>/fresh-config"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/fresh-config"shouldexist
Andafile "<<tmp>>/fresh-config/config.yaml"shouldexist
  @config @filesystem @custom_paths
Scenario:Customconfigurationpathsworkcorrectly
GivenIrun "opencenterconfigsetpaths.clustersDir <<tmp>>/custom-clusters --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterinitcustom-path-test --opencenter.meta.organization=custom-org"
Thentheexitcodeshouldbe0
Andadirectory "<<tmp>>/custom-clusters/custom-org/infrastructure/clusters/custom-path-test"shouldexist
Andafile "<<tmp>>/custom-clusters/custom-org/.custom-path-test-config.yaml"shouldexist
  @config @filesystem @permissions
Scenario:Configurationfilesarecreatedwithproperpermissions
WhenIrun "opencenterconfigview --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthefile "<<tmp>>/conf/config.yaml"shouldexist
  @config @precedence @environment
Scenario:Environmentvariablesoverrideconfigurationfilevalues
GivenIrun "opencenterconfigsetpaths.clustersDir <<tmp>>/config-clusters --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
AndIsetenvironmentvariable "OPENCENTER_CONFIG_DIR"to "<<tmp>>/env-config"
WhenIrun "opencenterconfigpath"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "<<tmp>>/env-config/config.yaml"
  @config @precedence @flags_over_config
Scenario:Command-lineflagsoverrideconfigurationfilevalues
GivenIrun "opencenterconfigsetlogging.levelinfo --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterlist --log-leveldebug --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
  @config @precedence @set_flag_highest
Scenario: --setflaghashighestprecedenceforconfigurationvalues
GivenIrun "opencenterconfigsetbehavior.dryRunfalse --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterlist --setbehavior.dryRun=true --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
  @config @precedence @complete_hierarchy
Scenario:Completeprecedencehierarchyworkscorrectly
GivenIrun "opencenterconfigsetlogging.levelwarn --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterlist --log-levelinfo --setlogging.level=debug --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
  @config @cross_platform @path_expansion
Scenario:Pathexpansionworkscorrectlyacrossplatforms
WhenIrun "opencenterconfigsetpaths.clustersDir ~/test-clusters --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Configurationupdated"
WhenIrun "opencenterconfiggetpaths.clustersDir --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldnotcontain "~"
  @config @cross_platform @environment_expansion
Scenario:Environmentvariableexpansionworksinconfiguration
GivenIsetenvironmentvariable "TEST_CLUSTER_DIR"to "<<tmp>>/env-clusters"
WhenIrun "opencenterconfigsetpaths.clustersDir ${TEST_CLUSTER_DIR} --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterconfiggetpaths.clustersDir --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "<<tmp>>/env-clusters"
  @config @error_handling @invalid_config
Scenario:Systemhandlesinvalidconfigurationgracefully
Givenafile "<<tmp>>/conf/config.yaml"withcontent:
  """
invalid:yaml:content:
  """
WhenIrun "opencenterconfigview --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "failedtoloadconfiguration"
  @config @error_handling @missing_permissions
Scenario:Systemhandlespermissionerrorsgracefully
WhenIrun "opencenterconfigview --config-dir /root/no-permission"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "failed"
  @config @error_handling @recovery
Scenario:Systemcanrecoverfromconfigurationerrors
Givenafile "<<tmp>>/conf/config.yaml"withcontent:
  """
logging:
level:invalid-level
  """
WhenIrun "opencenterconfigreset --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Configurationresettodefaultvalues"
WhenIrun "opencenterconfiggetlogging.level --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldmatchregex "^warn$"
  @config @integration @cluster_lifecycle
Scenario:Configurationsystemintegrateswithcompleteclusterlifecycle
GivenIrun "opencenterconfigsetdefaults.provideropenstack --config-dir <<tmp>>/conf"
AndIrun "opencenterconfigsetdefaults.region {{ .OpenCenter.Cluster.ClusterRegion }} --config-dir <<tmp>>/conf"
AndIrun "opencenterconfigsetbehavior.autoConfirmtrue --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterinitlifecycle-test --opencenter.meta.organization=lifecycle-org --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andadirectory "<<tmp>>/conf/clusters/lifecycle-org/infrastructure/clusters/lifecycle-test"shouldexist
Andtheclusterconfiguration "lifecycle-test"shouldhave "opencenter.infrastructure.provider"setto "openstack"
WhenIrun "opencenterclusterselectlifecycle-test --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Activeclustersettolifecycle-test"
WhenIrun "opencenterclusterinfo --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Activecluster:lifecycle-test"
Andstdoutshouldcontain "organization:lifecycle-org"
Andstdoutshouldcontain "provider:openstack"
  @config @integration @gitops_setup
Scenario:ConfigurationsystemworkswithGitOpssetup
GivenIrun "opencenterclusterinitgitops-test --opencenter.meta.organization=gitops-org --opencenter.gitops.git_dir=<<tmp>>/gitops-repo --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterclusterrendergitops-test --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andadirectory "<<tmp>>/gitops-repo"shouldexist
Andadirectory "<<tmp>>/gitops-repo/applications"shouldexist
Andadirectory "<<tmp>>/gitops-repo/infrastructure"shouldexist
