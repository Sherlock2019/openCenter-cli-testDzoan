Feature:JSONschemageneration
  @schema
Scenario:GeneratetheclusterconfigurationJSONschema
WhenIrun "opencenterclusterschema --pretty"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain '"title": "opencenterClusterConfiguration"'
Andstdoutshouldcontain '"$schema": "https://json-schema.org/draft/2020-12/schema"'
