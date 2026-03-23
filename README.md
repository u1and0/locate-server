# Locate Server
ファイルパスを検索し結果をJSONで返すサーバーを立て、ブラウザに表示するファイルを配信します。

## ***DEMO***
![out](https://user-images.githubusercontent.com/16408916/143503512-6e172a98-f973-4c80-b1dc-99ea0ede0a71.gif)

## Description
ウェブブラウザからの入力で指定ディレクトリ下にあるファイル内の文字列に対してlocateコマンドを使用した正規表現検索を行い、結果をhtmlにしてウェブブラウザに表示します。

## Requirement
* plocate
* [gocate](https://github.com/u1and0/gocate)

Windows, Linux OK

MacOS 未テスト

## Usage

```
Usage of ./locate-server:
  -d string
      Path of locate database directory (default "/var/lib/plocate")
  -debug
    Debug mode
  -dir string
    Path of locate database directory (default "/var/lib/plocate")
  -p string
    Server port number. Default access to http://localhost:8080/ (default "8080")
  -port string
    Server port number. Default access to http://localhost:8080/ (default "8080")
  -r string
    DB insert prefix for directory path
  -root string
    DB insert prefix for directory path
  -s    OS path split windows backslash
  -t string
    DB trim prefix for directory path
  -trim string
    DB trim prefix for directory path
  -v    show version
  -version
    show version
  -windows-path-separate
    OS path separate windows backslash
```

```
$ locate-server \
  -d /home/mydir/plocate \
  -windows-path-separate \
  -trim '\\gr.jp\share' \
```

## Installation
```
$ go install github.com/u1and0/locate-server@latest
```

or use docker

```
$ docker pull u1and0/locate-server
```


## GLIBC not found
locate-server実行時にglibcが必要

```
./locate-server: /lib/x86_64-linux-gnu/libc.so.6: version `GLIBC_2.32' not found (required by ./locate-server)
```

cgoを無効にしてビルドすれば解決。

```
CGO_ENABLED=0 go build
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
  * 例: **golang (pdf|txt)** => **golang及びpdf**並びに**golang及びtxt**を検索します。
* [a-zA-Z0-9]の正規表現が使えます。
  * 例: file[xy] txt で**filex及びtxt並びに*と**filey及びtxt** を検索します。
  * 例: file[x-z] txt で**filex及びtxt**並びに**filey及びtxt**と**filez.txt** を検索します。
  * 例: 201[6-9]S  => **2016S**, **2017S**, **2018S**, **2019S**を検索します。
* 0文字か1文字の正規表現`?`が使えます。
  * 例: **jpe?g** => **jpeg** と **jpg**を検索します。
* 単語の頭に半角ハイフン"-"をつけるとその単語を含まないファイルを検索します。(NOT検索)
	* 例: **gobook txt -doc**=>**gobook**と**txt**を含み**doc**を含まないファイルを検索します。
* AND検索は順序を守って検索をかけますが、NOT検索は順序は問わずに除外します。
	* 例: **gobook txt -doc** と**txt gobook -doc** は異なる検索結果ですが、 **gobook txt -doc** と**gobook -doc txt**は同じ検索結果になります。
* ファイル拡張子を指定するときは、文字列の最後を表す**$**記号を行末につけます。
	* 例: **gobook pdf$ **=>**gobook**を含み、**pdf**が行末につくファイルを検索します。

### ファイル/フォルダ表示機能
* 検索結果はリンク付で最大1000件まで表示します。(v2.X.Xまで)
* リンクをクリックするとファイルが開きます。
* **<<** マークあるいはフォルダアイコンをクリックするとそのファイルがあるフォルダが開きます。

### ブラウザ履歴機能との連携
ページタイトルに検索ワードが付属するので、ブラウザの**戻る**を長押ししたときに検索履歴が表示されます。

### ブラウザブックマーク機能との連携
ブックマークすることで、ワンクリックで検索を開始し、結果を表示できます。

### リンク機能
検索バーのURLは他人に送付することができます。
URLを送られた人はリンクをクリックするだけで検索バーに入力した文字列で検索を開始し、結果を閲覧することができます。

### 検索候補の表示
検索ツールボックスにはこれまで検索した検索語を検索候補として表示します。

### API

| 説明 | メソッド | URI | パラメータ | パラメータ例 |
|----|------|-----|-------|-------|
| ファイルパスの検索結果をJSONで返す | GET | /json |  q=, logging= | http://localhost:8080/json?q=keyword <br>http://localhost:8080/json?q=keyword&logging=false <br>loggingの値はboolian, falseのとき、検索キーワードを履歴に残さない |
| ファイルパスの検索結果をHTMLで表示する | GET | /search |  q=, logging= | http://localhost:8080/json?q=keyword <br>http://localhost:8080/json?q=keyword&logging=false <br>loggingの値はboolian, falseのとき、検索キーワードを履歴に残さない |
| 検索履歴を見る | GET | /history |  gt=, lt= |  http://localhost:8080/history?gt=10&lt=1000 <br>10以上、1000未満のスコアの履歴のみJSONで返す |
| DBの状態確認 | GET | /status |  なし | http://localhost:8080/status <br>ステータス表示 |

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


# Deploy
Dockerコンテナによるシステム構成

## data volume用のコンテナdbを作る
```
docker create --name db -v /var/lib/plocate -v /ShareUsers:/ShareUsers:ro busybox
```

このコマンドではdbコンテナの`/varlib/plocate`を外部に晒して、
ホストのShareUsersをdbコンテナにマウントする。
ShareUsersが`locate`コマンドをかける対象のディレクトリ。


## updatedb用のコンテナappを作る

```
docker run --name app\
    --volumes-from db\
    -e UPDATEDB_PATH=/ShareUsers/<path to the db root>\
    -e OUTPUT=plocatepersonal.db\
    u1and0/upadtedb
```

このコマンドではdbコンテナのボリュームを参照し、
`updatedb`をかけるパスを`UPDATEDB_PATH`で指定している。
dbでマウントしているのでこのコンテナで再度マウントする必要はない。
環境変数`OUTPUT`は出力するファイル名を指定する。
ディレクトリは`/var/lib/plocate`に固定される。


## locateコマンドでファイル検索するコンテナwebを作る

`docker run --name web --volumes-from db u1and0/locate-server [OPTIONS]`

```
docker run --name web --rm -t\
   --volumes-from db\
   -e TZ='Asia/Tokyo'\
   -e LOCATE_PATH='/var/lib/plocate/plocatepersonal.db:/var/lib/plocate/plocatecommon.db'\
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
LOCATE_PATH=/var/lib/plocate/plocatepersonal.db:/var/lib/plocate/plocatecommon.db:/var/lib/plocate/plocatecommunication.db
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
LANG=C.UTF-8
```

#### 検索パスの追加

1. updatedbするコンテナを作成
```shell-session
docker run --name personal --volumes-from db\
  -e TZ='Asia/Tokyo'\
  -e UPDATEDB_PATH=/ShareUsers/UserTokki/Personal\
  -e OUTPUT=plocatepersonal.db\
  -d u1and0/updatedb
```


2. locate-server実行コンテナに対して、環境変数`LOCATE_PATH`の内容を変更したものを再度作成( run )する
2.1. `docker stop web`
2.2. `docker rename web web_old`  # 今まで使っていたコンテナを退避(バックアップ)
2.3. 新しい環境変数を設定したコンテナをrun `docker run ... -e LOCATE_PATH="..."``


# Bugs
既知のバグ報告。

## 検索ワードハイライトが検索順序を守らない。
内部的にString.ReplaceAll()を使用しているため。
