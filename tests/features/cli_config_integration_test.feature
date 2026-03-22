#CLIConfigurationSystemIntegrationTests
#AdditionalintegrationteststoverifyCLIconfigurationsystemworkswithclustercommands
Feature:CLIConfigurationSystemIntegrationTests
Background:
Givenanemptydirectory "<<tmp>>/conf"
  @integration @config_commands
Scenario:CLIconfigurationcommandsworkend-to-end
WhenIrun "opencenterconfigpath --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "<<tmp>>/conf/config.yaml"
WhenIrun "opencenterconfigview --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "logging:"
Andstdoutshouldcontain "level:warn"
WhenIrun "opencenterconfigsetlogging.leveldebug --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Configurationupdated:logging.level =debug"
WhenIrun "opencenterconfiggetlogging.level --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "debug"
  @integration @global_flags
Scenario:Globalflagsworkwithclustercommands
WhenIrun "opencenterclusterlist --log-leveldebug --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclusterlist --dry-run --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
  @integration @file_operations
Scenario:Configurationsystemcreatesdirectoriesautomatically
Giventhedirectory "<<tmp>>/fresh-config"doesnotexist
WhenIrun "opencenterconfigview --config-dir <<tmp>>/fresh-config"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/fresh-config"shouldexist
Andafile "<<tmp>>/fresh-config/config.yaml"shouldexist
  @integration @precedence
Scenario:Configurationprecedenceworkscorrectly
GivenIrun "opencenterconfigsetlogging.levelinfo --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterconfiggetlogging.level --log-leveldebug --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
  @integration @cross_platform
Scenario:Pathhandlingworksacrossplatforms
WhenIrun "opencenterconfigsetpaths.clustersDir <<tmp>>/custom-clusters --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Configurationupdated"
WhenIrun "opencenterconfiggetpaths.clustersDir --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "<<tmp>>/custom-clusters"
  @integration @error_handling
Scenario:Configurationvalidationworksproperly
WhenIrun "opencenterconfigsetlogging.levelinvalid --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "invalidvalue"
WhenIrun "opencenterconfigsetbehavior.autoConfirmmaybe --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "expectedbooleanvalue"
  @integration @reset
Scenario:Configurationresetworksproperly
GivenIrun "opencenterconfigsetlogging.leveldebug --config-dir <<tmp>>/conf"
Andtheexitcodeshouldbe0
WhenIrun "opencenterconfigreset --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Configurationresettodefaultvalues"
WhenIrun "opencenterconfiggetlogging.level --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "warn"