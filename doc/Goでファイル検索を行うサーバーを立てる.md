ファイルパスを検索し結果をJSONで返すREST APIサーバーを立てます。
ひとまず動きのイメージを掴むデモです。

![out](https://user-images.githubusercontent.com/16408916/143503512-6e172a98-f973-4c80-b1dc-99ea0ede0a71.gif)

検索窓に検索キーワードを入力し、検索ボタンを押すとlocateコマンドを走らせて、結果をブラウザに表示します。

本記事は[locate-server](https://github.com/u1and0/locate-server) v3.1.0の時点のREADMEを補完するドキュメントを記事としました。


## 実行
### 前提条件
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


### サーバーサイド

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


### クライアントサイド
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

## 内部動作
### 動作概要
* ウェブブラウザからの入力で指定ディレクトリ下にあるファイル内の文字列に対してlocateコマンドを使用した正規表現検索を行い、結果をJSONにしてクライアントに送ります。
* JSONを受け取ったクライアントは、static下に配置されたJavaScriptファイルでHTMLに変換して描画します。

### ディレクトリ構造
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

### ページの表示

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
	route.GET("/json", fetchJSON)  //...(5)
	route.GET("/status", fetchStatus)  //...(5)

	// Listen and serve on 0.0.0.0:8080
	route.Run(":" + strconv.Itoa(port)) // => :8080

// Listen and serve on 0.0.0.0:8080
route.Run(":" + port) //...(6)
```

1. フレームワークは[gin](https://github.com/gin-gonic/gin)を使っています。
2. css, JavaScript, favicon用pngファイルはstaticに置いてあります。
3. トップページの表示はtemplates/index.htmlに`gin.H{}`構造体の内容を埋め込んで表示します。
4. トップページと検索ページは結果の表示がされているかどうかだけで、同じテンプレートを使用します。
5. APIを今のところ3つ用意しています。いずれもGETメソッドです。
* /historyは検索履歴をFrecencyスコア順にしてJSONで取得します。
* /jsonはlocate検索を走らせて検索結果をJSONで取得します。
* /statusはDBの`locate -S`の出力をJSONで取得します。
6. デフォルトでは8080ポートでサーバーを公開します。


### 検索結果を返すページ
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


### 検索結果を返すAPIの実装
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
		paths, err = l.Locate()
		(*cache)[k] = &paths  //...(2)
	} else {
		paths = *v  //...(3)
	}
	return
}
```

1. 検索語(SearchWords)と除外語(ExcludeWords)を正規化(Normalize)して、MapのKeyとします。
2. cacheからキーワードkを探して、結果がなければlocate(gocate)コマンドで検索します。(後述)
3. cacheからキーワードkを探して、結果があればその値vを返します。


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
	locate = append(locate, "--regex", strings.Join(l.SearchWords, ".*"))  //...(2)

	pipeline = append(pipeline, locate)

	// Exclude PATTERNs
	for _, ex := range l.ExcludeWords {
		// COMMAND | grep -ivE EXCLUDE1 | grep -ivE EXCLUDE2
		pipeline = append(pipeline, []string{"grep", "-ivE", ex})  //...(3)
	}

	// Limit option
	if l.Query.Limit > 0 {
		pipeline = append(pipeline, []string{"head", "-n", strconv.Itoa(l.Query.Limit)})  //...(4)
	}

	if l.Args.Debug {
		log.Debugf("Execute command %v", pipeline)
	}
	return  // => locate ... | grep -ivE ... | head -n ... ...(5)
}
```

1. `l.CmdGen()`でコマンド文字列の生成を行い、locate(gocate)コマンドを実行します。`pipeline.Output()`の結果は[]byteで返ってくるので、[]stringに変えて返却します。
2. locate(gocate)コマンドの文字列を生成します。`l.SearchWords`はSliceなので、locateに渡せるように".\*"を挟みます。
3. 除外するキーワードを`grep -v`で排除します。`-i`でignore case, `-E`で正規表現。
4. 出力行数を`head -n`で制御します。テスト環境(Linux上のDocker)ではうまくいって、本番環境(Windows上のDocker)でうまく動いていないような...。`gocate`を改造して--limitオプション付けるか思案中。
5. 最終的にコマンドラインに入力する文字列 `locate "検索語" | grep -ivE "除外語" | head -n "結果上限数" を返します。


### 検索履歴のスコアを返すAPIの実装
> locater-server v3.1.0以降
/historyで返す検索履歴のfrecency(Frequently+recency)
アクセスを解析しスコアリングをリスト・オブ・オブジェクトの形式のJSONで返します。

```go:main.go/history
route.GET("/history", func(c *gin.Context) {
	searchHistory, err := cmd.Datalist(LOGFILE)  //...(1)
	if err != nil {
		log.Error(err)
		c.JSON(404, searchHistory)
	}
	if locater.Args.Debug {
		log.Debug(searchHistory)
	}
	c.JSON(http.StatusOK, searchHistory)
})
```

Scoreは

```go:frecency.go
// Scoring : 日時から頻出度を算出する
func Scoring(t time.Time) int {
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

```


### 


### 

## API

サーバーを立ち上げた状態で

```shell-session
$ curl -fsSL localhost:8080/search?q=u1and0+locate+go
```

とすると、検索結果をJSONにして標準出力に表示します。

APIについてはv3.0.0より上位で色々出来るよう拡張中です。



## デプロイ
docker



###################

## JavaScript 上のエラーハンドリング

```javascript
function fetchLocatePath(url){
  return fetch(url)
    .then(response =>{
      if (!response.ok) {
        return Promise.reject(new Error(`{${response.status}: ${response.statusText}`));
      } else{
        return response.json(); //.then(userInfo =>  ここはmain()で解決
      }
    });
}
```

statusがOK(code 200)ではないときは常にPromise.rejectで拒否されます。

response.statusでコード100~500番のコードを取得します。。
response.statusTextにはコードの詳しい内容(404ならNot Foundとか)が書かれています。


[](https://ja.javascript.info/promise-error-handling?utm_source=pocket_mylist)の例に見られるように、statusをチェックしたあと、200でなければエラーをthrowするようにします。

```javascript
class HttpError extends Error { // (1)
  constructor(response) {
    super(`${response.status} for ${response.url}`);
    this.name = 'HttpError';
    this.response = response;
  }
}

function loadJson(url) {
  return fetch(url)
    .then(response => {
      if (response.status == 200) { // (2)
        return response.json();
      } else {
        throw new HttpError(response);
      }
    })
}
```

1. HTTPerror カスタムクラス
1. 非200ステータスをエラーとする

使い方は以下の(2)のところ。

```javascript
function demoGithubUser() {
  let name = prompt("Enter a name?", "iliakan");

  return loadJson(`https://api.github.com/users/${name}`)
    .then(user => {
      alert(`Full name: ${user.name}.`); // (1)
      return user;
    })
    .catch(err => {
      if (err instanceof HttpError && err.response.status == 404) { // (2)
        alert("No such user, please reenter.");
        return demoGithubUser();
      } else {
        throw err;
      }
    });
}

demoGithubUser();
```

