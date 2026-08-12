---
name: go-lint
description: Goコード（src/以下）に対して goimports によるフォーマット自動修正と staticcheck による静的解析を実行する。Goのコード変更後、コミット前、または「lintして」「goimportsかけて」「staticcheckして」と言われたときに使う。
---

# go-lint

以下のコマンドを、devcontainer（`docker compose exec app`）内で実行するだけ。それ以外は何もしない。

```sh
docker compose exec -w /workspace/src app goimports -w . && docker compose exec -w /workspace/src app staticcheck ./...
```

- 許可を求めない。確認しない。追加の調査（git status等）もしない。
- 出力をそのまま報告するだけで、修正の提案・実施はしない。
