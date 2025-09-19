# データモデル: Specifyコマンドシステム

**フェーズ**: 1 - 設計＆契約
**日付**: 2025-09-18
**ステータス**: 定義済み

## エンティティ定義

### 1. FeatureSpecification (機能仕様)

**説明**: 開発する機能の仕様を表現するエンティティ

**フィールド**:
| フィールド名 | 型 | 必須 | 説明 |
|------------|---|-----|------|
| id | string | ✓ | 一意識別子 (例: "001-specify-is-running") |
| branch | string | ✓ | Gitブランチ名 |
| title | string | ✓ | 機能のタイトル |
| description | string | ✓ | 機能の詳細説明 |
| status | enum | ✓ | ステータス (draft, ready, in_progress, completed) |
| createdAt | datetime | ✓ | 作成日時 |
| updatedAt | datetime | ✓ | 更新日時 |
| requirements | []Requirement | ✓ | 機能要件のリスト |
| scenarios | []Scenario | ✓ | 受け入れシナリオのリスト |
| clarifications | []Clarification | ○ | 明確化必要項目のリスト |

**検証ルール**:
- idは数字3桁-kebab-case形式
- branchはGit有効なブランチ名
- titleは最大100文字
- statusの遷移: draft → ready → in_progress → completed

### 2. ImplementationPlan (実装計画)

**説明**: 機能仕様から生成される実装計画

**フィールド**:
| フィールド名 | 型 | 必須 | 説明 |
|------------|---|-----|------|
| specificationId | string | ✓ | 関連する仕様ID |
| technicalContext | TechContext | ✓ | 技術的コンテキスト |
| phases | []Phase | ✓ | 実行フェーズのリスト |
| constitution | ConstitutionCheck | ✓ | 憲法チェック結果 |
| complexity | []ComplexityItem | ○ | 複雑性追跡項目 |
| progress | ProgressTracking | ✓ | 進捗状況 |

**検証ルール**:
- specificationIdは存在する仕様を参照
- phasesは順序付けられ、依存関係を持つ
- constitutionチェックがPASSでない場合、警告を発生

### 3. TechnicalContext (技術コンテキスト)

**説明**: プロジェクトの技術的詳細

**フィールド**:
| フィールド名 | 型 | 必須 | 説明 |
|------------|---|-----|------|
| language | string | ✓ | プログラミング言語とバージョン |
| dependencies | []string | ○ | 主要依存関係 |
| storage | string | ○ | ストレージタイプ |
| testing | string | ✓ | テストフレームワーク |
| platform | string | ✓ | 対象プラットフォーム |
| projectType | enum | ✓ | プロジェクトタイプ (single, web, mobile) |
| performanceGoals | string | ○ | パフォーマンス目標 |
| constraints | []string | ○ | 制約条件 |
| scale | string | ○ | スケール/スコープ |

**検証ルール**:
- languageは有効な言語/バージョン形式
- projectTypeによって必須フィールドが変わる
- performanceGoalsは測定可能な指標を含む

### 4. Phase (実行フェーズ)

**説明**: 実装の各フェーズ

**フィールド**:
| フィールド名 | 型 | 必須 | 説明 |
|------------|---|-----|------|
| number | int | ✓ | フェーズ番号 (0-5) |
| name | string | ✓ | フェーズ名 |
| status | enum | ✓ | ステータス (pending, in_progress, completed) |
| artifacts | []Artifact | ○ | 生成される成果物 |
| dependencies | []int | ○ | 依存するフェーズ番号 |

**検証ルール**:
- numberは0-5の範囲
- 依存関係は循環しない
- statusの遷移: pending → in_progress → completed

### 5. Task (タスク)

**説明**: 実行可能な個別タスク

**フィールド**:
| フィールド名 | 型 | 必須 | 説明 |
|------------|---|-----|------|
| id | int | ✓ | タスクID |
| title | string | ✓ | タスクタイトル |
| description | string | ○ | 詳細説明 |
| type | enum | ✓ | タスクタイプ (test, implement, document) |
| phase | int | ✓ | 所属フェーズ |
| parallel | boolean | ✓ | 並列実行可能フラグ |
| dependencies | []int | ○ | 依存タスクID |
| status | enum | ✓ | ステータス (pending, in_progress, completed, failed) |
| estimatedTime | duration | ○ | 推定時間 |

**検証ルール**:
- TDD原則: testタイプはimplementより先
- 依存関係は同一フェーズ内または前フェーズ
- parallelがtrueの場合、依存関係なし

### 6. Artifact (成果物)

**説明**: 生成される成果物

**フィールド**:
| フィールド名 | 型 | 必須 | 説明 |
|------------|---|-----|------|
| name | string | ✓ | 成果物名 |
| path | string | ✓ | ファイルパス |
| type | enum | ✓ | タイプ (document, code, test, config) |
| format | string | ✓ | フォーマット (md, go, yaml, json) |
| generated | boolean | ✓ | 自動生成フラグ |

**検証ルール**:
- pathは相対パスで指定
- formatは既知のファイル形式
- generatedがtrueの場合、手動編集警告

## 関係性

### エンティティ関係図

```
FeatureSpecification (1) ---> (1) ImplementationPlan
                     (1) ---> (n) Requirement
                     (1) ---> (n) Scenario
                     (1) ---> (n) Clarification

ImplementationPlan   (1) ---> (1) TechnicalContext
                     (1) ---> (n) Phase
                     (1) ---> (1) ConstitutionCheck
                     (1) ---> (n) ComplexityItem
                     (1) ---> (1) ProgressTracking

Phase               (1) ---> (n) Artifact
                    (1) ---> (n) Task

Task                (n) ---> (n) Task (依存関係)
```

## 状態遷移

### FeatureSpecification状態遷移

```
[draft] --> [ready] --> [in_progress] --> [completed]
   |           |              |
   +-----------|------------- +---------> [archived]
```

### Task状態遷移

```
[pending] --> [in_progress] --> [completed]
                    |
                    +----------> [failed]
                                    |
                                    v
                               [pending] (再試行)
```

## データ永続化

### ファイルベースストレージ

**仕様ファイル**: `/specs/{id}/spec.md`
- YAML frontmatterでメタデータ
- Markdownボディで内容

**計画ファイル**: `/specs/{id}/plan.md`
- 同様のYAML + Markdown構造

**タスクファイル**: `/specs/{id}/tasks.md`
- チェックリスト形式のMarkdown

### メタデータ管理

**インデックスファイル**: `/specs/index.json`
```json
{
  "specifications": [
    {
      "id": "001-specify-is-running",
      "branch": "001-specify-is-running",
      "status": "in_progress",
      "lastUpdated": "2025-09-18T10:00:00Z"
    }
  ]
}
```

## 検証とビジネスルール

### 必須検証

1. **一意性**: ID、ブランチ名の重複禁止
2. **参照整合性**: 外部キー参照の検証
3. **状態遷移**: 定義された遷移のみ許可
4. **依存関係**: 循環依存の検出と防止

### ビジネスルール

1. **TDD強制**: テストタスクが実装タスクより先
2. **憲法準拠**: 憲法チェック失敗時の警告/エラー
3. **段階的実行**: フェーズは順序通りに実行
4. **並列化制約**: 依存関係のないタスクのみ並列実行可能

## APIインターフェース

このデータモデルは以下のAPIエンドポイントで操作されます:
- `/specifications` - 仕様のCRUD操作
- `/plans` - 計画の生成と管理
- `/tasks` - タスクの生成と実行管理
- `/artifacts` - 成果物の生成と取得

詳細なAPI契約は`contracts/`ディレクトリに定義されます。