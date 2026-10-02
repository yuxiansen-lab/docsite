/**
 * theme.js — 亮/暗主题：跟随系统、手动切换、localStorage 持久化。
 * 暴露 window.Theme.{toggle, get, apply, icon}
 */
(function () {
  "use strict";

  const KEY = "docsite:theme"; // "light" | "dark" | "auto"
  const root = document.documentElement;

  const sunSVG =
    '<svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"/></svg>';
  const moonSVG =
    '<svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>';

  function systemPrefersDark() {
    return (
      window.matchMedia &&
      window.matchMedia("(prefers-color-scheme: dark)").matches
    );
  }

  function get() {
    try {
      return localStorage.getItem(KEY) || "auto";
    } catch {
      return "auto";
    }
  }

  function resolve(mode) {
    if (mode === "dark" || mode === "light") return mode;
    return systemPrefersDark() ? "dark" : "light";
  }

  function apply(mode) {
    root.setAttribute("data-theme", resolve(mode));
    updateIcon();
  }

  // 按钮显示"将要切换到的目标主题"图标
  function updateIcon() {
    const btn = document.getElementById("btnTheme");
    if (!btn) return;
    const dark = root.getAttribute("data-theme") === "dark";
    btn.innerHTML = dark ? sunSVG : moonSVG;
    btn.title = dark ? "切换到亮色" : "切换到暗色";
    btn.setAttribute("aria-label", dark ? "切换到亮色" : "切换到暗色");
  }

  function toggle() {
    const cur = resolve(get());
    const next = cur === "dark" ? "light" : "dark";
    try {
      localStorage.setItem(KEY, next);
    } catch {
      /* ignore */
    }
    apply(next);
  }

  // 绑定切换按钮：此前遗漏了这一步，导致点击没有任何反应
  function bindToggle() {
    const btn = document.getElementById("btnTheme");
    if (!btn || btn.dataset.themeBound === "1") return;
    btn.dataset.themeBound = "1";
    btn.addEventListener("click", toggle);
  }

  // 跟随系统变化（仅在 auto 模式）
  if (window.matchMedia) {
    window
      .matchMedia("(prefers-color-scheme: dark)")
      .addEventListener("change", () => {
        if (get() === "auto") apply("auto");
      });
  }

  window.Theme = { toggle, get, apply, resolve, bind: bindToggle };

  // 立即应用，避免 FOUC
  apply(get());

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", bindToggle);
  } else {
    bindToggle();
  }
})();
