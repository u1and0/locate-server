# Locate Server
ブラウザ経由でファイルパスを検索し、結果を最大1000件まで表示します。

## ***DEMO***
![demo](demo)

## Description
ウェブブラウザからの入力で指定ディレクトリ下にあるファイル内の文字列に対してlocateコマンドを使用した正規表現検索を行い、結果をhtmlにしてウェブブラウザに表示します。

## Requirement
* mlocate

Windows, Linux OK

MacOS 未テスト

## Usage

```
$ locate-server -h
Usage of locate-server:
  -P xargs -P
        Search in multi process by xargs -P (default 1)
  -d string
        Path of locate database file (ex: /path/something.db:/path/another.db) (default "/var/lib/mlocate/mlocate.db")
  -debug
        Debug mode
  -l int
        Maximum limit for results (default 1000)
  -r string
        DB insert prefix for directory path
  -s    OS path split windows backslash
  -t string
        DB trim prefix for directory path
  -v    show version
  -version
        show version
```

```
$ locate-server \
  -d $(paste -sd: <(find /var/lib/mlocate -name '*.db')) \
  -s \
  -t '\\gr.jp\share' \
  -l 2000 \
  -P 0
```

## Installation
```
$ go get github.com/u1and0/locate-server
```

or use docker

```
$ docker pull u1and0/locate-server
```

## Test

```
$ go test
```

## Features
### 検索機能
* 検索ワードを指定して検索を押すかEnterキーを押すと共有フォルダ内のファイルを高速に検索します。
* 対象文字列は2文字以上の文字列を指定してください。
* 英字 大文字/小文字は無視します。
* 全角/半角スペースで区切ると0文字以上の正規表現(\.\*)に変換して検索されます。(AND検索)
* `(aaa|bbb)`のグループ化表現が使えます。(OR検索)
  * 例: **golang\\\.(pdf|txt)** => **golang\.pdf**と**golang\.txt**を検索します。
* [a-zA-Z0-9]の正規表現が使えます。
  * 例: file[xy].txt で**filex.txt**と**filey.txt** を検索します。
  * 例: 201[6-9]S  => **2016S**, **2017S**, **2018S**, **2019S**を検索します。
* 0文字か1文字の正規表現`?`が使えます。
  * 例: **jpe?g** => **jpeg** と **jpg**を検索します。
* 単語の頭に半角ハイフン"-"をつけるとその単語を含まないファイルを検索します。(NOT検索)
	* 例: **gobook txt -doc**=>**gobook**と**txt**を含み**doc**を含まないファイルを検索します。
* AND検索は順序を守って検索をかけますが、NOT検索は順序は問わずに除外します。
	* 例: **gobook txt -doc** と**txt gobook -doc** は異なる検索結果ですが、 **gobook txt -doc** と**gobook -doc txt**は同じ検索結果になります。

### ファイル/フォルダ表示機能
* 検索結果はリンク付で最大1000件まで表示します。
* リンクをクリックするとファイルが開きます。
* **<<** マークをクリックするとそのファイルがあるフォルダが開きます。

### ブラウザ履歴機能との連携
ページタイトルに検索ワードが付属するので、ブラウザの**戻る**を長押ししたときに検索履歴が表示されます。

### ブラウザブックマーク機能との連携
ブックマークすることで、ワンクリックで検索を開始し、結果を表示できます。

### リンク機能
検索バーのURLは他人に送付することができます。
URLを送られた人はリンクをクリックするだけで検索バーに入力した文字列で検索を開始し、結果を閲覧することができます。

### 検索候補の表示
検索ツールボックスにはこれまで検索した検索語を検索候補として表示します。

---


## リンクをクリックしてもファイルが開かない現象について
### IEでリンクをクリックしてもファイルが開かない現象について
インターネット設定からhttp://(ホストマシンのIPアドレス)を信頼するサイトに追加します。

参考: [MS11-057　KB2559049　更新後　file://プロトコルでリンクしている共有ファイルが開けない](https://answers.microsoft.com/ja-jp/windows/forum/windows_xp-update/ms11-057-kb2559049-%E6%9B%B4%E6%96%B0%E5%BE%8C/9d18541c-faed-4cc5-bb8a-0830add7ccc1)


### GoogleChromeでリンクをクリックしてもファイルが開かない現象について
拡張機能を追加します。

[ローカルファイルリンク有効化](https://chrome.google.com/webstore/detail/enable-local-file-links/nikfmfgobenbhmocjaaboihbeocackld)


### Microsoft Edgeでリンクをクリックしてもファイルが開かない現象について
"GoogleChromeでリンクをクリックしてもファイルが開かない現象について" を参照してください。


### Firefoxでリンクをクリックしてもファイルが開かない現象について
アドオンを追加します。

[Local Filesystem Links](https://addons.mozilla.org/ja/firefox/addon/local-filesystem-links/?src=search)


# Release Note
## v2.1.0: 検索履歴
## v2.0.0: マルチスレッドlocate検索
* ~~マルチスレッド検索を有効化するオプション`locate-server -P [NUM]`を実装しました。~~
* マルチスレッド機能は現在調整中。`-P 1`以外の指定で検索されないときがあります。
* オプションの使い方は`man xargs` の `-P`オプションを参照してください。
* 検索履歴を読み込んで検索キーワードを使いまわしやすくしました。

## v1.0.4: ファイル・コマンドチェック、文字列ハイライトを1ワードに、検索文字列正規化縮小
* locateコマンドの実行可否チェック、ファイルアクセスチェックを追加しました。
* defer節によるfileクローズを明示しました。
* ハイライト文字列を全ワードから1ワードに変更しました。
* "\W"などの大文字の正規表現をログに記録できるように、バックスラッシュの後を小文字に正規化しないようにしました。

## v1.0.3: Title receive value
* ページタイトルに検索ワードを追加して、履歴を辿りやすくしました。

## v1.0.2: Structure optimization / Update DB & initialize cache daily
* `locate -S`の結果にerrorがあった場合の挙動を加えました。
* cacheの初期化を簡易にしました。
* errorに関するコメントを削除しました。
* データベースの更新およびキャッシュの初期化を1日1回にしました。

## v1.0.1: Prefixが多重に追加される問題を解決
* AddPrefix()とChangeSep()をPathMapのメソッドに移動しました。

## v1.0.0: Cache & NOT Search implemented
* 最大8時間以内に検索したワードはメモリ上にキャッシュされ、再度検索する際はキャッシュから検索結果を取り出します。
* NOT検索を実装しました。検索後の頭に「-」ハイフンをつけて検索するとその語を含む結果は除外されて表示されます。
* 構文の最適化を行い、パフォーマンスの改善を行いました。

## v0.3.0: Highlight search words & Show DB status
* 検索文字の背景を黄色で表示するようにしました。
* データベース情報を表示するリンクを追加しました。
* 構文の最適化を行い、パフォーマンスの改善を行いました。

## v0.2.0: Error check and OR search
* 検索文字列のエラーチェックで2文字以上のときだけ検索します。
* (aaa|bbb)のようにしてaaaまたはbbbを検索します。
* 検索説明文を追加しました。

## v0.1.2: Fix APP container
* [rm] locate command -v option cause of compress disk space @app/Docker
* [mod] wipe default run-part command @app/Dockerfile

## v0.1.0: Query search
* URLをquery表示することで前回の検索履歴を他人から見られなくしました。

## v0.0.0
* 検索ワードを指定して検索を押すかEnterキーを押すと共有フォルダ内のファイルを高速に検索します。
* 英数字の大文字小文字は無視します。
* 全角/半角スペースで区切るとは0文字以上の正規表現(\*)に変換して検索されます。(=and検索)
* 検索結果はリンク付で最大1000件まで表示し、リンクをクリックするとファイルが開きます。
* ?や[a-zA-Z0-9]の正規表現が使えます。
* 前回の検索履歴がアクセスした人すべてに見られてしまいます。(改善予定)
* フォルダジャンプ機能に対応しました。
> リンク右端の"<<"をクリックすると、そのファイルがあるフォルダがファイルエクスプローラーにて開きます。


# Deploy
Dockerコンテナによるシステム構成

## data volume用のコンテナdbを作る
```
docker create --name db -v /var/lib/mlocate -v /ShareUsers:/ShareUsers:ro busybox
```

このコマンドではdbコンテナの`/varlib/mlocate`を外部に晒して、
ホストのShareUsersをdbコンテナにマウントする。
ShareUsersが`locate`コマンドをかける対象のディレクトリ。


## updatedb用のコンテナappを作る

```
docker run --name app\
    --volumes-from db\
    -e UPDATEDB_PATH=/ShareUsers/<path to the db root>\
    -e OUTPUT=mlocatepersonal.db\
    u1and0/upadtedb
```

このコマンドではdbコンテナのボリュームを参照し、
`updatedb`をかけるパスを`UPDATEDB_PATH`で指定している。
dbでマウントしているのでこのコンテナで再度マウントする必要はない。
環境変数`OUTPUT`は出力するファイル名を指定する。
ディレクトリは`/var/lib/mlocate`に固定される。


## locateコマンドでファイル検索するコンテナwebを作る

`docker run --name web --volumes-from db u1and0/locate-server [OPTIONS]`

```
docker run --name web --rm -t\
   --volumes-from db\
   -e TZ='Asia/Tokyo'\
   -e LOCATE_PATH='/var/lib/mlocate/mlocatepersonal.db:/var/lib/mlocate/mlocatecommon.db'\
   -p 8081:8080\
   u1and0/locate-server -s -r '\\DFS' # オプションのみ
```

TZを指定しないとDBの更新日時がGMTになってしまう。
`LOCATE_PATH`はappコンテナで指定したパスの数だけ`:`で区切って記述する。
u1and0/locate-serverコンテナはENTRYPOINTで動くのでコンテナの指定後はオプションのみを記述する。

### コンテナ内で有効になっている検索パス
#### 環境変数の確認

``` shell-session
$ docker inspect --format='{{range .Config.Env}}{{println .}}{{end}}' web
TZ=Asia/Tokyo
LOCATE_PATH=/var/lib/mlocate/mlocatepersonal.db:/var/lib/mlocate/mlocatecommon.db:/var/lib/mlocate/mlocatecommunication.db
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
LANG=C.UTF-8
```

#### 検索パスの追加

1. updatedbするコンテナを作成
```shell-session
docker run --name personal --volumes-from db\
  -e TZ='Asia/Tokyo'\
  -e UPDATEDB_PATH=/ShareUsers/UserTokki/Personal\
  -e OUTPUT=mlocatepersonal.db\
  -d u1and0/updatedb
```


2. locate-server実行コンテナに対して、環境変数`LOCATE_PATH`の内容を変更したものを再度作成( run )する
2.1. `docker stop web`
2.2. `docker rename web web_old`  # 今まで使っていたコンテナを退避(バックアップ)
2.3. 新しい環境変数を設定したコンテナをrun `docker run ... -e LOCATE_PATH="..."``

# Authors
u1and0<e01.ando60@gmail.com>

# License
This project is licensed under the MIT License - see the LICENSE.md file for details
このプロジェクトは MIT ライセンスの元にライセンスされています。
