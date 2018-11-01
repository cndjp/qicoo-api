#!/bin/bash -xe
HUB="2.6.0"

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
export PATH="$PATH:$HOME/hub-linux-amd64-$HUB"

# リポジトリに変更をコミットする
hub clone "https://github.com/cndjp/qicoo-api-manifests.git" _