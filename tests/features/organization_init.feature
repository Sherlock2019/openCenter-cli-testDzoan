Feature:Organization-basedclusterinitialization
Asauser,Iwanttoinitializeclusterswithinorganization-baseddirectorystructures
sothatIcanorganizemyclustersbyteamorenvironment.
Background:
Givenanemptydirectory "<<tmp>>/conf"
Scenario:Initclusterwithorganizationcreatesorganization-baseddirectorystructure
WhenIrun "opencenterclusterinitweb-app --opencenter.meta.organization=dev-team --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andadirectory "<<tmp>>/conf/clusters/dev-team"shouldexist
Andadirectory "<<tmp>>/conf/clusters/dev-team/infrastructure"shouldexist
Andadirectory "<<tmp>>/conf/clusters/dev-team/infrastructure/clusters"shouldexist
Andadirectory "<<tmp>>/conf/clusters/dev-team/infrastructure/clusters/web-app"shouldexist
Andadirectory "<<tmp>>/conf/clusters/dev-team/applications"shouldexist
Andadirectory "<<tmp>>/conf/clusters/dev-team/applications/overlays"shouldexist
Andadirectory "<<tmp>>/conf/clusters/dev-team/applications/overlays/web-app"shouldexist
Andadirectory "<<tmp>>/conf/clusters/dev-team/secrets"shouldexist
Andadirectory "<<tmp>>/conf/clusters/dev-team/secrets/age"shouldexist
Andadirectory "<<tmp>>/conf/clusters/dev-team/secrets/age/keys"shouldexist
Scenario:Initclusterwithorganizationcreatesclusterconfigurationincorrectlocation
WhenIrun "opencenterclusterinitapi-service --opencenter.meta.organization=prod-team --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andafile "<<tmp>>/conf/clusters/prod-team/.api-service-config.yaml"shouldexist
Andtheclusterconfiguration "api-service"shouldhave "opencenter.meta.organization"setto "prod-team"
Andtheclusterconfiguration "api-service"shouldhave "opencenter.gitops.git_dir"containing "clusters/prod-team"
Scenario:InitclusterwithorganizationgeneratesSOPSkeyinorganizationstructure
WhenIrun "opencenterclusterinitdatabase --opencenter.meta.organization=data-team --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andafile "<<tmp>>/conf/clusters/data-team/secrets/age/keys/database-key.txt"shouldexist
Andthefile "<<tmp>>/conf/clusters/data-team/secrets/age/keys/database-key.txt"shouldcontain "AGE-SECRET-KEY-1"
Andafile "<<tmp>>/conf/clusters/data-team/.sops.yaml"shouldexist
Andthefile "<<tmp>>/conf/clusters/data-team/.sops.yaml"shouldcontain "creation_rules:"
Andtheclusterconfiguration "database"shouldhave "secrets.sops_age_key_file"containing "data-team/secrets/age/keys/database-key.txt"
Scenario:Initclusterwithoutorganizationusesclusternameasorganization
WhenIrun "opencenterclusterinitlegacy-app --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andafile "<<tmp>>/conf/legacy-app.yaml"shouldexist
Scenario:InitmultipleclustersinsameorganizationshareGitOpsroot
WhenIrun "opencenterclusterinitfrontend --opencenter.meta.organization=web-team --config-dir <<tmp>>/conf"
AndIrun "opencenterclusterinitbackend --opencenter.meta.organization=web-team --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andadirectory "<<tmp>>/conf/clusters/web-team/infrastructure/clusters/frontend"shouldexist
Andadirectory "<<tmp>>/conf/clusters/web-team/infrastructure/clusters/backend"shouldexist
Andafile "<<tmp>>/conf/clusters/web-team/.frontend-config.yaml"shouldexist
Andafile "<<tmp>>/conf/clusters/web-team/.backend-config.yaml"shouldexist
Andtheclusterconfiguration "frontend"shouldhave "opencenter.gitops.git_dir"containing "clusters/web-team"
Andtheclusterconfiguration "backend"shouldhave "opencenter.gitops.git_dir"containing "clusters/web-team"
Scenario:Initclusterwithorganizationandforceflagoverwritesexisting
WhenIrun "opencenterclusterinittest-service --opencenter.meta.organization=qa-team --config-dir <<tmp>>/conf"
AndIrun "opencenterclusterinittest-service --opencenter.meta.organization=qa-team --force --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andafile "<<tmp>>/conf/clusters/qa-team/.test-service-config.yaml"shouldexist
Andtheclusterconfiguration "test-service"shouldhave "opencenter.meta.organization"setto "qa-team"
Scenario:Initclusterwithorganizationfailswhenclusterexistswithoutforce
WhenIrun "opencenterclusterinitexisting-service --opencenter.meta.organization=ops-team --config-dir <<tmp>>/conf"
AndIrun "opencenterclusterinitexisting-service --opencenter.meta.organization=ops-team --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe1
Andstderrshouldcontain "alreadyexistsinorganization 'ops-team'"
Scenario:InitclusterwithorganizationcreatesseparateSOPSkeyspercluster
WhenIrun "opencenterclusterinitservice-a --opencenter.meta.organization=shared-team --config-dir <<tmp>>/conf"
AndIrun "opencenterclusterinitservice-b --opencenter.meta.organization=shared-team --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andafile "<<tmp>>/conf/clusters/shared-team/secrets/age/keys/service-a-key.txt"shouldexist
Andafile "<<tmp>>/conf/clusters/shared-team/secrets/age/keys/service-b-key.txt"shouldexist
Andthefile "<<tmp>>/conf/clusters/shared-team/secrets/age/keys/service-a-key.txt"shouldcontain "AGE-SECRET-KEY-1"
Andthefile "<<tmp>>/conf/clusters/shared-team/secrets/age/keys/service-b-key.txt"shouldcontain "AGE-SECRET-KEY-1"
Scenario:Initclusterwithorganizationandno-sops-keygenflagskipskeygeneration
WhenIrun "opencenterclusterinitno-sops-service --opencenter.meta.organization=security-team --no-sops-keygen --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andadirectory "<<tmp>>/conf/clusters/security-team/infrastructure/clusters/no-sops-service"shouldexist
Andthefile "<<tmp>>/conf/clusters/security-team/secrets/age/keys/no-sops-service-key.txt"shouldnotexist
Andtheclusterconfiguration "no-sops-service"shouldhave "secrets.sops_age_key_file"setto ""