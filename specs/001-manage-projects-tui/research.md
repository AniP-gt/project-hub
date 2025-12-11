# Research

## Decision: gh CLIをデータソースとして採用しJSONをパース
- Rationale: 既存の認証状態と権限管理を流用でき、Projects v2情報を公式CLI経由で取得できるため開発負荷が低い。ローカル環境での利用が主で要件に合致。
- Alternatives considered: REST/GraphQLを直接呼び出す (認証・スキーマ追従の実装が増えるため不採用); キャッシュを持つ常駐エージェント (初期スコープを超え複雑化するため不採用)。

## Decision: Bubbletea + LipglossでMUVパターンを実装
- Rationale: TUIでの宣言的描画と状態管理が整備され、キーバインドやビュー分割を安全に構成できる。Lipglossで視覚的差分を簡潔に表現可能。
- Alternatives considered: termbox/curses系の低レベル制御 (描画と状態管理の自前実装が大きくなるため不採用); Charmbracelet以外のTUIフレームワーク (コミュニティ規模と事例の多さでBubbleteaを優先)。

## Decision: キーボード操作のレイテンシ目標を1秒以内に設定
- Rationale: 仕様の成功基準(SC-001/002/003)と整合し、GitHub API/CLIの応答時間を含めても体験を損なわない実用値。
- Alternatives considered: 500ms目標 (CLI応答や描画負荷で安定達成が難しい); 2秒許容 (ユーザー体験が低下するため不採用)。

## Decision: オフライン非対応とし、エラーは非ブロッキング通知
- Rationale: gh CLI依存のためオンライン前提。入力を阻害せず再試行を促すことで操作継続性を担保。
- Alternatives considered: ローカルキャッシュ・リプレイ機構 (初期スコープを超え複雑化); 完全ブロッキングエラー (操作を止め体験を悪化させるため不採用)。

## Decision: テスト戦略は単体 + ゴールデン + ghモック統合
- Rationale: Bubbleteaの状態遷移は単体テストが適し、ビュー出力はゴールデンで安定比較がしやすい。gh依存部はモック/録画で再現性を確保。
- Alternatives considered: E2Eのみで確認 (失敗時の切り分けが困難); スナップショット無し (UI回 regressions の検知力低下)。

## Clarifications
- 追加のNEEDS CLARIFICATIONはありません（仕様で必要事項が満たされています）。
