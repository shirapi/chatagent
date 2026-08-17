# TODO

- [x] domain層の確定
- [x] infra/notification（Slack）の実装
- [x] interface/controller・interactor/usecaseの実装（チャンネル制限・リアクション追加まで）
- [x] agent/: SSM Parameter Storeからのキャラクター定義・モデルID読み込み
- [x] agent/: AgentCoreへのデプロイ（`agentcore deploy`）
- [x] infra/chat_agent_repository.go（AgentCoreクライアント）の実装: Go側からのAgentCore呼び出し
- [x] usecase.Execの完成: AgentCore呼び出し・スレッド返信
- [x] 結合確認: Slack実チャンネルでの一連の会話動作確認（履歴の保持含む）
- [x] agent/: Webサーチスキル
- [ ] テストの整備（Go: `src/`各層のユニットテスト、Python: `agent/`のユニットテスト）
- [ ] devcontainer・CI/CDの仕上げ: `cfn/template.yaml` の不備修正、AgentCore/SAMデプロイのCI/CD自動化（SSM経由でのRuntime ARN受け渡し含む）
