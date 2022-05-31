declare const $: any;

type Args = {
  dbpath: string; // 検索対象DBパス /path/to/database:/path/to/another
  pathSplitWin: boolean; // TrueでWindowsパスセパレータを使用する
  root: string; // 追加するドライブパス名
  trim: string; // 削除するドライブパス名
  debug: boolean; // Debugフラグ
};
type Stats = {
  lastUpdateTime: string; // 最後のDBアップデート時刻
  searchTime: number; // 検索にかかった時間
  items: string; // 検索対象のすべてのファイル数
};

export class Locater {
  args: Args;
  query: string[];
  searchWords: string[];
  excludeWords: string[];
  paths: string[];
  stats: Stats;
  error: string;
  constructor(json) {
    this.args = json.args; // command line argument
    this.query = json.query; // API args
    this.searchWords = json.searchWords; // search word for searching
    this.excludeWords = json.excludeWords; // exclude word for searching
    this.paths = json.paths; // result of locate command
    this.stats = json.stats; // stats info at database
    this.error = json.error; // Error message
  }

  static displayStats(str: string): void {
    const divElem: HTMLElement | null = document.getElementById(
      "search-status",
    );
    if (divElem === null) return;
    const newElem: HTMLElement | null = document.createElement("b");
    newElem.textContent = str;
    divElem.appendChild(newElem);
    const br: HTMLElement | null = document.createElement("br");
    divElem.appendChild(br);
  }

  // 検索パス遅延表示
  lazyLoad(n: number, shift: number): void {
    const folderIcon =
      '<i class="far fa-folder-open" title="クリックでフォルダを開く"></i>';
    const sep: string = this.args.pathSplitWin ? "\\" : "/";
    const dataArray: string[] = this.paths.slice(n, n + shift);
    dataArray.forEach((p: string) => {
      const modified: string = this.pathModify(p);
      const highlight: string = this.highlightRegex(modified);
      const dir: string = Locater.dirname(modified, sep);
      let result = `<a href="file://${modified}">${highlight}</a>`;
      result += `<a href="file://${dir}"> ${folderIcon} </a>`;
      const resultElement: HTMLElement | null = document.getElementById(
        "result",
      );
      // Insert result in element
      const tr = document.createElement("tr");
      const td = document.createElement("td");
      td.innerHTML = result;
      resultElement.appendChild(tr).appendChild(td);
    });
  }

  pathModify(str: string): string {
    if (str.startsWith(this.args.trim)) {
      str = str.slice(this.args.trim.length);
    }
    if (this.args.pathSplitWin) {
      str = str.replaceAll("/", "\\");
    }
    if (this.args.root) {
      str = this.args.root + str;
    }
    return str;
  }

  highlightRegex(str: string): string {
    this.searchWords.forEach((q: string) => {
      const re = new RegExp(q, "i"); // second arg "i" for ignore case
      // $&はreのマッチ結果
      str = str.replace(
        re,
        "<span style='background-color:#FFCC00;'>$&</span>",
      );
    });
    return str;
  }

  static dirname(str: string, sep: string): string {
    const idx: number = str.lastIndexOf(sep); // sep == "/" or "\\"
    return str.slice(0, idx);
  }
}

export function rollingNextData(locater: Locater, n = 0, shift = 100) {
  locater.lazyLoad(n, shift);
  $(window).on("scroll", function () { // scrollで下限近くまで来ると次をロード
    const inner = $(window).innerHeight();
    const outer = $(window).outerHeight();
    const bottom: number = inner - outer;
    const tp = $(window).scrollTop();
    if (tp * 1.05 >= bottom) {
      //スクロールの位置が下部5%の範囲に来た場合
      n += shift;
      locater.lazyLoad(n, shift);
    }
  });
}
