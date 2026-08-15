# ChatAgent

> 🚧 開発中。現在の実装状況は [docs/TODO.md](./docs/TODO.md) に記載。

過去 LangChainGo で実装した [Slack Bot](https://github.com/shirapi/slackbot_go) を再構築する。
AIエージェントは AWS AgentCore + Strands で作成し、Slack 側は Go で Clean Architecture により実装する。

詳細な設計は [docs/DESIGN.md](./docs/DESIGN.md) を参照。

## デモ

![デモ](./docs/demo.png)

## 仕様

- Botを使用できるのは特定のチャンネルに限定する。
  対象外チャンネルからのメンションには定型文（`このチャンネルでは回答できません`）で応答する。
- 質問はBotにメンションして開始し、Botは発言したユーザーにメンションしてスレッドで返信する。
  メンション無しでチャンネルに投稿しただけ、スレッドに返信しただけではBotは反応しない。
  ※ Botが無条件に反応すると、チャンネル内・スレッド内でのユーザー同士のやり取りがしづらくなるため。
- 応答に時間がかかるため、処理中であることを示すためにBotは発言に対して先にリアクションをつける（リアクション名は環境変数で指定）。
- スレッドの内容についてBotは記憶がある状態で会話できる（スレッド単位で文脈を保持する）。

## 全体アーキテクチャ

![構成図](./docs/architecture.png)

```
Slack → Go (AWS Lambda) → AgentCore Runtime (AWS) → Strandsエージェント (Python)
                                  ↓
Slack ← Go (AWS Lambda) ←←←←←←←← レスポンス
```

- **Go (Lambda)**: Slackイベント受信・検証・AgentCore呼び出し・Slack返信
- **Python (AgentCore)**: キャラクター定義・Webサーチスキル

### ポイント

- 会話履歴はAgentCoreがセッション単位で管理するため、Go側での履歴取得は不要。
- SlackのスレッドタイムスタンプをAgentCoreのSessionIDとして使用する。
