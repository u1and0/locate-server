main()

function main(){
  const url = new URL(window.location.href);
  fetchSearchHistory(url.origin + "/history");
  const query = url.searchParams.get("q");
  if (query){  // queryがなければ終了,あればサーバーからJSON呼び出し
    fetchJSONPath(url.href.replace("search", "json"));
  }
}

class Locater {
  constructor(json){
    this.paths = json.paths;
    this.args = json.args;
    this.stats = json.stats;
    this.searchWords = json.searchWords;
    this.excludeWords = json.excludeWords;
  }

  static displayStats(str){
    const divElem = document.getElementById("search-status");
    const newElem = document.createElement("b");
    newElem.textContent = str;
    divElem.appendChild(newElem);
    const br = document.createElement("br");
    divElem.appendChild(br);
  }

  // 検索パス表示
  displayRoll(n, shift){
    const folderIcon = '<i class="far fa-folder-open" title="クリックでフォルダを開く"></i>';
    const sep = this.args.pathSplitWin ? "\\" : "/";
    const dataArray = this.paths.slice(n, n + shift);
    // $.each(dataArray, function(i){
    dataArray.forEach((p) =>{
      const modified = pathModify(p, this.args);
      const highlight = highlightRegex(modified, this.searchWords);
      const dir = dirname(modified, sep);
      let result = `<a href="file://${modified}">${highlight}</a>`;
      result += `<a href="file://${dir}"> ${folderIcon} </a>`;
      $("#result").append("<tr><td>" + result + "</td></tr>");
    });
  }
}

async function fetchSearchHistory(url){
  try{
    const searchHistoryJSON = await fetchLocatePath(url);
    console.dir(searchHistoryJSON)
    fillSearchHistory(searchHistoryJSON)
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
    // locater.fillSearchHistory();  // 検索キーワード履歴のdatalist <id=search-history>を埋める
    const hitCount = `ヒット数: ${locater.paths.length}件`;
    Locater.displayStats(hitCount);
    const searchTime = `${locater.stats.searchTime.toFixed(3)}msec で\
                        約${locater.stats.items}件を検索しました。`;
    Locater.displayStats(searchTime);
    // Rolling next data
    let n = 0;
    const shift = 100;
    locater.displayRoll(n, shift);
    $(window).on("scroll", function(){ // scrollで下限近くまで来ると次をロード
      const inner = $(window).innerHeight();
      const outer = $(window).outerHeight();
      const bottom = inner - outer;
      const tp = $(window).scrollTop();
      const ob = {
        "inner": inner,
        "outer": outer,
        "bottom": bottom,
        "tp": tp,
      }
      if (locater.args.debug){
        console.log("scroll position: ", ob);
      }
      if (tp * 1.05 >= bottom) {
        //スクロールの位置が下部5%の範囲に来た場合
        n += shift;
        locater.displayRoll(n, shift);
      }
    });
  } catch(error) {
    console.error(`Error occured (${error})`); // Promiseチェーンの中で発生したエラーを受け取る
  }
}

// fetchの返り値のPromiseを返す
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

function pathModify(str, args){
  if (str.startsWith(args.trim)){
    str = str.slice(args.trim.length);
  }
  if (args.pathSplitWin){
    str = str.replaceAll("/", "\\");
  }
  if (args.root){
    str = args.root + str;
  }
  return str;
}

function highlightRegex(str, searchWords){
  searchWords.forEach((q) =>{
    const re = new RegExp(q, "i"); // second arg "i" for ignore case
    // $&はreのマッチ結果
    str = str.replace(re, "<span style='background-color:#FFCC00;'>$&</span>");
  })
  return str;
}

function dirname(str, sep){
  const idx = str.lastIndexOf(sep); // sep == "/" or "\\"
  return str.slice(0,idx);
}

// 検索キーワード履歴のdatalist <id=search-history>を埋める
function fillSearchHistory(json){
  json.forEach((h) =>{
    if (h.word) {
      $("#search-history").append("<option>" + h.word + "</option>");
    }
  });
}
