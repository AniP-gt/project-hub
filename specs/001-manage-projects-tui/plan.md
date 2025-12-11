# Implementation Plan: GitHub Projects TUI management

**Branch**: `001-manage-projects-tui` | **Date**: 2025-12-10 | **Spec**: specs/001-manage-projects-tui/spec.md
**Input**: Feature specification from `/specs/001-manage-projects-tui/spec.md`

## Summary

Keyboard-first TUIでGitHub Projectsを操作する。カンバン/テーブル/ロードマップの切替、ステータス更新、フィルタリング、インライン編集・担当者割当を高速に行う。Go + Bubbletea/Lipglossで描画と状態管理を行い、データはgh CLIのJSON出力をパースして扱う。

## Technical Context

**Language/Version**: Go 1.22 (CLI向け最新安定版を想定)
**Primary Dependencies**: Bubbletea, Lipgloss, gh CLI (外部コマンド)、encoding/json
**Storage**: なし（起動時取得・メモリ保持）
**Testing**: go test（unit/interaction）、ゴールデンテキスト比較、gh CLI出力のモック/キャプチャによる統合テスト
**Target Platform**: ターミナル（macOS/Linux）、ローカルでgh認証済み環境
**Project Type**: シングルCLIプロジェクト
**Performance Goals**: ビュー切替・ステータス更新・フィルタ適用の反映が1秒以内（仕様のSC-001/002/003に準拠）
**Constraints**: gh CLI依存（オフライン非対応）、エラー時は非ブロッキング通知で操作継続
**Scale/Scope**: 単一プロジェクト内で数百〜千件程度のアイテムを想定

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Constitutionファイルがプレースホルダーのみで具体的な拘束なし。一般原則として「テスト観点の明確化」「仕様と実装分離」を遵守する。現状違反なし → Proceed.

## Project Structure

### Documentation (this feature)

```text
specs/001-manage-projects-tui/
├── plan.md              # This file (/speckit.plan output)
├── research.md          # Phase 0 output (/speckit.plan)
├── data-model.md        # Phase 1 output (/speckit.plan)
├── quickstart.md        # Phase 1 output (/speckit.plan)
├── contracts/           # Phase 1 output (/speckit.plan)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/projects-tui/        # エントリーポイント (main)
internal/app/            # ルートModel/Update/Viewとメッセージ配線
internal/ui/board/       # カンバン描画・列移動ハンドリング
internal/ui/table/       # テーブル描画・ソート/フォーカス
internal/ui/roadmap/     # ロードマップ（期間軸表示）
internal/ui/components/  # ヘッダー/フッター/共通ウィジェット
internal/github/         # gh CLI呼び出し・JSONパース
internal/state/          # フィルタ/モード/フォーカス状態管理
internal/config/         # 設定・キー割当の読み込み

tests/
├── unit/                # 状態遷移・フォーカス・フィルタロジック
├── integration/         # ghコマンドモックを使ったデータ取得/更新
└── contract/            # 契約テキスト・期待出力のゴールデン
```

**Structure Decision**: Go製CLIの標準構成としてcmd + internal階層を採用。ビュー別にui配下へ分割し、gh連携をinternal/githubで隔離。テストはunit/integration/contractで粒度分けし、ゴールデン比較を容易にする。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| _None_ | - | - |
