# 開発環境セットアップ・実行手順

## 概要
勤怠管理システムの開発環境セットアップと実行方法について説明します。

## システム構成
- **Frontend**: Next.js 15 + React 19 + TypeScript + Tailwind CSS
- **Backend**: Go + Echo Framework + GORM
- **Database**: PostgreSQL (Supabase Local Development)

## セットアップ手順

### 一括起動（おすすめ：1コマンド）
- Windows: ルートで `npm run dev` または `./dev.ps1`
- macOS/Linux: ルートで `npm run dev:unix` または `bash ./dev.sh`

実行内容:
- Supabase ローカル（Docker）を起動し、DBポート 54322 を待機
- Backend（Go）を起動（`backend/.env`の`DATABASE_URL`を使用）
- Frontend（Next.js）を起動

起動後のURL:
- Frontend: http://localhost:3000
- Backend: http://localhost:8080 (health: /health)
- Supabase Studio: http://127.0.0.1:54323

### 1. Supabase Local Development環境の起動
```bash
cd C:\Users\yudai\Programs\kintai
supabase start
```

### 2. バックエンドサーバーの起動

#### 2.1 依存関係の取得
```bash
cd backend
go mod tidy
```

#### 2.2 環境変数の確認
`backend/.env`ファイルが存在し、以下の内容になっていることを確認：
```
DATABASE_URL="postgresql://postgres:postgres@127.0.0.1:54322/postgres"
```

#### 2.3 サーバー起動
```bash
go run main.go
```

成功すると以下のようなログが表示されます：
```
Server starting on port 8080
```

### 3. フロントエンドサーバーの起動

#### 3.1 依存関係のインストール
```bash
cd frontend
npm install
```

#### 3.2 開発サーバー起動
```bash
npm run dev
```

### Supabase 認証（ログイン）
- 環境変数を設定（`frontend/.env.local` を作成）
  - 例は `frontend/.env.local.example` を参照
  - `NEXT_PUBLIC_SUPABASE_URL` と `NEXT_PUBLIC_SUPABASE_ANON_KEY` を設定（`supabase status` で取得可能）
- ログイン画面: `http://localhost:3000/login`
- 成功時: セッションCookieが設定され、保護ルートへ遷移します

ミドルウェアで未認証の保護ルート（例: `/employee`, `/admin`）は `/login` にリダイレクトされます。

## 動作確認

### 1. ブラウザでアクセス
http://localhost:3000 にアクセスします。

### 2. 接続テスト
ページ上の「接続テスト」ボタンをクリックして、フロントエンドとバックエンドの接続を確認します。

正常に接続されている場合：
```
接続成功: {"status":"OK","message":"Backend is running"}
```

### 3. 各サービスのURL確認
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Backend Health Check**: http://localhost:8080/health
- **Supabase Studio**: http://127.0.0.1:54323

## トラブルシューティング

### Supabase コンテナ名の衝突（"name is already in use"）
既存の Supabase/Docker コンテナと名前が衝突しています。

- 対処: リポジトリ直下（または `supabase` ディレクトリ）で `supabase stop` を実行。
- それでも解消しない場合、該当コンテナを削除：
  - 一覧: `docker ps -a --format "{{.ID}}\t{{.Names}}" | findstr kintai`
  - 削除: `docker rm -f <IDまたは名前>`
  - その後、`npm run dev` を再実行。

### ポートが使用されている場合
- Frontend (3000): `npm run dev -- -p 3001` で別ポート使用
- Backend (8080): 環境変数 `PORT=8081` で別ポート指定

### データベース接続エラー
1. Supabaseが起動していることを確認: `supabase status`
2. 環境変数の確認: `backend/.env`ファイルのDATABASE_URL
3. それでも `54322` に接続できない場合は、`supabase stop` → `supabase start` をやり直し

### 認証トラブル（JWT検証）
- `backend/.env` に `SUPABASE_JWT_SECRET` が設定されているか確認
  - 値は `supabase status` に表示される「JWT secret」と一致させてください
- 401エラー時はフロントのリクエストに `Authorization: Bearer <access_token>` が付いているか確認
  - `frontend` 側は `apiClient.get(url, { auth: true })` で自動付与されます

### CORS エラー
バックエンドのmain.goでCORS設定済み：
```go
e.Use(middleware.CORS())
```

## 開発に役立つコマンド

### Frontend
```bash
cd frontend
npm run dev    # 開発サーバー起動
npm run build  # プロダクションビルド
npm run lint   # コードチェック
```

### Backend
```bash
cd backend
go run main.go     # サーバー起動
go mod tidy       # 依存関係整理
```

### Supabase
```bash
supabase start    # ローカル環境起動
supabase stop     # ローカル環境停止
supabase status   # ステータス確認

## ログ出力

- ログディレクトリ: ルートの `log/`
  - フロントエンド: `log/frontend.log`
  - （将来）バックエンドも同ディレクトリへ出力予定
- いまはログイン失敗、employeeページのAPIエラー/操作失敗を自動で追記します。
```

## API エンドポイント

### Public Endpoints
- `GET /health` - ヘルスチェック

### Protected Endpoints (認証が必要)
- `POST /api/v1/attendance` - 勤怠登録
- `GET /api/v1/attendance/me` - 自分の勤怠取得
- `POST /api/v1/leaves` - 休暇申請
- `GET /api/v1/leaves` - 休暇一覧
- `GET /api/v1/schedules` - スケジュール一覧

### Admin Endpoints
- `GET /api/v1/admin/reports/monthly` - 月次レポート
