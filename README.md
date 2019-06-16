# About
http経由でlocateコマンドを発行し、結果をブラウザに表示.
ページ描画に時間がかかるため、最大表示件数1000にする、つもり。
その代わり検索条件を正規表現で選べるように工夫が必要.


# Feature

* -i ignorecase 大文字小文字無視(デフォルトでは無効、case sensitive)
* -l limit [NUM] 最大数(goでカウントするからいらないか)
* -r regex 正規表現

