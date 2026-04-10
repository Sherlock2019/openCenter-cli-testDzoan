package sops

import v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"

func newSOPSTestConfig(clusterName, provider, ageKeyFile string) *v2.Config {
	cfg, err := v2.NewV2Default(clusterName, provider)
	if err != nil {
		panic(err)
	}

	cfg.Secrets.SopsAgeKeyFile = ageKeyFile
	return cfg
}
