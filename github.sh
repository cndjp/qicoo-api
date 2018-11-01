#!/bin/bash -xe
HUB="2.6.0"

VERSION=$1

# 認証情報を設定する
mkdir -p "$HOME/.config"
set +x
echo "https://${GITHUB_TOKEN}:@github.com" > "$HOME/.config/git-credential"
echo "github.com:
- oauth_token: $GITHUB_TOKEN
  user: $GITHUB_USER" > "$HOME/.config/hub"
unset GH_TOKEN
set -x

# Gitを設定する
git config --global user.name  "${GITHUB_USER}"
git config --global user.email "${GITHUB_USER}@users.noreply.github.com"
git config --global core.autocrlf "input"
git config --global hub.protocol "https"
git config --global credential.helper "store --file=$HOME/.config/git-credential"

# hubをインストールする
curl -LO "https://github.com/github/hub/releases/download/v$HUB/hub-linux-amd64-$HUB.tgz"
tar -C "$HOME" -zxf "hub-linux-amd64-$HUB.tgz"
export PATH="$PATH:$HOME/hub-linux-amd64-$HUB/bin"

# リポジトリに変更をコミットする
hub clone "https://github.com/cndjp/qicoo-api-manifests.git" _
cd _
hub checkout -b "travis/$VERSION"
sed -i -e "s/cndjp\/qicoo-api:CURRENT/cndjp\/qicoo-api:$VERSION/g" ./overlays/staging/qicoo-api-patch.yaml
hub add .
hub diff
hub commit -m "コミットメッセージ"

# Pull Requestを送る
hub push --set-upstream origin "travis/$VERSION"
hub pull-request -m "Pull Requestメッセージ"