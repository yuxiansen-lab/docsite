/**
 * main.js — 应用入口：加载站点配置，应用主题色与品牌，暴露全局 Site。
 * 在其余模块之后加载（router.js 启动时读取 window.Site）。
 */
(function () {
  "use strict";

  async function boot() {
    let site = {};
    try {
      const res = await fetch("/api/config");
      if (res.ok) site = await res.json();
    } catch {
      /* 离线兜底：使用内置默认值 */
    }

    window.Site = site;

    // 主题色 → CSS 变量
    if (site.themeColor) {
      document.documentElement.style.setProperty("--accent", site.themeColor);
    }

    // 品牌
    if (site.logoText) {
      const logo = document.getElementById("brandLogo");
      if (logo) logo.textContent = site.logoText;
    }
    if (site.title) {
      const name = document.getElementById("brandName");
      if (name) name.textContent = site.title;
      document.title = site.title;
    }
    if (site.description) {
      const meta = document.querySelector('meta[name="description"]');
      if (meta) meta.setAttribute("content", site.description);
    }

    // 配置就绪后再启动路由（router.js 本身不自启）
    if (window.Router) window.Router.start();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", boot);
  } else {
    boot();
  }
})();
