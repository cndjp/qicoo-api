# qicoo-api

test を動かすにはmysqlサービスのインストールが必要です。

## 開発

makefileから このレポジトリのトップディレクトリに `.env` で定義された環境定数をインストールするようになっている。  
本来なら `travis` 越しで環境変数は定義するが、ローカルでの開発にも対応しつつクレデンシャル情報を秘匿する為。

手元になければ以下のコマンドで作れる。  

```
$ make create-dotenv
$ cat .env
DB_USER=root
DB_PASSWORD=root
DB_URL=localhost:3306
REDIS_URL=localhost:6379
IS_TRAVISENV=
```

travisでの動作が見たい場合は `IS_TRAVISENV=` を `IS_TRAVISENV=true` と書き換えればよい。

## ローカル開発とtravis CI環境との差分

環境変数 `IS_TRAVISENV` で判定。
主に `go test` の動作だと思うが、

- MySQLを今localhostで動いているサービスで叩くのがtravisでモックで叩くのがローカル
- Redisを今localhostで動いているサービスで叩くのがtravisでモックで叩くのがローカル

くらいかな。
