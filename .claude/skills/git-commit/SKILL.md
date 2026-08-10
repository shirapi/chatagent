---
name: git-commit
description: ステージ済みの変更をgitでローカルコミットする。git addやpush、PR作成は行わない。「コミットして」「commitして」と言われたときに使う。
---

# git-commit

`git`（`gh`やMCPは使わない）で、ステージ済みの内容をローカルコミットするだけ。`git add`・push・PR作成はしない（ステージは呼び出し元の責任）。

## 手順

1. `git status` / `git diff --staged` でステージ済みの内容を確認する。ステージされているものが無ければ、その旨を伝えて終了する（勝手に`git add`しない）。
2. ステージ済みの変更内容から1〜2文程度のコミットメッセージを作成する（このリポジトリの直近のコミットメッセージのスタイルがあれば揃える）。
3. 作成したコミットメッセージをユーザーに提示し、コミットしてよいか必ず許可を求める。承認が得られるまで `git commit` は実行しない。
4. 承認を得たら、以下の形式でコミットする。メッセージの最後は必ず改行を挟んで `AI-assistant: ClaudeCode (claude-sonnet-5)` を付ける。

```sh
git commit -m "$(cat <<'EOF'
<コミットメッセージ>

AI-assistant: ClaudeCode (claude-sonnet-5)
EOF
)"
```

5. `git status` で成功を確認する。
