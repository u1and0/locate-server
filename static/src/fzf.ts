import { Fzf } from "../node_modules/fzf/dist/fzf.es.js";
export const fzfSearch = (list:string[], keyword:string)=>{
  const fzf = new Fzf(list)
  const entries = fzf.find(keyword)
  const ranking:string[] = entries.map((entry:Fzf)=> entry.item)
  return ranking
}
