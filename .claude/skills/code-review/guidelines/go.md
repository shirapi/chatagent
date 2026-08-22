# go.md（src/）

## エラーハンドリング

- エラーを握りつぶしていないか（ログを出さずに`nil`や汎用エラーレスポンスにしてしまっていないか）
- `./scripts/lint.sh`（goimports + staticcheck）を通しているか

## 呼び出し元への応答

- 時間のかかる外部呼び出しの前に、呼び出し元への応答を返しているか（呼び出し元のタイムアウト要件を超えないか）
- エラー発生時、呼び出し元に通知する手段があるか（無反応のまま終わっていないか）

## Clean Architectureのlayer

- `domain/`: 外部API固有の型・JSON構造が漏れていないか
- `infra/`: 外部API固有の変換処理はここに閉じ込められているか
- `usecase/interactor/`: フロー制御のみで、外部SDKを直接呼んでいないか（`domain.XxxRepository`経由になっているか）
