ファイルパスを検索し結果をJSONで返すREST APIサーバーを立てます。
ひとまず動きのイメージを掴むデモです。

![out](https://user-images.githubusercontent.com/16408916/143503512-6e172a98-f973-4c80-b1dc-99ea0ede0a71.gif)

検索窓に検索キーワードを入力し、検索ボタンを押すとlocateコマンドを走らせて、結果をブラウザに表示します。

本記事は[locate-server](https://github.com/u1and0/locate-server) v3.1.0の時点のREADMEを補完するドキュメントを記事としました。


## 必要な知識
この記事を読むために必要な知識

* Go
	* Gin(フレームワーク)
* ShellScript
* HTML5
* JavaScript
  * jQuery
  * Ajax
* Docker(プラットフォーム)
	* Docker Compose

どれも入門レベルの知識で済むと思います。

# 実行
## 前提条件
サーバーを立ち上げるホストマシンに下記パッケージが必要です。

* mlocate
* [gocate](https://github.com/u1and0/gocate)

`mlocate`は`locate`、`updatedb`を実行するパッケージです。
普通のLinuxディストリビューションには標準で入っていると思います。
`gocate`は`locate`と`updatedb`コマンドを並列実行できるコマンドです。この`locate-server`のために自作しました。これのおかげで検索実行時間が20秒から4秒台に縮まりました。[^2]

[^2]: たぶんファイルシステムがntfsではなくext4だったら`locate`でも十分速いと思います。 テスト環境(Linux)でほぼ同数のファイル検索に10秒以上かかったことがありません。テスト環境のドライブがHDDではなくSSDであることも起因しているかもしれません。 テスト環境のCPUスレッド数が6に対して、本番環境(Windows)のスレッド数が24あるので、並列実行できるプロセス数は本番環境のほうが多いはずですが、パフォーマンスは本番環境のほうが悪いです。

自前記事ですが、上記コマンドについて詳細をつめた際、制作したドキュメントです。

`locate`について参考: [あなたの知らないlocateの世界](https://qiita.com/u1and0/items/98620f9af3dadad4ced1)
`gocate`について参考: [並列実行できるlocateコマンドの実装](https://qiita.com/u1and0/items/964be5817da800b82603)


## サーバーサイドの実行

* Linuxファイルシステムの検索
* /var/lib/mlocate に`updatedb`または `gocate -init` で作成したデータベースファイルが既にある

とした場合

```shell-session
$ locate-server
```

だけで実行できます。


主なコマンドラインオプションをつけて説明すると、

```shell-session
$ locate-server \
  -dir /home/mydir/mlocate \  # XXX.dbが保存されているディレクトリの指定(default: /var/lib/mlocate)
  -trim '/mnt'             \  # ファイル名のprefixを削除します
  -root '\\ns\FileShare'   \  # ファイル名のprefixに追加します
  -windows-path-separate   \  # ファイルセパレータ'\'を使用します(default: false)
```

ショートオプションで縮めて書くと下記のようになります。

```shell-session
$ locate-server -d /home/mydir/mlocate -t '/mnt' -r '\\ns\FileShare' -s
```


## クライアントサイドの実行
サーバーサイドでサーバーを立ち上げたら、クライアントはブラウザのURL欄に `localhost:8080` と入力するとトップページが表示されます。


<!-- ## 制作のきっかけ -->
<!-- 社内の共有ファイルサーバーにあるファイル数が600万件超、共有サーバーを使用している作業者が40名超です。 -->
<!-- 自分のファイルを探すならまだしも、過去に誰かが作ったファイルを探しにいくときに大変時間がかかります。 -->
<!-- フォルダ構成のルールやガイドラインがおおまかにしかないため[^1]、個々人が思い思いのフォルダ構成にして、作った本人ですら個人用に保存したか、共有用フォルダに保存したか、どのプロジェクトに保存したか、そもそも他人には見えない権限の場所にファイルを保存したか忘れてしまうことがあります。 -->
<!-- **その他**とか言うフォルダ名やめてほしいですよね！(とりあえずの名称つけたいキモチはわかる) -->
<!--  -->
<!-- [^1]: それが楽ではあるのですが。共有フォルダとはいえ、個人用フォルダ、共有用フォルダくらいのくくりはあります。 -->
<!--  -->
<!-- ファイル探しをする状況が、20,30年前のプロジェクトであることも何度もあり、ファイル作成者も退職、異動していることもあって、そうなってしまうと探せないファイルは持っていないものと同じことです。 無駄にファイルサーバーの容量を食いつぶしているだけの存在です。 -->
<!--  -->
<!-- 私はLinuxが使えるので、ファイルサーバーをシステム上にマウントしてlocateコマンド打っていれば検索できますが、そのパスはLinuxファイルセパレータで書かれているのでコピペでそのファイルにアクセスできるわけもなく、多少便利になった程度でした。 -->
<!--  -->

# 内部動作(サーバーサイド)
## サーバーサイドの動作概要
* ウェブブラウザからの入力で指定ディレクトリ下にあるファイル内の文字列に対してlocateコマンドを使用した正規表現検索を行い、結果をJSONにしてクライアントに送ります。
* JSONを受け取ったクライアントは、static下に配置されたJavaScriptファイルでHTMLに変換して描画します。

## ディレクトリ構造
パッケージのディレクトリ構造は次のようになります。

```
locate-server
├── main.go
├── cmd
│   ├── api
│   │   ├── api.go
│   │   ├── api_test.go
│   │   ├── query.go
│   │   └── query_test.go
│   ├── cache
│   │   └── cache.go
│   └── locater
│       ├── command.go
│       ├── command_test.go
│       ├── frecency.go
│       ├── frecency_test.go
│       ├── locater.go
│       └── locater_test.go
├── static
│   ├── datalist.js
│   ├── distributeResult.js
│   ├── icons8-検索-50.png
│   ├── search-location-solid.png
│   ├── style.css
│   └── tooltips.js
├── templates
│   └── index.tmpl
├── test
├── Dockerfile
└── docker-compose.yml
```

## ページの表示

```go:main.go
import (
	/* snip...*/
	"github.com/gin-gonic/gin"  //...(1)
)

func main() {

	/* snip...*/

	// Open server
	route := gin.Default()
	route.Static("/static", "./static")  //...(2)
	route.LoadHTMLGlob("templates/*")  //...(3)

	// Top page
	route.GET("/", topPage)  //...(4)

	// Result view
	route.GET("/search", searchPage)  //...(4)

	// API
	route.GET("/history", fetchHistory)  //...(5)
	route.GET("/json", fetchJSON)  //...(6)
	route.GET("/status", fetchStatus)  //...(7)

	// Listen and serve on 0.0.0.0:8080
	route.Run(":" + strconv.Itoa(port)) // => :8080 ...(8)
}
```

1. フレームワークは[gin](https://github.com/gin-gonic/gin)を使っています。
2. css, JavaScript, favicon用pngファイルはstaticに置いてあります。
3. トップページの表示はtemplates/index.htmlに`gin.H{}`構造体の内容を埋め込んで表示します。
4. トップページと検索ページは結果の表示がされているかどうかだけで、同じテンプレートを使用します。
5. APIを今のところ3つ用意しています。いずれもGETメソッドです。/historyは検索履歴をFrecencyスコア順にしてJSONで取得します。
6. /jsonはlocate検索を走らせて検索結果をJSONで取得します。
7. /statusはDBの`locate -S`の出力をJSONで取得します。
8. デフォルトでは8080ポートでサーバーを公開します。


```html:template/index.tmpl
<html>
    <head>
		<!-- snip -->
    </head>
    <body>

		<!-- 1 -->

      <!-- GET method URI-->
      <form name="form1" method="get" action="/search">
        <a href=/ class="fas fa-home" title="Locate Server Home"></a>
        <!-- 検索窓 -->
        <input type="text" name="q" value="{{ .query }}" size="50" list="search-history" placeholder="検索キーワードを入力">
        <!-- 検索履歴 Frecency リスト -->
        <datalist id="search-history"></datalist>
        <!-- 検索ボタン -->
        <input type="submit" id="submit" value="&#xf002;" class="fas">
        <input type="button" onclick="toggleMenu('hidden-explain')" value="&#xf05a;" class=fas title="Help"> <!--// Help折りたたみ展開ボタン -->
      </form>

		<!-- snip -->

		<!-- 2 -->

      <!-- Database status -->
      <div id="search-status">
        <b><a href=/status>DB</a> last update: {{ .lastUpdateTime }}</b><br>
      </div>

		<!-- 3 -->

      <!-- Search result -->
      <div class="loader-wrap">
        <div class="loader">Loading...</div>
      </div>
      <table id="result"></table>
      <div id="error-view"><div>


		<!-- 4 -->

    <script src="https://ajax.googleapis.com/ajax/libs/jquery/3.5.1/jquery.min.js"></script>
    <script type="text/javascript" src="/static/distributeResult.js"></script>
    <script type="text/javascript" src="/static/tooltips.js"></script>
    <script type="text/javascript" src="/static/datalist.js"></script>
  </body>
</html>
```

1. 検索フォームです。/search?q=キーワードのページに飛びます。
2. DBのステータス表示です。
3. 検索結果とエラーを表示します。ページを読み込むまで、ロードスピナーが回ります。
4. 使用するJavaScriptファイルです。


## 検索結果を返すページ
ユーザーが主にアクセスするページです。
topPage()はsearchPage()の簡略版なので、省略します。

```go:main.go/searchPage()
func searchPage(c *gin.Context) {
	// 検索文字数チェックOK
	/* LocateStats()の結果が前と異なっていたら
	locateS更新
	cacheを初期化 */
	if l, err := cmd.LocateStats(locater.Dbpath); string(l) != string(locateS) {  //...(1)
		// DB更新されていたら
		if err != nil {
			log.Error(err)
		}
		locateS = l // 保持するDB情報の更新
		// Initialize cache
		// nil map assignment errorを発生させないために必要
		caches = cache.New() // Reset cache ...(2)
		// Count number of search target files
		var n int64
		n, err = cmd.LocateStatsSum(locateS)
		if err != nil {
			log.Error(err)
		}
		locater.Stats.Items = cmd.Ambiguous(n)  //...(3)
		// Update LastUpdateTime for database
		locater.Stats.LastUpdateTime = cmd.DBLastUpdateTime(locater.Dbpath)  //...(3)
	}
	// Response
	q := c.Query("q")
	c.HTML(http.StatusOK, "index.tmpl", gin.H{  //...(4)
		"title":          q,
		"lastUpdateTime": locater.Stats.LastUpdateTime,
		"query":          q,
	})
}
```

1. `locate -S`の結果が検索前と異なっていないかのチェックです。異なる場合、DBが更新されたので、2,3の処理を行います。
2. キャッシュをリセットします。
3. DBのファイル数、DBの更新時間を再取得します。
4. タイトルと更新時間と、ページ遷移前に検索窓に入力された文字列を検索窓に再入力してページを表示します。


## 検索結果を返すAPIの実装
```go:main.go/json
func fetchJSON(c *gin.Context) {
	// locater.Query initialize
	// Shallow copy locater to local
	// for blocking to rewrite
	// locater{} struct while searching
	local := locater  //...(1)

	// Parse query
	query, err := api.New(c)  //...(2)
	local.Query = api.Query{
		Q:       query.Q,
		Logging: query.Logging,
		Limit:   query.Limit,
	}

	/* snip...*/

	local.SearchWords, local.ExcludeWords, err = api.QueryParser(query.Q)  //..

	/* snip...*/

	// Execute locate command
	start := time.Now()  //...(3)
	result, ok, err := caches.Traverse(&local) // err <- OS command error ...(4)
	/* snip...*/
	end := (time.Since(start)).Nanoseconds()  //...(3)
	local.Stats.SearchTime = float64(end) / float64(time.Millisecond)

	// Response & Logging
	if err != nil {
		log.Errorf("%s [ %-50s ]", err, query.Q)
		c.JSON(500, local)
		// 500 Internal Server Error
		// 何らかのサーバ内で起きたエラー
		return
	}
	local.Paths = result
	getpushLog := "PUSH result to cache"
	if ok {
		getpushLog = "GET result from cache"
	}
	if !query.Logging {
		getpushLog = "NO LOGGING result"
	}
	l := []interface{}{len(local.Paths), local.Stats.SearchTime, getpushLog, query.Q}  //...(6)
	log.Noticef("%8dfiles %3.3fmsec %s [ %-50s ]", l...)  //...(6)
		if len(local.Paths) == 0 {
			local.Error = "no content"
			c.JSON(204, local)  //...(7)
			// 204 No Content
			// リクエストに対して送信するコンテンツは無いが
			// ヘッダは有用である
			return
		}
		c.JSON(http.StatusOK, local)  //...(7)
		// 200 OK
		// リクエストが正常に処理できた
	}
}
```

1. global変数locaterをshallow copyしてlocalに代入します。関数の引数として与えられればわざわざこんな行は必要ないのですが、`route.GET()`関数に渡せるのは関数のみですので、どうしたらいいやら。クロージャ使えばいいのか？
2. Query構造体を、gin.Contextを基に、ページのコンテキストから取得します。(後述)
3. コマンド実行時間の計測を行い、ミリ秒で返します。
4. キャッシュの中を検索し、検索結果があればresultに結果を入れて、okにtrueが入ります。キャッシュ内に検索結果がなければ、`locate`(`gocate`)コマンドを実行して、resultに検索結果を格納し、okにfalseが入ります。この一行が"検索サーバー"としてのメインの仕事を担います。(後述)
5. logに表示する奴らをまとめています。型がバラバラなので、`[]interface`を使います。
6. log表示です。URLに`&logging=false`を指定すると後述する検索履歴のスコアに加算されないようにしてログへ記録します。ログへ記録することが検索履歴のスコアへ影響を及ぼすため、テスト用または今後実装する機能のためにログへの記録制御を行います。
7. 検索結果、クエリ制御、エラー等々ひっくるめてLocate構造体に入れて、JSONオブジェクトを返します。


```go:cmd/api/query.go
type (
	// Query : URL で指定されてくるAPIオプション
	Query struct {  //...(2)
		Q       string `form:"q"`       // 検索キーワード,除外キーワードクエリ
		Logging bool   `form:"logging"` // LOGFILEに検索記録を残すか default ture
		// 検索結果上限数
		// LimitをUintにしなかったのは、head の-nオプションが負の整数も受け付けるため。
		// 負の整数を受け付けた場合は、-n=-1と同じく、制限なしに検索結果を出力する
		Limit int `form:"limit"`
	}
)

// New : Query constructor
// Default value Logging: ture <= always log search query
//									if ommited URL request &logging
// Default value Limit: -1 <= dump all result
//									if ommited URL request &limit
func New(c *gin.Context) (*Query, error) {
	query := Query{Logging: true, Limit: -1}  //...(2)
	err := c.ShouldBind(&query)  //...(1)
	return &query, err
}
```

1. Query構造体を、gin.Contextを基に、ページのコンテキストから取得します。
2. 構造体を指定してから、`gin.Context.ShouldBind()`を使うと、boolianやint型を類推して構造体に当てはめてくれるので、`strconv.Atoi()`とかしなくて済むので大変楽です。


```go:cmd/cache/cache.go
type (
	// Map is normalized queries key and PathMap value pair
	Map map[Key]*cmd.Paths
	// Key : cache Map key
	Key struct {
		Word  string // Normalized query
		Limit int    // Number of results
	}
)

// Traverse : 検索結果をcacheの中から探し、あれば検索結果と検索数を返し、
// なければLocater.Cmd()を走らせて検索結果と検索数を得る
func (cache *Map) Traverse(l *cmd.Locater) (paths cmd.Paths, ok bool, err error) {
	w := cmd.Normalize(l.SearchWords, l.ExcludeWords)  //...(1)
	k := Key{w, l.Query.Limit}  //...(1)
	if v, ok := (*cache)[k]; !ok {  //...(2)
		// normalizedがcacheになければresultsをcacheに登録
		paths, err = l.Locate()  //...(3)
		(*cache)[k] = &paths  //...(3)
	} else {
		paths = *v  //...(4)
	}
	return
}
```

1. 検索語(SearchWords)と除外語(ExcludeWords)を正規化(Normalize)して、MapのKeyとします。
2. cacheからキーワードkを探します。
3. 結果がなければlocate(gocate)コマンドで検索し、pathsをcacheに登録して返します。(後述)
4. 結果があればその値vをpathsとして返します。


```go:cmd/locater/locater.go
// Locate excute locate (or gocate) command
// split from Locater.Cmd()
func (l *Locater) Locate() (Paths, error) {
	out, err := pipeline.Output(l.CmdGen()...)  //...(1)
	outslice := strings.Split(string(out), "\n")  //...(1)
	outslice = outslice[:len(outslice)-1] // Pop last element cause \\n
	return outslice, err
}

// CmdGen : shell実行用パイプラインコマンドを発行する
func (l *Locater) CmdGen() (pipeline [][]string) {
	locate := []string{  //...(2)
		"gocate",               // locate command path
		"--database", l.Dbpath, //Add database option
		"--",            // Inject locate option
		"--ignore-case", // Ignore case distinctions when matching patterns.
		"--quiet",       // Report no error messages about reading databases
		"--existing",    // Print only entries that refer to files existing at the time locate is run.
		"--nofollow",    // When  checking  whether files exist do not follow trailing symbolic links.
	}
	// -> gocate --database -- --ignore-case --quiet --regex hoge.*my.*name

	// Include PATTERNs
	// -> locate --ignore-case --quiet --regex hoge.*my.*name
	locate = append(locate, "--regex", strings.Join(l.SearchWords, ".*"))  //...(3)

	pipeline = append(pipeline, locate)

	// Exclude PATTERNs
	for _, ex := range l.ExcludeWords {
		// COMMAND | grep -ivE EXCLUDE1 | grep -ivE EXCLUDE2
		pipeline = append(pipeline, []string{"grep", "-ivE", ex})  //...(4)
	}

	// Limit option
	if l.Query.Limit > 0 {
		pipeline = append(pipeline, []string{"head", "-n", strconv.Itoa(l.Query.Limit)})  //...(5)
	}

	if l.Args.Debug {
		log.Debugf("Execute command %v", pipeline)
	}
	return  // => locate ... | grep -ivE ... | head -n ... ...(6)
}
```

1. `l.CmdGen()`でコマンド文字列の生成を行い、locate(gocate)コマンドを実行します。`pipeline.Output()`の結果は[]byteで返ってくるので、[]stringに変えて返却します。
2. `locate`(`gocate`)コマンドの文字列を生成します。構造体の定義は常時つくオプションです。
3. `l.SearchWords`はSliceなので、`locate`に渡せるように".\*"を挟みます。
4. 除外するキーワードを`grep -v`で排除します。`-i`でignore case, `-E`で正規表現。
5. 出力行数を`head -n`で制御します。テスト環境(Linux上のDocker)ではうまくいって、本番環境(Windows上のDocker)でうまく動いていないような...。`gocate`を改造して--limitオプション付けるか思案中。
6. 最終的にコマンドラインに入力する文字列 `locate "検索語" | grep -ivE "除外語" | head -n "結果上限数" を返します。


## 検索履歴のスコアを返すAPIの実装
/historyで返す検索履歴を解析し、frecencyスコア算出、その順序でJSONオブジェクトにして返します。
スコアはfrecency(frequently 頻繁に + recency 最近のからなる造語)を算出します。、

```go:main.go/history
route.GET("/history", func(c *gin.Context) {
	searchHistory, err := cmd.Datalist(LOGFILE)  //...(1)
	/* snip...*/
	c.JSON(http.StatusOK, searchHistory)
})
```


```go:frecency.go
// Scoring : 日時から頻出度を算出する
func Scoring(t time.Time) int {  //...(2)
	since := time.Since(t).Hours()
	switch {
	case since < 6:
		return 32
	case since < 24:
		return 16
	case since < 24*7:
		return 8
	case since < 24*14:
		return 4
	case since < 24*28:
		return 2
	default:
		return 1
	}
}

//ScoreSum : 履歴マップの検索日時リストからスコア合計を算出する
func ScoreSum(tl []time.Time) (score int) {
	for _, t := range tl {
		score += Scoring(t)
	}
	return
}
```

1. LOGFILEを解析して、frecencyスコア順で返します。詳細は`cmd/locater/frecency.go`を参照してください。
2. スコアの参照は現在時刻からの経過時間(Hour単位)でスコアを出し、検索回数分足し算します。


```go:main.go/status
func fetchStatus(c *gin.Context) {
	l, err := cmd.LocateStats(locater.Args.Dbpath) // err <- OS command error ...(1)
	ss := strings.Split(string(l), "\n")  //...(2)
	/* snip...*/
	c.JSON(http.StatusOK, gin.H{  //...(3)
		"locate-S": ss,
		"error":    err,
	})
}
```

```go:command.go
// LocateStats : Result of `locate -S`
func LocateStats(s string) ([]byte, error) {
	dbs, err := filepath.Glob(s + "/*.db")  //...(4)
	if err != nil {
		return []byte{}, err
	}
	d := strings.Join(dbs, ":")  //...(5)
	b, err := exec.Command("locate", "-Sd", d).Output()  //...(6)
	// => locate -Sd /var/lib/mlocate/db1.db:/var/lib/mlocate/db2.db:...
	if err != nil {
		return b, err
	}
	return b, err
}
```

1. `LocateStats()`を実行して`locate -S`の結果を得ます。
2. []byte型なので、stringにし、改行で区切ってsliceとします。
3. 2の結果をJSONにして送ります。
4. dbファイルを列挙します。
5. ":"でつなげて `locate -Sd /var/lib/mlocate/db1.db:/var/lib/mlocate/db2.db:...` のように実行します。


# API

| 説明 | メソッド | URI | パラメータ |
|----|------|-----|-------|
| ファイルパスを検索する | GET | /json |  q=, logging=, limit= |
| 検索履歴を見る | GET | /history |  gt=, lt= |
| DBの状態確認 | GET | /status |   |


サーバーを立ち上げた状態で

```shell-session
$ curl -fsSL localhost:8080/json?q=usr+bin+sh&limit=10&logging=false
```

とすると、

* 検索上限数10
* Frecency スコアに影響しないログ出力
* `gocate -- --regex 'usr.*bin.*sh' `  (細かいオプションは省略)

上記の条件で検索した結果をJSONにして標準出力に表示します。


# 内部動作(クライアントサイド)
~~初めての~~フロントエンド開発していきます。

## クライアントサイドの動作概要


## 検索
ユーザーはトップページから検索ボタンをクリックすると/searchページに飛びます。
> ここまではサーバーサイドmain.goに書かれていること。
/searchページに飛ぶとdistributeResult.jsの`main()`が走ります。
JavaScriptでJSONをパースします。

[JavaScriptPrimer 第2部/Ajax通信](https://jsprimer.net/use-case/ajaxapp/)を参考にしました。


```javascript:distributeResult.js
function main(){
  const url = new URL(window.location.href);
  fetchSearchHistory(url.origin + "/history");
  const query = url.searchParams.get("q");
  if (query){  // queryがなければ終了,あればサーバーからJSON呼び出し
    fetchJSONPath(url.href.replace("search", "json"));  //...(1)
  }
}
```

1. URLのsearchをjsonに変えて、/json APIをたたきます。


## 検索キーワードサジェスト機能
検索履歴を検索フォームに入れて以前の検索キーワードをを探しやすくします。

```javascript:distributeResult.js
function main(){
  const url = new URL(window.location.href);
  fetchSearchHistory(url.origin + "/history");  //...(1)
  const query = url.searchParams.get("q");
  if (query){  // queryがなければ終了,あればサーバーからJSON呼び出し
    fetchJSONPath(url.href.replace("search", "json"));
  }
}
```


```javascript:distributeResult.js
async function fetchSearchHistory(url){
  try{
    const history = await fetchLocatePath(url);
    // 検索キーワード履歴のdatalist <id=search-history>を埋める
    history.forEach((h) =>{
      $("#search-history").append("<option>" + h.word + "</option>");  //...(1)
    });
  } catch(error) {
    console.error(`Error occured (${error})`); // Promiseチェーンの中で発生したエラーを受け取る
  }
}
```

```javascript:datalist.js
$("q").on('input', function () {  //...(2)
    var val = this.value;
    if($('#searched-words option').filter(function(){
        return this.value.toUpperCase() === val.toUpperCase();
    }).length) {
        //send ajax request
        alert(this.value);
    }
});
```

1. history APIをたたき、検索履歴を取得し、検索キーワード候補を検索窓に埋め込みます。
2. 一文字打つたびに、検索履歴をFrecency スコア順に表示します。


## 検索結果の遅延表示
ページ下部付近にくると検索結果を100件ごとに表示します。
なぜ遅延させているかというと、JavaScriptの正規表現が遅いことと、検索結果件数(1~数万件)によってページ読み込み時間が大幅に変わってきてしますためです。
100件ごとに正規表現ハイライトすれば、体感的に待たされる感覚がなくなります。

```javascript:distributeResult.js
async function fetchJSONPath(url){
  try {
    const locaterJSON = await fetchLocatePath(url);
    const locater = new Locater(locaterJSON);  //...(1)
		/* snip...*/
    if (!locater.error) {
			/* snip...*/
      // Rolling next data
      let n = 0;
      const shift = 100;
      locater.displayRoll(n, shift);  //...(2)
      $(window).on("scroll", function(){ // scrollで下限近くまで来ると次をロード  //...(2)
        const inner = $(window).innerHeight();
        const outer = $(window).outerHeight();
        const bottom = inner - outer;
        const tp = $(window).scrollTop();
        if (tp * 1.05 >= bottom) {
          //スクロールの位置が下部5%の範囲に来た場合
          n += shift;
          locater.displayRoll(n, shift);  //...(2)
        }
      });
    } else {
      console.error("error: ", locater.error);
      const err = document.getElementById("error-view");
      err.innerHTML = "<p>" + locater.error + "</p>";
    }
  // 今のところcatchする例外発生ない
  } catch(error) {
    console.error(`Error occured (${error})`); // Promiseチェーンの中で発生したエラーを受け取る
  }
}
```

1. /json APIを非同期に実行し、クラス構文でlocaterを生成します。
2. `locater.displayRoll()`では100件ずつ(n~n+100件)の行をリンクとしてHTMLテンプレートのid=resultに追加していきます。


```javascript:distributeResult.js/displayRoll()
// 検索パス表示
displayRoll(n, shift){
	const folderIcon = '<i class="far fa-folder-open" title="クリックでフォルダを開く"></i>';  //...(1)
	const sep = this.args.pathSplitWin ? "\\" : "/";
	const dataArray = this.paths.slice(n, n + shift);  //...(2)
	dataArray.forEach((p) =>{  //...(2)
		const modified = this.pathModify(p);  //...(3)
		const highlight = this.highlightRegex(modified);  //...(4)
		const dir = Locater.dirname(modified, sep);  //...(5)
		let result = `<a href="file://${modified}">${highlight}</a>`;
		result += `<a href="file://${dir}"> ${folderIcon} </a>`;
		$("#result").append("<tr><td>" + result + "</td></tr>");
	});
}
```

1. フォルダ―アイコンは[Font Awesome](https://fontawesome.com/v4.7/icon/folder-open)からフリーの物を選びました。
2. n~n+shift件ずつ処理します。関数呼び出し時に0~100, 100~200, ... と増えていきます。(sliceだから0~99件目、の処理か。)
3. (指定されていれば)パスのプレフィックスを追加、削除、パスセパレートをUNIX式からWindows式に変更します。
4. 正規表現を用いて、検索ワードの背景を黄色くします。前方一致でハイライトしていますので、`locate`コマンドのマッチとは差異があります。(既知のバグ)
5. ファイルの親ディレクトリをフォルダ―アイコンのリンクに指定します。
6. id=resultに追加していきます。



# デプロイ
Docker, Docker Composeを使用しています。

```dockerfile:Dockerfile
FROM golang:1.17.0-alpine3.14 AS go_official  #...(1)
RUN apk --update --no-cache add git &&\
    go install github.com/u1and0/gocate@v0.3.0  #...(2)
WORKDIR /go/src/github.com/u1and0/locate-server
# For go module using go-pipeline
ENV GO111MODULE=on
COPY ./main.go /go/src/github.com/u1and0/locate-server/main.go
COPY ./go.mod /go/src/github.com/u1and0/locate-server/go.mod
COPY ./go.sum /go/src/github.com/u1and0/locate-server/go.sum
COPY ./cmd /go/src/github.com/u1and0/locate-server/cmd
RUN go build -o /go/bin/locate-server

FROM frolvlad/alpine-glibc:alpine-3.14_glibc-2.33  #...(3)
RUN apk --update --no-cache add mlocate tzdata
WORKDIR /var/www  #...(4)
COPY --from=go_official /go/bin/locate-server /usr/bin/locate-server  #...(5)
COPY --from=go_official /go/bin/gocate /usr/bin/gocate  #...(5)
COPY ./static /var/www/static  #...(4)
COPY ./templates /var/www/templates  #...(4)
ENTRYPOINT ["/usr/bin/locate-server"]
```

1. multistage buildでlocate-serverのバイナリをbuildします。
2. 依存性のあるgocateもインストールします。
3. 実行するコンテナを作成します。
4. HTML, CSS, JS ファイルをコピーします。コマンドの実行ディレクトリと同じ場所にする必要があります。(同じ場所にしないとtemplateが見つからないエラー)
5. goのバイナリをビルドコンテナからコピーしてきます。


```yaml:docker-compose.yml(一例)
version: "3"
services:
    web:
        # image: u1and0/locate-server:latest  #...(1)
        build:  #...(1)
            context: .
        ports:
          - 8081:8080
        volumes:
            - db:/var/lib/mlocate  #...(2)
        environment:
            - TZ=Asia/Tokyo
        working_dir: /var/www
        entrypoint: /usr/bin/locate-server
        # command: ["-debug"]  #...(3)

    db:
        image: busybox
        volumes:
            - /var/lib/mlocate:/var/lib/mlocate  #...(2)

    app:
        build:
             context: ./app
        volumes:
            - db:/var/lib/mlocate  #...(2)

volumes:
    db:
```

| サービス名         | 説明                             |
|---------------|--------------------------------|
| web | locate-server実行コンテナ               |
| db  | appとwebが共有するデータベースコンテナ  |
| app | updatedbをcronで実行するコンテナ |

1. イメージをpullするか、git cloneした後のDockerfileからイメージを作成します。
2. webとappで共有するフォルダを指定します。
3. webコンテナのentrypointが`/usr/bin/locate-server`なので、`locate-server`のオプションはcommandに追記します。

appコンテナはホストマシンでupdatedbを行うなら不要です。
そうするとdbも不要です。直接ホスト上の/var/lib/mlocateディレクトリでもマウントしておけばいいわけですので。

```yaml
web:
    volumes:
        - /var/lib/mlocate:/var/lib/mlocate
```

appコンテナでupdatedbを定期実行させていても、dbコンテナは不要かも、と考えるかもしれません。
appにもwebにも直接ホスト上の/var/lib/mlocateディレクトリをマウントしておけばいいわけですので。

```yaml
web:
    volumes:
        - /var/lib/mlocate:/var/lib/mlocate

app:
    volumes:
        - /var/lib/mlocate:/var/lib/mlocate
```

しかしながら、私の本番環境ではWindows上のVirtualboxでDockerコンテナ立てています。
ホスト上のntfs形式のディレクトリにdb置いたら検索がとてつもなく遅くなってしまったので、あえてdb置く場所はコンテナ上にしております。


# まとめ
