export class Locater {
    args;
    query;
    searchWords;
    excludeWords;
    paths;
    stats;
    error;
    constructor(json) {
        this.args = json.args; // command line argument
        this.query = json.query; // API args
        this.searchWords = json.searchWords; // search word for searching
        this.excludeWords = json.excludeWords; // exclude word for searching
        this.paths = json.paths; // result of locate command
        this.stats = json.stats; // stats info at database
        this.error = json.error; // Error message
    }
    static displayStats(str) {
        const divElem = document.getElementById("search-status");
        if (divElem === null)
            return;
        const newElem = document.createElement("b");
        newElem.textContent = str;
        divElem.appendChild(newElem);
        const br = document.createElement("br");
        divElem.appendChild(br);
    }
    // 検索パス遅延表示
    lazyLoad(n, shift) {
        // Clear child node
        $("#result tr").remove();
        // let clone = resultElement.cloneNode(false);
        // resultElement.parentNode.replaceChild(clone, resultElement);
        const folderIcon = '<i class="far fa-folder-open" title="クリックでフォルダを開く"></i>';
        const sep = this.args.pathSplitWin ? "\\" : "/";
        const dataArray = this.paths.slice(n, n + shift);
        dataArray.forEach((p) => {
            const modified = this.pathModify(p);
            const highlight = this.highlightRegex(modified);
            const dir = Locater.dirname(modified, sep);
            let result = `<a href="file://${modified}">${highlight}</a>`;
            result += `<a href="file://${dir}"> ${folderIcon} </a>`;
            const resultElement = document.getElementById("result");
            // Insert result in element
            const tr = document.createElement("tr");
            const td = document.createElement("td");
            td.innerHTML = result;
            resultElement.appendChild(tr).appendChild(td);
        });
    }
    rollingNextData(n = 0, shift = 100) {
        this.lazyLoad(n, shift);
        $(window).on("scroll", function () {
            const inner = $(window).innerHeight();
            const outer = $(window).outerHeight();
            const bottom = inner - outer;
            const tp = $(window).scrollTop();
            if (tp * 1.05 >= bottom) {
                //スクロールの位置が下部5%の範囲に来た場合
                n += shift;
                this.lazyLoad(n, shift);
            }
        });
    }
    pathModify(str) {
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
    highlightRegex(str) {
        this.searchWords.forEach((q) => {
            const re = new RegExp(q, "i"); // second arg "i" for ignore case
            // $&はreのマッチ結果
            str = str.replace(re, "<span style='background-color:#FFCC00;'>$&</span>");
        });
        return str;
    }
    static dirname(str, sep) {
        const idx = str.lastIndexOf(sep); // sep == "/" or "\\"
        return str.slice(0, idx);
    }
}
