async function main(){
  try {
    const query = document.getElementById("q").value;
    const resultPath = await fetchLocatePath(query);
    displayView(resultPath);
  } catch(error) {
    console.error(`Error occured (${error})`); // Promiseチェーンの中で発生したエラーを受け取る
  }
}

// fetchの返り値のPromiseを返す
function fetchLocatePath(query){
  return fetch(`http://localhost:8080/json?q=${makeQuery(query)}`)
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
