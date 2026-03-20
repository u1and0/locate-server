(() => {
  // src/tooltips.ts
  window.toggleMenu = toggleMenu;
  function toggleMenu(elem) {
    const el = document.getElementById(elem);
    el.style.display = el.style.display === "none" ? "block" : "none";
  }
  function fadeOut(el) {
    el.style.transition = "opacity 0.4s";
    el.style.opacity = "0";
    setTimeout(() => {
      el.style.display = "none";
    }, 400);
  }
  document.addEventListener("DOMContentLoaded", () => {
    const loader = document.querySelector(".loader-wrap");
    if (!loader) return;
    window.addEventListener("load", () => fadeOut(loader));
    setTimeout(() => fadeOut(loader), 5e3);
  });
})();
