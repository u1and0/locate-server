async function main(){
  try {
    const query = document.getElementById("q").value;
    const resultPath = await fetchLocatePath(query);
    const view = createView(resultPath);
    displayView(view);
    // .then((userInfo) => createView(userInfo)) // JSONオブジェクトで解決されるPromise
    // .then((view)=> displayView(view))  // HTML文字列で解決されるPromise
  } catch(error) {
    console.error(`Error occured (${error})`); // Promiseチェーンの中で発生したエラーを受け取る
  }
}

// データの取得だけ行うように変更
function fetchLocatePath(query){
  // fetchの返り値のPromiseを返す
  // encodeURIComponentは/や%など ただの文字列として扱えるようにエスケープする
  return fetch(`http://localhost:8080/json?q=bin+ls`)
    .then(response =>{
      console.log(response.status);
      if (!response.ok) {
        // console.error("Error response", response);
        return Promise.reject(new Error(`{${response.status}: ${response.statusText}`));
      } else{
        return response.json(); //.then(userInfo =>  ここはmain()で解決
      }
      console.log(query);
    });
}
//
// HTMLの組立
function createView(resultPath){
  // タグ関数escapeHTMLの呼び出しは``の前
  return escapeHTML`
  <h4>${resultPath.query}</h4>
  <table>
      <tr>
        <td>${resultPath.paths[0]}</td>
      </tr>
  </table>
  `;
}

// HTMLの挿入
function displayView(view){
  const result = document.getElementById("result");
  result.innerHTML = view;
}

// escapeHTML``から呼び出される
function escapeSpecialChars(str){
  return str
  .replace(/&/g, "&amp;")
  .replace(/</g, "&lt;")
  .replace(/>/g, "&gt;")
  .replace(/"/g, "&quot;")
  .replace(/'/g, "&#039;");
}

// テンプレートリテラルをタグ付けすることで、
// 明示的にエスケープ用の関数を呼び出す必要がないようにする
// タグ関数
// 第一引数に文字列リテラルの配列
// 第二引数に埋め込まれる値の配列
// 値が文字列であればエスケープする
// タグ関数はテンプレートリテラルに対してタグ付けして使う
function escapeHTML(strings, ...values){
  return strings.reduce((result, str, i) => {
    const value = values[i-1];
    if (typeof value === "string"){
      return result + escapeSpecialChars(value) + str;
    }else {
      return result + String(value) + str;
    }
  });
}
