Feature:Clusterinitialisation
Asauser,Iwanttoinitialiseanewclusterconfigurationusing
the `init`command,sothatIcanstartdefiningmyclusterlayout.
  @priority8
Scenario:Initialiseanewclusterwithdefaultsettings
WhenIrun "opencenterclusterinittest-cluster"
Thenaclusterconfiguration "test-cluster"shouldexist
Andtheclusterconfiguration "test-cluster"shouldhave "opencenter.cluster.cluster_name"setto "test-cluster"
Andthefileshouldnotcontain "local."
Scenario:Initialiseaclusterandoverridestringsettingsfromflags
WhenIrun "opencenterclusterinittest-cluster --opencenter.gitops.git_dir=/opt/opencenter/test-cluster --opencenter.cluster.kubernetes.master_count=5"
Thenaclusterconfiguration "test-cluster"shouldexist
Andtheclusterconfiguration "test-cluster"shouldhave "opencenter.gitops.git_dir"setto "/opt/opencenter/test-cluster"
Andtheclusterconfiguration "test-cluster"shouldhave "opencenter.cluster.kubernetes.master_count"setto "5"
Scenario:InitgeneratesaSOPSkeywhennotprovided
WhenIrun "opencenterclusterinitdemo --opencenter.gitops.git_dir=<<tmp>>/repo-demo"
Thenafile "~/.config/opencenter/clusters/opencenter/secrets/age/keys/demo-key.txt"shouldexist
Andthefile "~/.config/opencenter/clusters/opencenter/secrets/age/keys/demo-key.txt"shouldcontain "AGE-SECRET-KEY-1"
Scenario:InitdoesnotgenerateaSOPSkeywhendisabled
WhenIrun "opencenterclusterinitdemo2 --opencenter.gitops.git_dir=<<tmp>>/repo-demo2 --no-sops-keygen"
Thenthefile "~/.config/opencenter/clusters/opencenter/secrets/age/keys/demo2-key.txt"shouldnotexist
Andtheclusterconfiguration "demo2"shouldhave "secrets.sops_age_key_file"setto ""
  @priority8
Scenario:Initwithfullschemaincludeslocalreferences
WhenIrun "opencenterclusterinitfull-one --full-schema"
Thenaclusterconfiguration "full-one"shouldexist
Andthefileshouldcontain "local."
Scenario:Initcreatesclusterssubdirectoryandclusterdirectorystructure
WhenIrun "opencenterclusterinitnew-cluster"
Thenadirectory "~/.config/opencenter/clusters"shouldexist
Andadirectory "~/.config/opencenter/clusters/opencenter"shouldexist
Andadirectory "~/.config/opencenter/clusters/opencenter/infrastructure/clusters/new-cluster"shouldexist
Andafile "~/.config/opencenter/clusters/opencenter/.new-cluster-config.yaml"shouldexist
Scenario:Initcreatescluster-specificsecretsdirectorystructure
WhenIrun "opencenterclusterinitsecrets-test"
Thenadirectory "~/.config/opencenter/clusters/opencenter/secrets"shouldexist
Andadirectory "~/.config/opencenter/clusters/opencenter/secrets/age"shouldexist
Andadirectory "~/.config/opencenter/clusters/opencenter/secrets/age/keys"shouldexist
Andafile "~/.config/opencenter/clusters/opencenter/secrets/age/keys/secrets-test-key.txt"shouldexist
Scenario:Forceflagoverwritesexistingclusterdirectory
WhenIrun "opencenterclusterinitforce-test"
AndIrun "opencenterclusterinitforce-test --force"
Thenthecommandshouldsucceed
Andaclusterconfiguration "force-test"shouldexist
Scenario:Initfailswhenclusterdirectoryexistswithoutforceflag
WhenIrun "opencenterclusterinitexisting-test"
AndIrun "opencenterclusterinitexisting-test"
Thenexitcodeshouldbe1
Andstderrshouldcontain "alreadyexists"
Scenario:Configurationloadingworkswithnewdirectorystructureonly
WhenIrun "opencenterclusterinitload-test"
AndIrun "opencenterclusterselectload-test"
Thentheactiveclustershouldbe "load-test"
Andthecommandshouldsucceed
Scenario:SOPSkeygenerationusescluster-specificdirectory
WhenIrun "opencenterclusterinitsops-dir-test"
Thenafile "~/.config/opencenter/clusters/opencenter/secrets/age/keys/sops-dir-test-key.txt"shouldexist
Andtheclusterconfiguration "sops-dir-test"shouldhave "secrets.sops_age_key_file"containing "clusters/opencenter/secrets/age/keys/sops-dir-test-key.txt"
Scenario:Clusterdirectorycreationwithspecialcharactersinname
WhenIrun "opencenterclusterinittest-cluster-123"
Thenadirectory "~/.config/opencenter/clusters/opencenter/infrastructure/clusters/test-cluster-123"shouldexist
Andafile "~/.config/opencenter/clusters/opencenter/.test-cluster-123-config.yaml"shouldexist
Scenario:Initclusterwithorganizationcreatesorganization-baseddirectorystructure
WhenIrun "opencenterclusterinitweb-app --opencenter.meta.organization=dev-team"
Thenadirectory "~/.config/opencenter/clusters/dev-team"shouldexist
Andadirectory "~/.config/opencenter/clusters/dev-team/infrastructure"shouldexist
Andadirectory "~/.config/opencenter/clusters/dev-team/infrastructure/clusters"shouldexist
Andadirectory "~/.config/opencenter/clusters/dev-team/infrastructure/clusters/web-app"shouldexist
Andadirectory "~/.config/opencenter/clusters/dev-team/applications"shouldexist
Andadirectory "~/.config/opencenter/clusters/dev-team/applications/overlays"shouldexist
Andadirectory "~/.config/opencenter/clusters/dev-team/applications/overlays/web-app"shouldexist
Andadirectory "~/.config/opencenter/clusters/dev-team/secrets"shouldexist
Andadirectory "~/.config/opencenter/clusters/dev-team/secrets/age"shouldexist
Andadirectory "~/.config/opencenter/clusters/dev-team/secrets/age/keys"shouldexist
Scenario:Initclusterwithorganizationcreatesclusterconfigurationincorrectlocation
WhenIrun "opencenterclusterinitapi-service --opencenter.meta.organization=prod-team"
Thenafile "~/.config/opencenter/clusters/prod-team/.api-service-config.yaml"shouldexist
Andtheclusterconfiguration "api-service"shouldhave "opencenter.meta.organization"setto "prod-team"
Andtheclusterconfiguration "api-service"shouldhave "opencenter.gitops.git_dir"containing "clusters/prod-team"
Scenario:InitclusterwithorganizationgeneratesSOPSkeyinorganizationstructure
WhenIrun "opencenterclusterinitdatabase --opencenter.meta.organization=data-team"
Thenafile "~/.config/opencenter/clusters/data-team/secrets/age/keys/database-key.txt"shouldexist
Andthefile "~/.config/opencenter/clusters/data-team/secrets/age/keys/database-key.txt"shouldcontain "AGE-SECRET-KEY-1"
Andafile "~/.config/opencenter/clusters/data-team/.sops.yaml"shouldexist
Andthefile "~/.config/opencenter/clusters/data-team/.sops.yaml"shouldcontain "creation_rules:"
Andtheclusterconfiguration "database"shouldhave "secrets.sops_age_key_file"containing "data-team/secrets/age/keys/database-key.txt"
Scenario:Initclusterwithoutorganizationusesopencenterorganization
WhenIrun "opencenterclusterinitlegacy-app"
Thenadirectory "~/.config/opencenter/clusters/opencenter"shouldexist
Andadirectory "~/.config/opencenter/clusters/opencenter/infrastructure/clusters/legacy-app"shouldexist
Andafile "~/.config/opencenter/clusters/opencenter/.legacy-app-config.yaml"shouldexist
Andtheclusterconfiguration "legacy-app"shouldhave "opencenter.meta.organization"setto "opencenter"
Scenario:InitmultipleclustersinsameorganizationshareGitOpsroot
WhenIrun "opencenterclusterinitfrontend --opencenter.meta.organization=web-team"
AndIrun "opencenterclusterinitbackend --opencenter.meta.organization=web-team"
Thenadirectory "~/.config/opencenter/clusters/web-team/infrastructure/clusters/frontend"shouldexist
Andadirectory "~/.config/opencenter/clusters/web-team/infrastructure/clusters/backend"shouldexist
Andafile "~/.config/opencenter/clusters/web-team/.frontend-config.yaml"shouldexist
Andafile "~/.config/opencenter/clusters/web-team/.backend-config.yaml"shouldexist
Andtheclusterconfiguration "frontend"shouldhave "opencenter.gitops.git_dir"containing "clusters/web-team"
Andtheclusterconfiguration "backend"shouldhave "opencenter.gitops.git_dir"containing "clusters/web-team"
Scenario:Initclusterwithorganizationandforceflagoverwritesexisting
WhenIrun "opencenterclusterinittest-service --opencenter.meta.organization=qa-team"
AndIrun "opencenterclusterinittest-service --opencenter.meta.organization=qa-team --force"
Thenthecommandshouldsucceed
Andafile "~/.config/opencenter/clusters/qa-team/.test-service-config.yaml"shouldexist
Andtheclusterconfiguration "test-service"shouldhave "opencenter.meta.organization"setto "qa-team"
Scenario:Initclusterwithorganizationfailswhenclusterexistswithoutforce
WhenIrun "opencenterclusterinitexisting-service --opencenter.meta.organization=ops-team"
AndIrun "opencenterclusterinitexisting-service --opencenter.meta.organization=ops-team"
Thenexitcodeshouldbe1
Andstderrshouldcontain "alreadyexistsinorganization 'ops-team'"
Scenario:InitclusterwithorganizationcreatesseparateSOPSkeyspercluster
WhenIrun "opencenterclusterinitservice-a --opencenter.meta.organization=shared-team"
AndIrun "opencenterclusterinitservice-b --opencenter.meta.organization=shared-team"
Thenafile "~/.config/opencenter/clusters/shared-team/secrets/age/keys/service-a-key.txt"shouldexist
Andafile "~/.config/opencenter/clusters/shared-team/secrets/age/keys/service-b-key.txt"shouldexist
Andthefile "~/.config/opencenter/clusters/shared-team/secrets/age/keys/service-a-key.txt"shouldcontain "AGE-SECRET-KEY-1"
Andthefile "~/.config/opencenter/clusters/shared-team/secrets/age/keys/service-b-key.txt"shouldcontain "AGE-SECRET-KEY-1"
Scenario:Initclusterwithorganizationandno-sops-keygenflagskipskeygeneration
WhenIrun "opencenterclusterinitno-sops-service --opencenter.meta.organization=security-team --no-sops-keygen"
Thenadirectory "~/.config/opencenter/clusters/security-team/infrastructure/clusters/no-sops-service"shouldexist
Andthefile "~/.config/opencenter/clusters/security-team/secrets/age/keys/no-sops-service-key.txt"shouldnotexist
Andtheclusterconfiguration "no-sops-service"shouldhave "secrets.sops_age_key_file"setto ""
Scenario:Initclusterwithorganizationvalidatesorganizationnameinconfig
WhenIrun "opencenterclusterinitvalidation-test --opencenter.meta.organization=validation-team --strict"
Thenthecommandshouldsucceed
Andtheclusterconfiguration "validation-test"shouldhave "opencenter.meta.organization"setto "validation-team"
