# infra.md（template.yaml / cfn/ / agent/agentcore/cdk）

## IAM権限

- 最小権限になっているか（`Resource: "*"`を使う場合、本当に回避不能か確認する）
- 権限は対象リソースに絞られているか
