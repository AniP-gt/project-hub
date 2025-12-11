# Tasks: GitHub Projects TUI management

**Input**: Design documents from `/specs/001-manage-projects-tui/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: テストは任意（仕様で必須指定なし）。必要に応じて各タスク完了後に go test / ゴールデン比較を追加。

**Organization**: タスクはユーザーストーリー単位で独立実装・テストできるよう編成。

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: プロジェクト初期化と基本構成

- [X] T001 プロジェクト構成ディレクトリを作成（`cmd/projects-tui`, `internal/{app,ui,github,state,config}`）
- [X] T002 Goモジュール依存を宣言（Bubbletea, Lipgloss, gh CLI連携, encoding/json）`go.mod`
- [X] T003 [P] エントリーポイントをスキャフォールドしプロジェクトID引数を受け取る `cmd/projects-tui/main.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 全ストーリーに共通の基盤。完了前にユーザーストーリー着手不可。

- [X] T004 共通データ型（Project/Item/ViewContext/Timeline）と状態保持を定義 `internal/state/types.go`
- [X] T005 [P] gh CLI呼び出しとJSONパースのクライアントインターフェースを用意 `internal/github/client.go`
- [X] T006 [P] Bubbletea Model/Update/Viewのルート配線と初期状態ロードを実装 `internal/app/app.go`
- [X] T007 [P] キーバインドとモード管理（normal/filter/edit）を定義 `internal/state/keymap.go`
- [X] T008 エラー/通知コンポーネントを追加し非ブロッキング表示を用意 `internal/ui/components/notifications.go`

**Checkpoint**: 基盤準備完了後、ユーザーストーリー着手可能。

---

## Phase 3: User Story 1 - キーボードでビュー横断とステータス更新 (Priority: P1) 🎯 MVP

**Goal**: キーボードのみでビュー切替とステータス更新を完了できる。
**Independent Test**: プロジェクトを開き、1/b,2/t,3/rでビュー切替、h/lでステータス移動が即時反映することを確認。

### Implementation for User Story 1

- [X] T009 [P] [US1] ビュー切替とフォーカス維持のUpdate処理を実装 `internal/app/update_view_switch.go`
- [X] T010 [P] [US1] カンバンビューの列描画とアイテムフォーカス移動を実装 `internal/ui/board/view.go`
- [X] T011 [P] [US1] テーブルビューの行表示とフォーカス同期を実装 `internal/ui/table/view.go`
- [X] T012 [P] [US1] ロードマップビューの期間軸表示とフォーカス位置を実装 `internal/ui/roadmap/view.go`
- [X] T013 [US1] ステータス左右移動とgh更新コマンド送出を実装 `internal/app/update_status.go`
- [X] T014 [US1] ヘッダー/フッターにプロジェクト名・現在ビュー・主要キーバインドを表示 `internal/ui/components/header.go`

**Checkpoint**: US1単体で切替とステータス更新が完結・検証可能。

---

## Phase 4: User Story 2 - 高速フィルタリングで目的のアイテムに絞り込み (Priority: P2)

**Goal**: `/` からのフィルタ入力で表示を即時絞り込み・解除できる。
**Independent Test**: `/` でフィルタ条件入力→表示が1秒以内に反映、クリアで元に戻ることを確認。

### Implementation for User Story 2

- [X] T015 [US2] フィルタ入力モードとクエリ文字列保持を実装 `internal/app/update_filter.go`
- [X] T016 [P] [US2] フィルタパーサと適用ロジック（label/assignee/status）を実装 `internal/state/filter.go`
- [X] T017 [P] [US2] フィルタ結果の適用とゼロ件時の空状態表示を実装 `internal/ui/components/empty_state.go`
- [X] T018 [US2] フィルタ適用/解除のインジケータをヘッダー/フッターに追加 `internal/ui/components/header.go`

**Checkpoint**: US2単体でフィルタ適用/解除が検証可能。

---

## Phase 5: User Story 3 - インライン編集と担当者割り当て (Priority: P3)

**Goal**: フォーカス中アイテムのタイトル/説明編集と担当者割当をキーボードのみで完結。
**Independent Test**: `i`または`Enter`で編集→保存、`a`で担当者確定が即時反映することを確認。

### Implementation for User Story 3

- [X] T019 [US3] 編集モード遷移とタイトル/説明の保存・キャンセル処理を実装 `internal/app/update_edit.go`
- [X] T020 [P] [US3] ghパッチ呼び出しでタイトル/説明更新を実装 `internal/github/client.go`
- [X] T021 [US3] 担当者選択UIと状態更新・gh反映を実装 `internal/app/update_assignee.go`
- [X] T022 [P] [US3] 編集・割当操作のUI反映（ヒント/トースト）を実装 `internal/ui/components/edit_panel.go`

**Checkpoint**: US3単体で編集と割当が完結・検証可能。

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: 横断的な仕上げと品質向上

- [X] T023 ドキュメント/quickstart整備と操作手順の更新 `specs/001-manage-projects-tui/quickstart.md`
- [X] T024 [P] 操作レイテンシ計測/ログを追加し1秒目標を検証 `internal/app/metrics.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- Setup (Phase 1) → Foundational (Phase 2) → US1 (Phase 3) → US2 (Phase 4) → US3 (Phase 5) → Polish (Final)
- US1はMVP。US2/US3はFoundational完了後に並行着手可だが、表示・編集はUS1のビュー基盤に依存。

### User Story Dependencies

- US1: 基盤のみ依存。ビュー切替/ステータス更新を確立。
- US2: US1のビュー基盤とフィルタ適用先のUIに依存。
- US3: US1のフォーカス/ビュー基盤とghクライアントに依存。

### Within Each User Story

- モデル/状態 → Update処理 → UI反映の順。
- gh呼び出しを伴う処理はUI適用前にモックしやすい構造に分離。

### Parallel Opportunities

- Setup: T003は依存少なく並行可。
- Foundational: T005/T006/T007は独立作業可。
- US1: T010/T011/T012は異なるビューで並行可能。
- US2: T016/T017は並行可能。
- US3: T020/T022は並行可能。
- Polish: T024は他タスク完了後に計測のみで並行可。

---

## Parallel Execution Examples

- US1: `update_view_switch.go`, `board/view.go`, `table/view.go`, `roadmap/view.go` を並行実装し、最後に `update_status.go` で統合。
- US2: フィルタパーサ `internal/state/filter.go` と空状態UI `internal/ui/components/empty_state.go` を並行で進め、`update_filter.go` で結合。
- US3: ghパッチ処理 `internal/github/client.go` とUI反映 `internal/ui/components/edit_panel.go` を並行で進め、`update_edit.go`/`update_assignee.go` で統合。

---

## Implementation Strategy

### MVP First (User Story 1 Only)
1. Setup → Foundational → US1完了でMVP。ビュー切替とステータス更新を1秒以内で成立させる。
2. US1の独立テストを実施しデモ可能な状態を確認。

### Incremental Delivery
1. US1でMVP達成後、US2フィルタリングを追加し単独検証。
2. US3の編集/割当を追加し単独検証。
3. 各段階で quickstart に沿って go test / 手動操作確認を行う。

### Parallel Team Strategy
- Foundational完了後、US1/US2/US3を別担当で並行実装可能。ビュー/状態の共通IF（state/types, github/client）を固定してから着手。

---

## Notes

- [P]は依存の少ないファイルで並行可能な作業。
- 各タスクは明示パス付きで、単独でLLMが完了できる具体性を確保。
- テストが必要な場合は該当タスク後に go test やゴールデン比較を追加する。
