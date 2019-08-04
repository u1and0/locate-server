# Locate Server
共有フォルダのファイル名を検索し、結果を最大1000件まで表示します。

## 使い方
* 検索ワードを指定して検索を押すかEnterキーを押すと共有フォルダ内のファイルを高速に検索します。
* 対象文字列は2文字以上の文字列を指定してください。
* 英字 大文字/小文字は無視します。
* 全角/半角スペースで区切ると0文字以上の正規表現(\.\*)に変換して検索されます。(AND検索)
* `(aaa|bbb)`のグループ化表現が使えます。(OR検索)
  * 例: "電(気|機)工業" => **電気工業**と**電機工業**を検索します。
* [a-zA-Z0-9]の正規表現が使えます。
  * 例: file[xy].txt で**filex.txt**と**filey.txt** を検索します。
  * 例: 201[6-9]S  => **2016S**, **2017S**, **2018S**, **2019S**を検索します。
* 0文字か1文字の正規表現`?`が使えます。
  * 例: tx?t => **tt** と **txt**を検索します。
* `ccc -ddd`とするとcccを含み**dddを含まない**ファイルパスを検索します。
* 検索結果はリンク付で最大1000件まで表示し、リンクをクリックするとファイルが開きます。
* リンク右端の"<<"をクリックすると、そのファイルがあるフォルダがファイルエクスプローラーにて開きます。

---


## リンクをクリックしてもファイルが開かない現象について
### IEでリンクをクリックしてもファイルが開かない現象について
インターネット設定からhttp://(ホストマシンのIPアドレス)を信頼するサイトに追加します。

参考: [MS11-057　KB2559049　更新後　file://プロトコルでリンクしている共有ファイルが開けない](https://answers.microsoft.com/ja-jp/windows/forum/windows_xp-update/ms11-057-kb2559049-%E6%9B%B4%E6%96%B0%E5%BE%8C/9d18541c-faed-4cc5-bb8a-0830add7ccc1)


### GoogleChromeでリンクをクリックしてもファイルが開かない現象について
拡張機能を追加します。

[ローカルファイルリンク有効化](https://chrome.google.com/webstore/detail/enable-local-file-links/nikfmfgobenbhmocjaaboihbeocackld)


### Firefoxでリンクをクリックしてもファイルが開かない現象について
アドオンを追加します。

[Local Filesystem Links](https://addons.mozilla.org/ja/firefox/addon/local-filesystem-links/?src=search)


# Release Note
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


# Dockerコンテナによるシステム構成

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



Maintainer u1and0<e01.ando60@gmail.com>
