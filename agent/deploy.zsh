#!/bin/zsh

# for deploy (agentcore)

yn=""
vared -p "デプロイを開始します。よろしいですか？(y/N): " yn
case "$yn" in [yY]*) ;; *) echo "abort." ; exit ;; esac

agentcore deploy

runtime_arn=$(jq -r '.targets.default.resources.runtimes.chatagent.runtimeArn' agentcore/.cli/deployed-state.json)
aws ssm put-parameter --name /chatagent/AgentCoreRuntimeArn --value "$runtime_arn" --type String --overwrite

# delete
# aws cloudformation delete-stack --stack-name CDKToolkit

return 0
