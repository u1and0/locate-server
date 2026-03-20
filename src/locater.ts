interface LocaterArgs {
  dbpath: string;
  pathSplitWin: boolean;
  root: string;
  trim: string;
  debug: boolean;
}

interface LocaterQuery {
  q: string;
  logging: boolean;
  limit: number;
}

interface LocaterStats {
  lastUpdateTime: string;
  searchTime: number;
  items: string;
}

export interface LocaterJSON {
  args: LocaterArgs;
  query: LocaterQuery;
  searchWords: string[];
  excludeWords: string[];
  paths: string[];
  stats: LocaterStats;
  error: string;
}

export class Locater {
  args: LocaterArgs;
  query: LocaterQuery;
  searchWords: string[];
  excludeWords: string[];
  paths: string[];
  stats: LocaterStats;
  error: string;

  constructor(json: LocaterJSON) {
    this.args = json.args;
    this.query = json.query;
    this.searchWords = json.searchWords;
    this.excludeWords = json.excludeWords;
    this.paths = json.paths;
    this.stats = json.stats;
    this.error = json.error;
  }

  static displayStats(str: string): void {
    const divElem = document.getElementById("search-status")!;
    const newElem = document.createElement("b");
    newElem.textContent = str;
    divElem.appendChild(newElem);
    divElem.appendChild(document.createElement("br"));
  }

  lazyLoad(n: number, shift: number): void {
    const folderIcon =
      '<i class="far fa-folder-open" title="クリックでフォルダを開く"></i>';
    const sep = this.args.pathSplitWin ? "\\" : "/";
    const resultTable = document.getElementById("result")!;
    this.paths.slice(n, n + shift).forEach((p) => {
      const modified = this.pathModify(p);
      const highlight = this.highlightRegex(modified);
      const dir = Locater.dirname(modified, sep);
      const html = `<tr><td>` +
        `<a href="file://${modified}">${highlight}</a>` +
        `<a href="file://${dir}"> ${folderIcon} </a>` +
        `</td></tr>`;
      resultTable.insertAdjacentHTML("beforeend", html);
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
    this.searchWords.forEach((q) => {
      const re = new RegExp(q, "i");
      str = str.replace(
        re,
        "<span style='background-color:#FFCC00;'>$&</span>",
      );
    });
    return str;
  }

  static dirname(str: string, sep: string): string {
    return str.slice(0, str.lastIndexOf(sep));
  }
}
