main()

function main(){
  const query = getQ();
  if (!query) { // queryが""やnullや<empty string>のときは何もしない
    return
  }
  fetchJSONPath(query);
}

class Locater {
  constructor(json){
    this.paths = json.paths;
    this.args = json.args;
    this.stats = json.stats;
    this.searchWords = json.searchWords;
    this.excludeWords = json.excludeWords;
  }

  // ヒット件数表示
  displayHitCount(){
    const divElem = document.getElementById("search-status");
    const newElem = document.createElement("b");
    const len = this.paths.length;
    newElem.textContent = `ヒット数: ${len}件`;
    divElem.appendChild(newElem);
    const br = document.createElement("br");
    divElem.appendChild(br);
  }

  // 検索件数表示
  displaySearchTime(){
    const divElem = document.getElementById("search-status");
    const newElem = document.createElement("b");
    const searchTime = this.stats.searchTime.toFixed(3);
    newElem.textContent = `${searchTime}msec で約${this.stats.items}件を検索しました。`;
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
      table.insertAdjacentHTML('beforeend', `<tr><td>${result}</tr></td>`);
    });
  }
}

async function fetchJSONPath(query){
  try {
      const locaterJSON = await fetchLocatePath(query);
      const locater = new Locater(locaterJSON);
      console.log(locater);
      locater.displayHitCount();
      locater.displaySearchTime();
      locater.displayView();
  } catch(error) {
    console.error(`Error occured (${error})`); // Promiseチェーンの中で発生したエラーを受け取る
  }
}

function getQ() {
  const url = new URL(window.location.href);
  const params = url.searchParams;
  return params.get("q");
}

// fetchの返り値のPromiseを返す
function fetchLocatePath(query){
  const url="http://localhost:8080"
  return fetch(`${url}/json?q=${query.split(" ").join("+")}`)
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
