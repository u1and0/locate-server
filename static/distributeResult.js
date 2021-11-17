main()

function main(){
  const url = new URL(window.location.href);
  const query = url.searchParams.get("q");
  if (!query) { // queryが""やnullや<empty string>のときは何もしない
    return
  }
  fetchJSONPath(url);
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
  displayView(){
    const folderIcon = '<i class="far fa-folder-open" title="クリックでフォルダを開く"></i>';
    const table = document.getElementById("result");
    const sep = this.args.pathSplitWin ? "\\" : "/";
    this.paths.forEach((p) =>{
      let modified = pathModify(p, this.args);
      let highlight = highlightRegex(modified, this.searchWords);
      let dir = dirname(modified, sep);
      let result = `<a href="file://${modified}">${highlight}</a>`;
      result += `<a href="file://${dir}"> ${folderIcon} </a>`;
      table.insertAdjacentHTML('beforeend', `<tr><td>${result}</td></tr>`);
    });
  }
}

async function fetchJSONPath(url){
  try {
    const jsonURL = url.href.replace("search", "json")
    const locaterJSON = await fetchLocatePath(jsonURL);
    const locater = new Locater(locaterJSON);
    console.log(locater);
    const hitCount = `ヒット数: ${locater.paths.length}件`;
    Locater.displayStats(hitCount);
    const searchTime = `${locater.stats.searchTime.toFixed(3)}msec で\
                        約${locater.stats.items}件を検索しました。`;
    Locater.displayStats(searchTime);
    locater.displayView();
    $(function(){
      $("#result").pagination({
        itemElement: "> td"
      });
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
    let re = new RegExp(q, "i"); // second arg "i" for ignore case
    // $&はreのマッチ結果
    str = str.replace(re, "<span style='background-color:#FFCC00;'>$&</span>");
  })
  return str;
}

function dirname(str, sep){
  const idx = str.lastIndexOf(sep); // sep == "/" or "\\"
  return str.slice(0,idx);
}
