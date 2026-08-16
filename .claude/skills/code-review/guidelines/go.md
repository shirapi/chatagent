# go.md（src/）

## エラーハンドリング

- エラーを握りつぶしていないか（ログを出さずに`nil`や汎用エラーレスポンスにしてしまっていないか）
- `./scripts/lint.sh`（goimports + staticcheck）を通しているか

## Clean Architectureのlayer

- `domain/`: 外部API固有の型・JSON構造が漏れていないか
- `infra/`: 外部API固有の変換処理はここに閉じ込められているか
- `usecase/interactor/`: フロー制御のみで、外部SDKを直接呼んでいないか（`domain.XxxRepository`経由になっているか）
