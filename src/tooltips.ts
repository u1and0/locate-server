// toggleMenu をグローバルに公開（HTML の onclick 属性から呼ばれるため）
(window as unknown as Record<string, unknown>).toggleMenu = toggleMenu;

export function toggleMenu(elem: string): void {
  const el = document.getElementById(elem)!;
  el.style.display = el.style.display === "none" ? "block" : "none";
}

function fadeOut(el: HTMLElement): void {
  el.style.transition = "opacity 0.4s";
  el.style.opacity = "0";
  setTimeout(() => {
    el.style.display = "none";
  }, 400);
}

document.addEventListener("DOMContentLoaded", () => {
  const loader = document.querySelector<HTMLElement>(".loader-wrap");
  if (!loader) return;

  window.addEventListener("load", () => fadeOut(loader));
  setTimeout(() => fadeOut(loader), 5000);
});
