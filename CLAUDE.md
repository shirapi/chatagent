# CLAUDE.md

## 作業の進め方

- **一気に実装しない。ステップごとに進める。** 複数ファイルにまたがる変更や設計判断を伴う実装は、まず「何をするか」を提案し、ユーザーの指示を仰いでから着手する。
- 1ステップの作業が終わったら次のステップに進む前に区切る。まとめて全部作ろうとしない。
- 実装方針に選択肢がある場合（型の持たせ方、layer分担など）は、独断で決めずに提案して確認する。

## アーキテクチャ原則（Clean Architecture）

- **domain層は外部インターフェースのワイヤーフォーマットに依存しない。**
  例えば Slack Events API の JSON構造（`type` / `challenge` / `event.channel` など）はSlackという外部実装の都合であり、domain層が知るべきことではない。
  - domain層: 業務上必要な情報のみを持つ抽象的な型・interfaceを定義する（例: `NotificationRepository.Verify(...) (*VerifiedEvent, error)`）。
  - infra層: 外部サービス固有のJSON構造体・レスポンス形式はここに閉じ込め、domainの抽象型へ変換してから返す。
- interfaceは domain に置き、実装は infra に置く。他の層（interactor/interface）は domain の interface にのみ依存する。
