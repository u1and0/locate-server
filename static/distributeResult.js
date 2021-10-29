main()

async function main(){
  try {
    const query = getQ();
    if (!query) { // queryが""やnullや<empty string>のときは何もしない
      return
    }
      const resultPath = await fetchLocatePath(query);
      console.log(resultPath);
      displayView(resultPath);
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
  return fetch(`${url}/json?q=${makeQuery(query)}`)
    .then(response =>{
      console.log(response.status);
      if (!response.ok) {
        // console.error("Error response", response);
        return Promise.reject(new Error(`{${response.status}: ${response.statusText}`));
      } else{
        return response.json(); //.then(userInfo =>  ここはmain()で解決
      }
    });
}

function makeQuery(str){
  return str.split(" ").join("+")
}

// HTMLの挿入
function displayView(view){
  const table = document.getElementById("result");
  // result.innerHTML = view;
  view.paths.forEach((p) =>{
      let newRow = table.insertRow();
      let newCell = newRow.insertCell();
      newCell.appendChild(document.createTextNode(p));
  });
}
