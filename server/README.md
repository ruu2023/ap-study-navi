# AP科目B 学習ナビ（サーバー版）

https://lab.ruu-dev.com/ap/

GitHub Pages 版（`docs/`）の後継です。進捗と演習記録をサーバーに保存するので、Mac とスマホで共有できます。

- 構成: Go（標準ライブラリ）＋ htmx ＋ SQLite（`modernc.org/sqlite`、CGO なし）の単一バイナリ
- 配置: lab-ruu（Lightsail 1GB）のコンテナ `ap-study`。lab-ruu の Caddy が `/ap/*` を `ap-study:8080` に転送する（`~/projects/lab-ruu/Caddyfile`）
- データ: Docker ボリューム `ap-study_data` の `/data/ap-study.db`。新しい DB の初回起動時に 11/19〜11/21 の進捗と演習 9 件を取り込む（`store.go` の `seedOnce`）
- ログイン: Google（OpenID Connect の認可コードフロー＋PKCE）。`ALLOWED_EMAILS` のアドレスだけ入れる。Cookie は HMAC 署名で 90 日有効。Google Cloud プロジェクトは `ap-study-navi`（同意画面はテスト中）
- 秘密情報は `/opt/ap-study/.env`（600）にだけ置く: `AP_SECRET`, `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`, `ALLOWED_EMAILS`。`GOOGLE_CLIENT_ID` が無いときだけ、ローカル用のパスフレーズ（`AP_PASSWORD`）でログインする

## 画面

| タブ | 内容 |
|---|---|
| 今日 | 日ごとの課題、完了チェック、課題の追加 |
| 計画 | 11 日間の進み具合、模試の得点 |
| 記録 | 演習の記録と分野別の成績 |
| 分野 | メイン・保険・解かないの選択（分野別の成績つき） |
| 決まり文句 | 解答例から集めた言い回し |
| 本番 | 時間配分と CBT の注意点 |

## 開発

```bash
cd server && AP_PASSWORD=localtest123 AP_SECRET=$(openssl rand -hex 32) COOKIE_SECURE=0 DB_PATH=/tmp/ap.db ADDR=127.0.0.1:8099 go run .
```

http://127.0.0.1:8099/ap/

## デプロイ

```bash
server/deploy/deploy.sh
```

演習を記録する・分野を変える（サーバー上の CLI）:

```bash
ssh ubuntu@13.159.19.55 "docker exec ap-study /ap-study session-add -date 2026-11-21 -exam r5h -q 2 -min 15 -correct 4 -total 9 -memo 'メモ'"
```

```bash
ssh ubuntu@13.159.19.55 'docker exec ap-study /ap-study fields-set -main st,pm,nw,au,db -insurance es,sm'
```

ログを見る:

```bash
ssh ubuntu@13.159.19.55 'docker logs -f --tail 100 ap-study'
```

DB をバックアップする（WAL ファイルも含めてフォルダごと Mac に保存）:

```bash
ssh ubuntu@13.159.19.55 'rm -rf /tmp/ap-study-backup && docker cp ap-study:/data /tmp/ap-study-backup' && scp -r ubuntu@13.159.19.55:/tmp/ap-study-backup ./ap-study-backup
```
