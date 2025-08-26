了解です。添付資料の設計に沿って、\*\*ログインAPI（認証フロー）\*\*を要点だけコンパクトにまとめます。

---

# ログインAPI（認証フロー）まとめ

## 全体方針

* **トークンの発行はSupabaseが担当**（メール＋パスワードの認証とJWT発行）。
  \*\*GoバックエンドはJWTの“検証のみ”\*\*を行います。
* フロントエンド（Next.js）は **@supabase/ssr** とミドルウェアで **セッション管理／リフレッシュ** を実施します。

---

## 推奨フロー（実装粒度）

### 1) ログイン（サインイン）

* **実装場所**: Next.js の **Server Action** から Supabase の `signInWithPassword` を呼ぶ。
  → 資格情報はサーバ側に留まり、クッキーやセッションは @supabase/ssr が管理。
* **入力**: `email`, `password`
* **出力**: Supabase セッション（アクセストークン／リフレッシュトークン）
  ※ フロントでは通常、**Authorization: Bearer \<access\_token>** をGo API呼び出し時に付与。

### 2) アクセス保護

* **Next.js ミドルウェア**でログイン必須ルートを保護（未認証はログインページへリダイレクト）。
* **Go API** は各リクエストで **JWTを検証**（`golang-jwt/jwt` など）。**発行はしない**。

### 3) 認可（ロールに応じた制御）

* API設計上、`/attendance/me`（本人向け）と `/users/{userId}/attendance`（管理者向け）などで**権限が分離**。管理者系はミドルウェアで\*\*昇格権限（例: adminロール）\*\*を確認。

### 4) データベース側の最終防御（RLS）

* Supabase/Postgresの **Row Level Security** を有効化し、`auth.uid()` による**行レベルのアクセス制御**を適用。
  例：本人のみ `attendance_records` を読める／書けるポリシー。

---

## 具体的なAPI境界（自サービス側）

> 認証自体はSupabaseに委譲するため、**自前の「/auth/login」RESTエンドポイントは不要**です。どうしても必要なら**Server Actionの薄いラッパ**として用意します。

* **（任意）POST `/auth/login`**

  * Body: `{ "email": string, "password": string }`
  * 挙動: サーバ内で `signInWithPassword` を呼び、セッションをセットして200返却
  * エラー: 401（資格情報不正）、429（試行制限）など
  * 備考: 以降のAPI呼出は `Authorization: Bearer <access_token>` を付与
* **保護APIの例（Go/Echo）**

  * `GET /attendance/me`（本人の当日記録）
  * `POST /attendance`（打刻イベント作成）
  * `GET /users/{userId}/attendance`（管理者用） ほか
    いずれも **JWT検証必須**。

---

## ステータスコードとエラーポリシー（目安）

* `200 OK` / `201 Created` … 認証・処理成功
* `400 Bad Request` … 入力不備
* `401 Unauthorized` … 未ログイン／トークン不正・期限切れ
* `403 Forbidden` … 権限不足（管理者専用など）
  ※ エラーレスポンスは標準HTTPステータスに揃える。

---

## セキュリティ実務のポイント

* **API側**: CORS・セキュアヘッダー・（将来クッキーを使う場合の）CSRF対策をミドルウェアで。
* **DB側**: RLSを厳格運用（本人以外は見えない）。
* **フロント**: ルート保護とセッション自動更新（@supabase/ssr + Middleware）。

---

## まとめ（最小実装の指針）

1. \*\*ログインAPIはSupabase（`signInWithPassword`）\*\*に任せる。
2. **GoはJWT検証専任**（発行しない）。
3. **Next.jsミドルウェアでルート保護**＋**RLSで最終防御**。

---

