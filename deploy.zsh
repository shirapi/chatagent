#!/bin/zsh

# for develop

rm -rf ./.aws-sam # キャッシュがあるとエラー回避できないときがある

SlackOAuthToken=""
while [[ -z "$SlackOAuthToken" ]]; do
    vared -p "SlackOAuthToken: " -c SlackOAuthToken
done
SlackOAuthToken=${SlackOAuthToken//$'\n'/}

SlackChannelId=""
while [[ -z "$SlackChannelId" ]]; do
    vared -p "SlackChannelId: " -c SlackChannelId
done
SlackChannelId=${SlackChannelId//$'\n'/}

SlackSigningSecret=""
while [[ -z "$SlackSigningSecret" ]]; do
    vared -p "SlackSigningSecret: " -c SlackSigningSecret
done
SlackSigningSecret=${SlackSigningSecret//$'\n'/}

SlackReactionName="eyes"
vared -p "SlackReactionName (default: eyes): " -c SlackReactionName
SlackReactionName=${SlackReactionName//$'\n'/}

echo " "
echo "=== deploy info ==="
echo " "
echo "SlackOAuthToken: $SlackOAuthToken"
echo "SlackChannelId: $SlackChannelId"
echo "SlackSigningSecret: $SlackSigningSecret"
echo "SlackReactionName: $SlackReactionName"
echo " "
echo "=== deploy info ==="
echo " "

yn=""
vared -p "デプロイを開始します。よろしいですか？(y/N): " yn
case "$yn" in [yY]*) ;; *) echo "abort." ; exit ;; esac

# deploy
sam build && sam deploy --stack-name "chatagent" \
    --config-env "prod" \
    --parameter-overrides Env=prod SlackOAuthToken=${SlackOAuthToken} SlackChannelID=${SlackChannelId} SlackSigningSecret=${SlackSigningSecret} SlackReactionName=${SlackReactionName}
# watch
# sam sync --stack-name chatagent --watch
# delete
# sam delete --stack-name chatagent

return 0
