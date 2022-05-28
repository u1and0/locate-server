main()

function main(){
  const url = new URL(window.location.href);
  fetchSearchHistory(url.origin + "/history");
  const query = url.searchParams.get("q");
  if (query){  // queryがなければ終了,あればサーバーからJSON呼び出し
    fetchJSONPath(url.href.replace("search", "json"));
  }
}

// fetchの返り値のPromiseを返す
function fetchLocatePath(url){
  return fetch(url)
    .then(response =>{
      // if (!response.ok) {
        // return Promise.reject(new Error(`{${response.status}: ${response.statusText}`));
      // } else{
      return response.json();
    });
}


async function fetchSearchHistory(url){
  try{
    const history = await fetchLocatePath(url);
    // 検索キーワード履歴のdatalist <id=search-history>を埋める
    history.forEach((h) =>{
      $("#search-history").append("<option>" + h.word + "</option>");
    });
  } catch(error) {
    console.error(`Error occured (${error})`); // Promiseチェーンの中で発生したエラーを受け取る
  }
}

async function fetchJSONPath(url){
  try {
    const locaterJSON = await fetchLocatePath(url);
    const locater = new Locater(locaterJSON);
    if (locater.args.debug){
      console.dir(locater);
    }
    if (!locater.error) {
      const hitCount = `ヒット数: ${locater.paths.length}件`;
      Locater.displayStats(hitCount);
      const searchTime = `${locater.stats.searchTime.toFixed(3)}msec で\
                          約${locater.stats.items}件を検索しました。`;
      Locater.displayStats(searchTime);
      // Rolling next data
      let n = 0;
      const shift = 100;
      locater.lazyLoad(n, shift);
      $(window).on("scroll", function(){ // scrollで下限近くまで来ると次をロード
        const inner = $(window).innerHeight();
        const outer = $(window).outerHeight();
        const bottom = inner - outer;
        const tp = $(window).scrollTop();
        if (tp * 1.05 >= bottom) {
          //スクロールの位置が下部5%の範囲に来た場合
          n += shift;
          locater.lazyLoad(n, shift);
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

