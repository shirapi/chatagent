---
name: go-lint
description: Goコード（src/以下）に対して goimports によるフォーマット自動修正と staticcheck による静的解析を実行する。Goのコード変更後、コミット前、または「lintして」「goimportsかけて」「staticcheckして」と言われたときに使う。
---

# go-lint

以下のスクリプトを実行する。

```sh
./scripts/lint.sh
```

- エラーがあれば、その内容をそのまま報告する。修正の提案・実施はしない。（このスキルでは修正しない）
- エラーが無ければ何もしない。
