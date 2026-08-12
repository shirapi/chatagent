#!/bin/zsh

# for deploy (agentcore)

yn=""
vared -p "デプロイを開始します。よろしいですか？(y/N): " yn
case "$yn" in [yY]*) ;; *) echo "abort." ; exit ;; esac

agentcore deploy

# delete
# aws cloudformation delete-stack --stack-name CDKToolkit

return 0
