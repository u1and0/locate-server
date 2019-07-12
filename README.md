共有フォルダのファイル名を検索し、結果を最大1000件まで表示します。


# リンクをクリックしてもファイルが開かない現象について
## IEでリンクをクリックしてもファイルが開かない現象について
インターネット設定からhttp://(ホストマシンのIPアドレス)を信頼するサイトに追加します。

参考[MS11-057　KB2559049　更新後　file://プロトコルでリンクしている共有ファイルが開けない](https://answers.microsoft.com/ja-jp/windows/forum/windows_xp-update/ms11-057-kb2559049-%E6%9B%B4%E6%96%B0%E5%BE%8C/9d18541c-faed-4cc5-bb8a-0830add7ccc1)

## GoogleChromeでリンクをクリックしてもファイルが開かない現象について
拡張機能を追加します。
[ローカルファイルリンク有効化](https://chrome.google.com/webstore/detail/enable-local-file-links/nikfmfgobenbhmocjaaboihbeocackld)

## Firefoxでリンクをクリックしてもファイルが開かない現象について
試したけどうまくいかなかった
[Firefoxからファイルリソース(file://～)にアクセスする方法](https://plumpliver.com/post-5296/)


# Release Note
## v0.2.0: Error check and OR search
検索文字列のエラーチェックで2文字以上のときだけ検索します。
(aaa|bbb)のようにしてaaaまたはbbbを検索します。
検索説明文を追加しました。

* [add] error check (search queries are over 2char) @main.go
* [add] search log file
* [mod] use --regex opt at locate command @main.go

## v0.1.2: Fix APP container
* [rm] locate command -v option cause of compress disk space @app/Docker
* [mod] wipe default run-part command @app/Dockerfile

## v0.1.0: Query search
* URLをquery表示することで前回の検索履歴を他人から見られなくしました。

## v0.0.0
* 検索ワードを指定して検索を押すかEnterキーを押すと4～40秒程度で共有フォルダ内のファイルを検索します。
* 英数字の大文字小文字は無視します。
* 全角/半角スペースで区切るとは0文字以上の正規表現(\*)に変換して検索されます。(=and検索)
* 検索結果はリンク付で最大1000件まで表示し、リンクをクリックするとファイルが開きます。
* ?や[a-zA-Z0-9]の正規表現が使えます。
* 前回の検索履歴がアクセスした人すべてに見られてしまいます。(改善予定)
* フォルダジャンプ機能に対応しました。
> リンク右端の"<<"をクリックすると、そのファイルがあるフォルダがファイルエクスプローラーにて開きます。

Maintainer u1and0<e01.ando60@gmail.com>
