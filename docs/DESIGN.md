# ChatAgent 詳細設計書

全体像・仕様は [README.md](../README.md) を参照。ここでは各コンポーネントの詳細設計（ディレクトリ構成・各層の責務・外部連携時の制約・インフラ構成）を記す。

## ディレクトリ構成

```
.
├── .devcontainer/
│   ├── devcontainer.json
│   ├── Dockerfile               # Go （appサービス用）
│   └── Dockerfile.agent         # Python + uv（agentサービス用）
├── compose.yaml                  # devcontainer用（app / agent の2サービス）
├── cfn/
│   └── template.yaml             # CodePipeline / CodeBuild 定義
├── template.yaml                  # SAMテンプレート（Lambda定義）
├── samconfig.toml                 # SAM CLI設定
├── buildspec.yaml                  # CodeBuild ビルド定義
├── deploy.zsh                      # ローカルからの手動デプロイスクリプト（Go/SAM）
├── docs/
│   ├── DESIGN.md                    # 詳細設計書
│   └── TODO.md                      # 実装状況・残タスク
├── src/                             # Go: Lambda + Slackイベント処理
│   ├── domain/
│   │   ├── chat_agent.go           # infra インターフェース定義
│   │   └── env/
│   │       └── env.go              # 環境変数アクセサ
│   ├── infra/
│   │   ├── di/
│   │   │   └── di.go                # DIコンテナ（各層の生成・結線）
│   │   ├── notification/
│   │   │   └── slack.go             # slack-go ラッパー実装
│   │   └── chat_agent_repository.go # AgentCoreクライアント実装
│   ├── interface/
│   │   └── controller/
│   │       └── chat_agent.go         # ユースケース呼び出し
│   ├── usecase/
│   │   ├── chat_agent.go              # Usecaseインターフェース定義
│   │   └── interactor/
│   │       └── chat_agent.go          # フロー制御
│   ├── main.go                        # Lambdaエントリーポイント
│   ├── go.mod
│   └── go.sum
└── agent/                              # Python: Strandsエージェント（AgentCore CLIで生成）
    ├── deploy.zsh                       # `agentcore deploy` 実行スクリプト
    ├── agentcore/                       # CDK・デプロイ設定一式（Runtime/Memory定義、権限定義）
    └── app/chatagent/
        └── main.py                       # エントリーポイント（BedrockAgentCoreApp）
```

---

## 各層の責務 (Go)

### domain/chat_agent.go

外部依存はすべてinterfaceとして定義し、他層はこれに依存する。SlackのJSON形式などワイヤーフォーマットには依存しない。

```go
type ChatAgentRepository interface {
    Call(ctx context.Context, sessionID, message string) (string, error)
}

type NotificationRepository interface {
    Verify(r *http.Request) (VerifyResult, error)
    AddReaction(ctx context.Context, target NotificationTarget, reaction string) error
    PostReply(ctx context.Context, target NotificationTarget, message string) error
}
```

`sessionID`はAgentCore側のセッション継続に必要な値（後述）。domain層はAgentCore固有の制約（文字数・文字種）を知らなくてよく、渡すのは業務上の識別子（Slackスレッドのタイムスタンプ等）のみ。制約への変換はinfra層の責務。

### usecase/interactor/chat_agent.go

会話フローを制御する。外部依存はすべてinterfaceで受け取る。

1. `Verify`（`NotificationRepository.Verify` を呼ぶだけ）でSlackリクエストを検証し、challenge / MentionEvent を判別（リトライは`Verify`内で無視される）
2. 対象チャンネル（`AllowedChannel`）以外なら 「このチャンネルでは回答できません」 と返信して終了
3. `:reaction:`（`ReactionName`、環境変数で指定）を追加（処理中の目印）
4. `ChatAgentRepository.Call()` でAgentCoreを呼び出し
5. Slackにスレッド返信

### interface/controller/chat_agent.go

- `Usecase.Verify` の呼び出しとchallenge/Mentionの判別
- URLVerification challengeへの対応（そのままレスポンスボディとして返す）
- Mentionがあれば `Usecase.Exec` を呼ぶ

### infra/notification/slack.go

- slack-go を用いたSlack API呼び出しのラッパー
- `Verify`: `slack.NewSecretsVerifier` で署名検証、`slackevents.ParseEvent` でパースし、`domain.VerifyResult` に変換。リトライリクエスト（`X-Slack-Retry-Num`/`X-Slack-Retry-Reason: http_timeout`）はここで無視する
- リアクション追加（`AddReaction`）・スレッド返信（`PostReply`）を担う

### infra/chat_agent_repository.go（AgentCoreクライアント）

`github.com/aws/aws-sdk-go-v2/service/bedrockagentcore` の `InvokeAgentRuntime` で呼び出し。  
呼び出しにあたって、以下の制約への対応が必要。

- **RuntimeSessionId**: 33文字以上・`[a-zA-Z0-9][a-zA-Z0-9-_]*` という制約がある。Slackのスレッドタイムスタンプはそのままでは使えないため、`"chatagent-session-" + (ドットをハイフンに置換したタイムスタンプ)` のように変換する（プレフィックスを足すことで実質的に33文字を満たす）
- **RuntimeUserId（actor_id）**: AgentCore側のセッションは `(RuntimeSessionId, RuntimeUserId)` の組み合わせで一意に決まる（後述）。`RuntimeSessionId`側が既にスレッドごとに一意なので、`RuntimeUserId`は固定値（`"conversation-thread"`）とする。SlackのユーザーIDを使うと、スレッド内でも発言者ごとに会話が分断されるため不可
- **レスポンス**: `Response`は`io.ReadCloser`で、`text/event-stream`（SSE）形式。`main.py`側はBedrock Converse APIのストリーミングイベントを加工せず転送する仕様のため、Go側で`data: `行をJSONパースし、`event.contentBlockDelta.delta.text`を連結して最終的な応答テキストを組み立てる
- IAM権限は `bedrock-agentcore:InvokeAgentRuntime` / `InvokeAgentRuntimeForUser`（RuntimeUserIdを使うため） / `GetAgentRuntime` の3つが必要

---

## Python (agent/)

### main.py の役割

- SSM Parameter Store からキャラクター定義（システムプロンプト）とモデルIDを読み込む（詳細は「設定値の外部管理」を参照）
- Webサーチスキルの設定
- デプロイはAgentCore公式のCLIツール（`agentcore create` でプロジェクト生成、`agentcore deploy` でAWSへデプロイ）を使う。CloudFormationを直接書く方式ではなく、CLIが内部でCDKを生成・実行する

### Webサーチスキルについて

`strands-agents-tools` の `tavily_search` を使う（`TAVILY_API_KEY`が必要）。

APIキーは、システムプロンプト・モデルIDと同じ方式（SSM Parameter Storeからコンテナ起動時に1回読み込み）で取得し、`TAVILY_API_KEY`環境変数として設定する。

> APIキーのような実際の機密値は、本来はAWS Secrets Managerで管理すべきである。Secrets Managerは有料（シークレットごとに月額課金）なため、コストの都合上SSM Parameter Storeに寄せている。今後の検討事項とする。

### セッション（会話履歴）の管理

AgentCore Runtimeのセッションは `(RuntimeSessionId, RuntimeUserId)` の組み合わせで一意に決まり、この単位で会話履歴が保持される。

- 履歴の保持先は2段階ある。プロセス内キャッシュ（コンテナが生きている間のみ有効）と、AgentCore Memory（`agentcore create --memory shortTerm` で作成、`eventExpiryDuration: 30`で30日間永続化、コンテナが再起動しても復元される）
- MemoryリソースはCDKによりRuntimeへ自動的に紐付けられ、`MEMORY_CHATAGENTMEMORY_ID` 環境変数が自動注入される（手動配線は不要）
- Slackのスレッドは15分以上空くこともあり得るため、プロセス内キャッシュだけでは不十分。AgentCore Memoryを使う前提とする

---

## 設定値の外部管理

### 方針

キャラクター定義（システムプロンプト）・モデルID・APIキー（Tavily等）はコードに埋め込まず、**AWS Systems Manager Parameter Store** で管理する。

```
SSM パス:
  /chatagent/AgentCharacter  # システムプロンプト
  /chatagent/AgentModelId    # BedrockモデルID
  /chatagent/TavilyApiKey    # Tavily APIキー（本来はSecretsManagerで管理するのが良い）
```

- AWSコンソールからパラメーターを編集するだけで変更できる
- 読み込みはコンテナ起動時（モジュールレベル）に1回のみ。パラメータを変更しても、次にコンテナが再起動する（アイドルタイムアウトまたは再デプロイ）まで反映されない
- パラメータ自体はCloudFormation化せず、手動でSSMに作成する。`AWS::SSM::Parameter`は`SecureString`を作成できないというCFNの制約があり、機密値を扱う他のパラメータ運用と統一する

> 代替案としてAmazon Bedrock Prompt Management（`AWS::Bedrock::Prompt`）があるが、AgentCoreの公式デプロイ手段（`agentcore` CLI、CloudFormationを使わない命令的デプロイ）とはIaCの単位が噛み合わないため現時点では不採用。今後の検討事項とする。

### 読み込みイメージ（main.py / model/load.py）

```python
import boto3
from botocore.exceptions import ClientError

def _load_system_prompt() -> str:
    try:
        ssm = boto3.client("ssm")
        return ssm.get_parameter(Name="/chatagent/AgentCharacter")["Parameter"]["Value"]
    except ClientError:
        return DEFAULT_SYSTEM_PROMPT  # ローカル動作確認用フォールバック

SYSTEM_PROMPT = _load_system_prompt()  # モジュールレベルで一度だけ実行される
```

---

## devcontainer

Go と Python は別サービス（`app` / `agent`）に分離する。VS Codeは `app` にのみアタッチする方針とし、`agent` を直接操作する場合は `docker compose exec agent ...` または「Attach to Running Container」を使う。

| サービス | ランタイム | バージョン | working_dir | 用途 |
|---|---|---|---|---|
| `app` | Go | 1.26 | `/workspace/src` | Slack Bot 開発 |
| `agent` | Python | 3.14 | `/workspace/agent` | Strandsエージェント定義・デプロイ |

---

## 環境変数

| 変数名 | 用途 |
|---|---|
| `APP_ENV` | 実行環境（`local` / `prod`） |
| `SLACK_OAUTH_TOKEN` | Slack Bot トークン |
| `SLACK_CHANNEL_ID` | 応答するチャンネルID |
| `SLACK_SIGNING_SECRET` | Slack 署名検証シークレット |
| `SLACK_REACTION_NAME` | 処理中に追加するリアクション名 |
| `AGENTCORE_RUNTIME_ARN` | 呼び出し先のAgentCore RuntimeのARN |

---

## インフラ・CI/CD

### デプロイ構成

```
GitHub push (main) → CodePipeline → CodeBuild (buildspec.yaml) → sam deploy → Lambda
```

- `template.yaml`: SAMテンプレート。`ChatAgentFunction`（Lambda）、API Gateway（`ChatAgentApi`、アクセスログ有効）、対応するロググループを定義。Lambda実行ロールには`bedrock-agentcore:InvokeAgentRuntime`・`InvokeAgentRuntimeForUser`・`GetAgentRuntime`を`AgentCoreRuntimeArn`（および`/*`配下）に絞って付与する。
- `cfn/template.yaml`: CodePipeline / CodeBuild / IAMロール / アーティファクト用S3バケットを定義するCloudFormationテンプレート。GitHub連携（CodeStarConnections）でmainブランチへのpushをトリガーに自動ビルド・デプロイする。
- `buildspec.yaml`: CodeBuild内で `sam build` → `sam deploy` を実行。
- `samconfig.toml`: SAM CLIの環境別設定（`default` / `prod`）。
- `deploy.zsh`: ローカル環境から手動で `sam deploy` する際の補助スクリプト（Slackトークン等を対話入力）。
- `agent/deploy.zsh`: ローカル環境から手動で `agentcore deploy` する際の補助スクリプト。

### AgentCoreデプロイ時のIAM権限

`agentcore deploy`はCDKでAWSリソースを作成するため、初回はCDK bootstrap（アカウント・リージョンにつき1回、`CDKToolkit`スタックを作成）が必要。bootstrapが作成するリソース群（CDKのデフォルトqualifierに基づく命名規則のS3バケット・ECRリポジトリ・IAMロール）に絞って権限を付与する。

- S3（バケット作成・ポリシー設定）、ECR（リポジトリ作成・削除）、IAM（対象ロールの作成・削除・ポリシーアタッチ）
- KMS（`CreateKey`/`PutKeyPolicy`/`DescribeKey`等）。bootstrapがアセット用にカスタマー管理キーを作成するため必要。キー作成前はARNが決まらないので`CreateKey`系のみ`Resource: "*"`とする
- `bedrock-agentcore:InvokeAgentRuntime`（+`RuntimeUserId`を使うため`InvokeAgentRuntimeForUser`）。Lambda実行ロールと、手元の動作確認用IAMの双方に必要

`agent/agentcore/cdk/lib/cdk-stack.ts`は`agentcore create`が生成するファイル。RuntimeロールへのSSM読み取り権限（`ssm:GetParameter`、`/chatagent/AgentCharacter`・`/chatagent/AgentModelId`）は`agentcore add`等のCLIサブコマンドでは対応できない範囲のため、直接編集して追加する。CLI生成部分と区別できるよう、編集箇所は`BEGIN CUSTOM`/`END CUSTOM`のマーカーで囲む。

### AgentCore Runtime ARNの受け渡し（今後の方針）

`agent/`（AgentCore）と`src/`（Go/SAM）は別々にデプロイされるが、GoのLambdaは`AgentCoreRuntimeArn`パラメータでAgentCore RuntimeのARNを必要とする。CI/CDで自動化する際は以下の方針とする（slackbot_goでもCodeBuild側でSSMから値を読んでいた前例に倣う）。

- `agentcore deploy`後、`agentcore/.cli/deployed-state.json`からARNを抽出し、`aws ssm put-parameter`でSSM Parameter Storeに書き込む
- `template.yaml`の`AgentCoreRuntimeArn`パラメータ型を`AWS::SSM::Parameter::Value<String>`にし、SAMデプロイ時にSSMから自動解決させる
- Step Functions等の大掛かりなオーケストレーションは使わない（CodeBuildのシェルコマンド + SSMの疎結合で十分なため）

### Slack Events APIについての注意

Slack Appの設定で **Socket Mode を有効にしていると、`app_mention` 等の実イベントはHTTP（Request URL）ではなくWebSocket経由で配信され、Lambdaに一切届かない**（URL Verificationのchallengeのみ後方互換でHTTP応答する）。本構成はHTTP（Lambda + Request URL）前提のため、Socket ModeはOFFにする必要がある。
