import { Locater, type LocaterJSON } from "./locater.ts";

main();

function main(): void {
  const url = new URL(window.location.href);
  fetchSearchHistory(url.origin + "/history");
  const query = url.searchParams.get("q");
  if (query) {
    fetchJSONPath(url.href.replace("search", "json"));
  }
}

function fetchLocatePath(url: string): Promise<unknown> {
  return fetch(url).then((response) => response.json());
}

async function fetchSearchHistory(url: string): Promise<void> {
  try {
    const history = (await fetchLocatePath(url)) as Array<{ word: string }>;
    const datalist = document.getElementById("search-history")!;
    history.forEach((h) => {
      const option = document.createElement("option");
      option.value = h.word;
      datalist.appendChild(option);
    });
  } catch (error) {
    console.error(`Error occured (${error})`);
  }
}

async function fetchJSONPath(url: string): Promise<void> {
  try {
    const locaterJSON = (await fetchLocatePath(url)) as LocaterJSON;
    const locater = new Locater(locaterJSON);
    if (locater.args.debug) {
      console.dir(locater);
    }
    if (!locater.error) {
      Locater.displayStats(`ヒット数: ${locater.paths.length}件`);
      Locater.displayStats(
        `${locater.stats.searchTime.toFixed(3)}msec で` +
          `約${locater.stats.items}件を検索しました。`,
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
      const err = document.getElementById("error-view")!;
      err.innerHTML = "<p>" + locater.error + "</p>";
    }
  } catch (error) {
    console.error(`Error occured (${error})`);
  }
}
