import { Locater, rollingNextData } from "./locater.js";
import { fzfSearch } from "./fzf.js";
declare const $: any;

main();

async function main() {
  const url = new URL(window.location.href);
  await fetchSearchHistory(url.origin + "/history");
  const query = url.searchParams.get("q");
  if (!query) { // queryがなければ終了,あればサーバーからJSON呼び出し
    return;
  }
  const locaterJSON: Promise<Locater> = await fetchPath(
    url.href.replace("search", "json"),
  );
  const locater = new Locater(locaterJSON);
  displayResult(locater);

  // FZF on keyboard
  $(function () {
    $("#search-form").keyup(function () {
      const value = document.getElementById("search-form").value;
      const result = fzfSearch(locater.paths, value);
      const locaterClone: Locater = Object.assign(
        Object.create(Object.getPrototypeOf(locater)),
        locater,
      );
      locaterClone.paths = result;
      $("#result tr").remove(); // Clear child node
      rollingNextData(locaterClone);
      console.log(locater);
      console.log(locaterClone);
      // for (const r of result) {
      //   $("#search-result").append($("tr td").html(r))
      // }
    });
  });
}

// fetchの返り値のPromiseを返す
async function fetchPath(url: string): Promise<unknown> {
  return await fetch(url)
    .then((response) => {
      return response.json();
    })
    .catch((response) => {
      return Promise.reject(
        new Error(`{${response.status}: ${response.statusText}`),
      );
    });
}

type History = Array<Frecency>;
type Frecency = { word: string; score: number };

async function fetchSearchHistory(url: string): Promise<void> {
  try {
    const history: History = await fetchPath(url);
    // 検索キーワード履歴のdatalist <id=search-history>を埋める
    const searchHistory: HTMLElement | null = document.getElementById(
      "search-history",
    );
    if (searchHistory === null) return;
    history.forEach((f: Frecency) => {
      searchHistory.append("<option>" + f.word + "</option>");
    });
  } catch (error) {
    console.error(`Error occured (${error})`); // Promiseチェーンの中で発生したエラーを受け取る
  }
}

function displayResult(locater: Locater): void {
  if (locater.error) {
    console.error("error: ", locater.error);
    const err: HTMLElement | null = document.getElementById("error-view");
    if (err === null) return;
    err.innerHTML = "<p>" + locater.error + "</p>";
    return;
  }
  const hitCount = `ヒット数: ${locater.paths.length}件`;
  Locater.displayStats(hitCount);
  const searchTime = `${locater.stats.searchTime.toFixed(3)}msec で\
                        約${locater.stats.items}件を検索しました。`;
  Locater.displayStats(searchTime);
  $("#result tr").remove(); // Clear child node
  rollingNextData(locater);
}
