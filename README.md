[![Travis CI](https://travis-ci.org/cndjp/qicoo-api.svg?branch=master)](https://travis-ci.org/cndjp/qicoo-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/cndjp/qicoo-api)](https://goreportcard.com/report/github.com/cndjp/qicoo-api)



# qicoo-api

test を動かすにはdockerサービスのインストールが必要です。

## EKSでqicoo-api実行
東方仗助による自動上げ下げで、EKSのstaging,productionネームスペースに自動的にqicoo-apiが稼働します。
curlコマンドでの例を記載します。

### BASEURL設定

```
QICOOURL=https://api-s.qicoo.tokyo # staging
QICOOURL=https://api.qicoo.tokyo   # production
```

### Create Question
```
curl -vvv -H "Content-Type: application/json" -X POST "$QICOOURL/v1/jkd1812/questions" -d '
{
  "program_id": "1",
  "comment": "東方仗助の開発秘話を教えてください"
}
'
```

### List Question

```
curl -vvv -H "Content-Type: application/json" -X GET "$QICOOURL/v1/jkd1812/questions?start=1&end=10&sort=created_at&order=desc"
```

### Delete Question

```
curl -vvv -H "Content-Type: application/json" -X DELETE "$QICOOURL/v1/jkd1812/questions/qid"
```

### Like Question

```
curl -vvv -H "Content-Type: application/json" -X PUT "$QICOOURL/v1/jkd1812/questions/qid/like"
```

## ローカルでqicoo-api実行
ローカルで開発と同様に、以下コマンドでqicoo-apiをコンテナとして稼働させることが出来る

```
docker run --name qicoo-api-test-mysql --rm -d -e MYSQL_ROOT_PASSWORD=my-secret-pw --network host sugimount/qicoo-local-mysql:0.0.1 --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci;
docker run --name qicoo-api-test-redis --rm -d --network host redis:4.0.10
docker run --name qicoo-api-test --rm -d -e DB_URL="127.0.0.1:3306" -e DB_USER="root" -e DB_PASSWORD="my-secret-pw" -e REDIS_URL="127.0.0.1:6379" --network host cndjp/qicoo-api:v0.0.1-20181122-1707
```

※ image tag は、以下のDockerHubから最新のImageを指定して取得すると良いです

- QuestionList

```
curl -X GET 'http://localhost:8080/v1/jkd1812/questions?start=1&end=10&sort=created_at&order=desc'
```

- QuestionCreate
```
curl -X POST http://localhost:8080/v1/jkd1812/questions -d '
{
  "program_id": "1",
  "comment": "test1"
}
'
```

- QuestionDelete
```
curl -X DELETE 'http://localhost:8080/v1/jkd1812/questions/00000000-0000-0000-0000-000000000000'
```


## ローカルでgo test

testを実行するには、上述のとおりdockerをinstallしている環境で `make test` を実行すると良い。


## ローカルで開発

実際にMySQLとRedisと連携させて開発したい場合は以下の手順で DockerコンテナとしてMySQL・Redisを起動し、環境変数を設定すると良い

```
docker run --name qicoo-api-test-mysql --rm -d -e MYSQL_ROOT_PASSWORD=my-secret-pw -p 3306:3306 sugimount/qicoo-local-mysql:0.0.1 --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci;
docker run --name qicoo-api-test-redis --rm -d -p 6379:6379 redis:4.0.10;

export DB_URL="127.0.0.1:3306"
export DB_USER="root"
export DB_PASSWORD="my-secret-pw"
export REDIS_URL="127.0.0.1:6379"
```


## ローカル開発とtravis CI環境との差分

基本的にローカルとtravis CI間で差分はない。
両方とも実際のMySQLとRedisを使用してテストデータを読み書きしている。
