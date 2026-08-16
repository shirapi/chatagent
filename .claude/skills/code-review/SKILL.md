---
name: code-review
description: このリポジトリの変更（PR/ブランチ差分）をレビューする。diffだけでなく呼び出し元・テスト・既存の設計判断まで横断的に調査し、重要度付きで指摘を出す。「レビューして」「PRを見て」と言われたときに使う。
---

# code-review

## 手順

1. **worktreeでレビュー対象ブランチを展開する**
   作業中のブランチを崩さないよう、`git worktree add` でレビュー対象を別ディレクトリに展開してから読む。

   ```sh
   git worktree add /tmp/review-<branch> <branch>
   ```

   レビュー終了後は `git worktree remove /tmp/review-<branch>` で片付ける。

2. **diffだけでなく横断的に調査する（省略禁止）**
   - 変更された関数・型・interfaceを呼び出している既存コード
   - 変更によって壊れる可能性のある既存テスト
   - `docs/DESIGN.md` に書かれている既存の設計判断と矛盾していないか

3. **ガイドラインを読む**
   - `guidelines/cross-cutting.md` は常に読む
   - 変更が `src/` を含む場合は `guidelines/go.md` も読む
   - 変更が `agent/` を含む場合は `guidelines/agent.md` も読む
   - 変更が `template.yaml`・`cfn/`・`agent/agentcore/cdk/` を含む場合は `guidelines/infra.md` も読む

4. **1回目: 気になる点をすべて列挙する**
   重要度を気にせず、思いついた指摘を全部出す。

5. **2回目: 見直して重要度分類する**
   1回目の指摘を見直し、Critical/High/Medium/Lowに分類する。影響範囲の調査結果を踏まえ、本当に必要な指摘だけに絞る。

6. **ガイドラインの自己改善提案**
   同種の指摘が直近のレビューで繰り返し出ている、または既存ガイドラインでカバーされていない観点が見つかった場合、`guidelines/`への追記案を提示する（このスキル自身が勝手に書き換えることはしない、提案のみ）。

## 出力形式

重要度（Critical/High/Medium/Low）ごとに指摘をまとめる。各指摘には「ファイル:行」「問題」「根拠（呼び出し元・テスト・設計判断のどれに基づくか）」を添える。
