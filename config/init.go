package config

import "os"

type Config struct {
	BranchNameTemplate string
	MainBranch         string
}

func Init() *Config {
	return &Config{
		BranchNameTemplate: envVarOrDefault("WF_BRANCH_NAME_TEMPLATE", "%s/%s"),
		MainBranch:         envVarOrDefault("WF_MAIN_BRANCH", "main"),
	}
}

func envVarOrDefault(envVarName string, defaultValue string) string {
	value := os.Getenv(envVarName)

	if value != "" {
		return value
	}

	return defaultValue
}
