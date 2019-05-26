package config

import "os"

type Config struct {
	BranchNameTemplate string
}

func Init() *Config {
	return &Config{
		BranchNameTemplate: envVarOrDefault("WF_BRANCH_NAME_TEMPLATE", "%s/%s"),
	}
}

func envVarOrDefault(envVarName string, defaultValue string) string {
	value := os.Getenv(envVarName)

	if value != "" {
		return value
	}

	return defaultValue
}
