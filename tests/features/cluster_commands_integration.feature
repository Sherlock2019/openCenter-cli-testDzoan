Feature:Clustercommandsintegrationwithnewdirectorystructure
Background:
Givenanemptydirectory "<<tmp>>/conf"
Scenario:Clusterselect,info,andvalidateworkwithnewdirectorystructure
WhenIrun "opencenterclusterinitintegration-test --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andafile "<<tmp>>/conf/integration-test.yaml"shouldexist
WhenIrun "opencenterclusterselectintegration-test --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Activeclustersettointegration-test"
Andthefile "<<tmp>>/conf/.active"shouldmatchregex "^integration-test$"
WhenIrun "opencenterclusterinfo --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Activecluster:integration-test"
Andstdoutshouldcontain "cluster_name:integration-test"
WhenIrun "opencenterclusterinfointegration-test --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "cluster_name:integration-test"
WhenIrun "opencenterclusterinfointegration-test --validate --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Validationsuccessful"
  @priority3
Scenario:Clustercommandshandlenon-existentclusterscorrectly
WhenIrun "opencenterclusterselectmissing-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "clustermissing-clusternotfoundinanyorganization"
WhenIrun "opencenterclusterinfomissing-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "clusterconfigurationnotfound:missing-cluster"
WhenIrun "opencenterclustervalidatemissing-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "clusterconfigurationnotfound:missing-cluster"
  @priority3
Scenario:Multipleclustersworkcorrectlywithnewdirectorystructure
WhenIrun "opencenterclusterinitcluster-a --config-dir <<tmp>>/conf"
AndIrun "opencenterclusterinitcluster-b --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclusterlist --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "cluster-a"
Andstdoutshouldcontain "cluster-b"
WhenIrun "opencenterclusterselectcluster-a --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Activeclustersettocluster-a"
WhenIrun "opencenterclusterinfo --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Activecluster:cluster-a"
WhenIrun "opencenterclusterselectcluster-b --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Activeclustersettocluster-b"
WhenIrun "opencenterclusterinfo --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Activecluster:cluster-b"
