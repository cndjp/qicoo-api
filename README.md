[![Travis CI](https://travis-ci.org/cndjp/qicoo-api.svg?branch=master)](https://travis-ci.org/cndjp/qicoo-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/cndjp/qicoo-api)](https://goreportcard.com/report/github.com/cndjp/qicoo-api)


# qicoo-api

test を動かすにはdockerサービスのインストールが必要です。

## ローカルでの開発

testを実行するには、上述のとおりdockerをinstallしている環境で `make test` を実行すると良い。
なお、実際にMySQLとRedisと連携させて開発したい場合は以下の手順で DockerコンテナとしてMySQL・Redisを起動し、環境変数を設定すると良い

```
docker run --name qicoo-api-test-mysql --rm -d -e MYSQL_ROOT_PASSWORD=my-secret-pw -p 3306:3306 mysql:5.6.27 --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci;
docker run --name qicoo-api-test-redis --rm -d -p 6379:6379 redis:4.0.10;

export DB_URL="127.0.0.1:3306"
export DB_USER="root"
export DB_PASSWORD="my-secret-pw"
export REDIS_URL="127.0.0.1:6379"
```


## ローカル開発とtravis CI環境との差分

基本的にローカルとtravis CI間で差分はない。
両方とも実際のMySQLとRedisを使用してテストデータを読み書きしている。
