package env

import "os"

const (
	LOCAL = "local"
	PROD  = "prod"
)

func GetAppEnv() string {
	return os.Getenv("APP_ENV")
}

func GetSlackOAuthToken() string {
	return os.Getenv("SLACK_OAUTH_TOKEN")
}

func GetSlackChannelID() string {
	return os.Getenv("SLACK_CHANNEL_ID")
}

func GetSlackSigningSecret() string {
	return os.Getenv("SLACK_SIGNING_SECRET")
}

func GetSlackReactionName() string {
	return os.Getenv("SLACK_REACTION_NAME")
}

func GetAgentCoreRuntimeArn() string {
	return os.Getenv("AGENTCORE_RUNTIME_ARN")
}
