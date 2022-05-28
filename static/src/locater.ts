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
