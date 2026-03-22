Feature:SOPSagekeygeneration
Scenario:Generateanagekeytoaspecificpath
WhenIrun "opencentersecretskeysgenerate --key-file <<tmp>>/age.keys --update-sops-config=false"
Thentheexitcodeshouldbe0
Thenafile "<<tmp>>/age.keys"shouldexist
Andthefile "<<tmp>>/age.keys"shouldcontain "AGE-SECRET-KEY-1"
