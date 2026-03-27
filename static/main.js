(() => {
  var __defProp = Object.defineProperty;
  var __defNormalProp = (obj, key, value) => key in obj ? __defProp(obj, key, { enumerable: true, configurable: true, writable: true, value }) : obj[key] = value;
  var __publicField = (obj, key, value) => __defNormalProp(obj, typeof key !== "symbol" ? key + "" : key, value);

  // src/locater.ts
  var Locater = class _Locater {
    constructor(json) {
      __publicField(this, "args");
      __publicField(this, "query");
      __publicField(this, "searchWords");
      __publicField(this, "excludeWords");
      __publicField(this, "paths");
      __publicField(this, "stats");
      __publicField(this, "error");
      this.args = json.args;
      this.query = json.query;
      this.searchWords = json.searchWords;
      this.excludeWords = json.excludeWords;
      this.paths = json.paths;
      this.stats = json.stats;
      this.error = json.error;
    }
    static displayStats(str) {
      const divElem = document.getElementById("search-status");
      const newElem = document.createElement("b");
      newElem.textContent = str;
      divElem.appendChild(newElem);
      divElem.appendChild(document.createElement("br"));
    }
    lazyLoad(n, shift) {
      const folderIcon = '<i class="far fa-folder-open" title="\u30AF\u30EA\u30C3\u30AF\u3067\u30D5\u30A9\u30EB\u30C0\u3092\u958B\u304F"></i>';
      const sep = this.args.pathSplitWin ? "\\" : "/";
      const resultTable = document.getElementById("result");
      this.paths.slice(n, n + shift).forEach((p) => {
        const modified = this.pathModify(p);
        const highlight = this.highlightRegex(modified);
        const dir = _Locater.dirname(modified, sep);
        const html = `<tr><td><a href="file://${modified}">${highlight}</a><a href="file://${dir}"> ${folderIcon} </a></td></tr>`;
        resultTable.insertAdjacentHTML("beforeend", html);
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
        const re = new RegExp(q, "i");
        str = str.replace(
          re,
          "<span style='background-color:#FFCC00;'>$&</span>"
        );
      });
      return str;
    }
    static dirname(str, sep) {
      return str.slice(0, str.lastIndexOf(sep));
    }
  };

  // src/main.ts
  main();
  function main() {
    const url = new URL(window.location.href);
    fetchSearchHistory(url.origin + "/history");
    const query = url.searchParams.get("q");
    if (query) {
      fetchJSONPath(url.href.replace("search", "json"));
    }
  }
  function fetchLocatePath(url) {
    return fetch(url).then((response) => response.json());
  }
  async function fetchSearchHistory(url) {
    try {
      const history = await fetchLocatePath(url);
      const datalist = document.getElementById("search-history");
      history.forEach((h) => {
        const option = document.createElement("option");
        option.value = h.word;
        datalist.appendChild(option);
      });
    } catch (error) {
      console.error(`Error occured (${error})`);
    }
  }
  async function fetchJSONPath(url) {
    try {
      const locaterJSON = await fetchLocatePath(url);
      const locater = new Locater(locaterJSON);
      if (locater.args.debug) {
        console.dir(locater);
      }
      if (!locater.error) {
        Locater.displayStats(`\u30D2\u30C3\u30C8\u6570: ${locater.paths.length}\u4EF6`);
        Locater.displayStats(
          `${locater.stats.searchTime.toFixed(3)}msec \u3067\u691C\u7D22\u3057\u307E\u3057\u305F\u3002`
        );
        let n = 0;
        const shift = 100;
        locater.lazyLoad(n, shift);
        window.addEventListener("scroll", () => {
          const bottom = window.innerHeight - window.outerHeight;
          if (window.scrollY * 1.05 >= bottom) {
            n += shift;
            locater.lazyLoad(n, shift);
          }
        });
      } else {
        console.error("error: ", locater.error);
        const err = document.getElementById("error-view");
        err.innerHTML = "<p>" + locater.error + "</p>";
      }
    } catch (error) {
      console.error(`Error occured (${error})`);
    }
  }
})();
