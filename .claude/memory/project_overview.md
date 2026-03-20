---
name: locate-server project overview
description: プロジェクトの概要・構成・技術スタック
type: project
---

locate-server は plocate/gocate を使ったファイル検索 Web サーバー。

**技術スタック**:
- 言語: Go
- Web フレームワーク: gin-gonic/gin
- ログ: op/go-logging
- パイプライン実行: mattn/go-pipeline
- コンテナ: Docker / docker-compose

**ディレクトリ構成**:
- `main.go` — エントリポイント、gin ルーティング、起動設定
- `cmd/api/` — クエリパース (`api.go`, `query.go`)
- `cmd/cache/` — 検索結果キャッシュ
- `cmd/locater/` — locate コマンド実行ロジック (`locater.go`, `command.go`, `frecency.go`)

**エンドポイント**:
- `GET /` — トップページ
- `GET /search` — 検索ページ
- `GET /json` — 検索 API (JSON レスポンス)
- `GET /history` — 検索履歴
- `GET /status` — DB ステータス

**依存コマンド**: `locate`, `gocate`（plocate 系）
**DB パス**: `/var/lib/plocate`
**ログファイル**: `/var/lib/mlocate/locate.log`
**デフォルトポート**: 8080

**Why:** プロジェクト全体像の把握のため記録。
**How to apply:** 新機能追加・変更時にアーキテクチャへの影響を考慮する。
