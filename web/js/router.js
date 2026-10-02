/**
 * router.js — 轻量 hash 路由 + 文档加载。
 * URL 形如 #/guide/intro.md；负责：
 *  - 解析当前路径
 *  - 拉取 /api/doc?path=...
 *  - 渲染正文、TOC、上下篇、面包屑
 *  - 触发侧边栏选中态与动画
 *
 * start() 由 main.js 在站点配置就绪后调用（避免读到空 defaultDoc）。
 */
(function () {
  "use strict";

  const content = document.getElementById("content");
  const tocEl = document.getElementById("toc");
  const docnav = document.getElementById("docnav");
  const crumb = document.getElementById("crumb");

  let currentPath = null;
  let aborter = null;
  let started = false;

  // 只有 "#/..." 形式才算文档路由；"#锚点" 是页内跳转，必须忽略，
  // 否则点击右侧 TOC 会被误当成文档路径去请求。
  function pathFromHash() {
    const h = location.hash;
    if (!h.startsWith("#/")) return null;
    const p = h.slice(2);
    return p || null;
  }

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, (c) => ({
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#39;",
    }[c]));
  }

  function showSkeleton() {
    content.innerHTML =
      '<div class="skeleton"><div class="skeleton__title"></div>' +
      '<div class="skeleton__line w80"></div>' +
      '<div class="skeleton__line w100"></div>' +
      '<div class="skeleton__line w60"></div></div>';
    tocEl.hidden = true;
    docnav.innerHTML = "";
  }

  function renderToc(toc) {
    if (!toc || !toc.length) {
      tocEl.hidden = true;
      return;
    }
    tocEl.hidden = false;
    tocEl.innerHTML =
      '<div class="toc__title">目录</div><div class="toc__list">' +
      toc
        .map(
          (t) =>
            `<a class="toc__item" data-level="${t.level}" data-target="${escapeHtml(
              t.id
            )}" href="#${escapeHtml(t.id)}">${escapeHtml(t.text)}</a>`
        )
        .join("") +
      "</div>";

    const scroller = document.querySelector(".content-wrap");
    const items = Array.from(tocEl.querySelectorAll(".toc__item"));
    const map = new Map();
    items.forEach((a) => map.set(a.dataset.target, a));

    // 点击只做平滑滚动，不改动 location.hash，
    // 以免破坏 "#/文档.md" 这条路由（刷新后仍能回到当前文档）。
    items.forEach((a) => {
      a.addEventListener("click", (e) => {
        e.preventDefault();
        const el = document.getElementById(a.dataset.target);
        if (el) el.scrollIntoView({ behavior: "smooth", block: "start" });
      });
    });

    const obs = new IntersectionObserver(
      (entries) => {
        entries.forEach((e) => {
          if (e.isIntersecting) {
            items.forEach((i) => i.classList.remove("is-active"));
            const a = map.get(e.target.id);
            if (a) a.classList.add("is-active");
          }
        });
      },
      { root: scroller, rootMargin: "-60px 0px -70% 0px", threshold: 0 }
    );
    toc.forEach((t) => {
      const el = document.getElementById(t.id);
      if (el) obs.observe(el);
    });
    tocEl._obs = obs;
  }

  function renderDocnav(doc) {
    const parts = [];
    if (doc.prev) {
      parts.push(
        `<a class="docnav__card docnav__card--prev" href="#/${encodeURIComponent(doc.prev.path)}">
          <span class="docnav__label">← 上一篇</span>
          <span class="docnav__title">${escapeHtml(doc.prev.title)}</span>
        </a>`
      );
    } else {
      parts.push("<span></span>");
    }
    if (doc.next) {
      parts.push(
        `<a class="docnav__card docnav__card--next" href="#/${encodeURIComponent(doc.next.path)}">
          <span class="docnav__label">下一篇 →</span>
          <span class="docnav__title">${escapeHtml(doc.next.title)}</span>
        </a>`
      );
    } else {
      parts.push("<span></span>");
    }
    docnav.innerHTML = parts.join("");
  }

  function setCrumb(title) {
    crumb.innerHTML = `DocSite&nbsp; / &nbsp;<strong>${escapeHtml(title)}</strong>`;
  }

  async function load(path) {
    if (aborter) aborter.abort();
    aborter = new AbortController();
    showSkeleton();

    try {
      const url = "/api/doc?path=" + encodeURIComponent(path);
      const res = await fetch(url, { signal: aborter.signal });
      if (!res.ok) {
        let msg = "文档不存在";
        try {
          const j = await res.json();
          if (j.error) msg = j.error;
        } catch {
          /* ignore */
        }
        throw new Error(msg);
      }
      const doc = await res.json();
      paint(doc, path);
      if (window.Sidebar) window.Sidebar.syncActive(path);
    } catch (err) {
      if (err.name === "AbortError") return;
      paintError(err.message || String(err));
    }
  }

  function paint(doc, path) {
    currentPath = path;
    content.innerHTML = doc.html;
    if (window.MarkdownUI) window.MarkdownUI.enhance(content);
    if (tocEl._obs) tocEl._obs.disconnect();
    renderToc(doc.toc);
    renderDocnav(doc);
    setCrumb(doc.title);
    document.title = doc.title + " · DocSite";
    content.style.animation = "none";
    void content.offsetWidth;
    content.style.animation = "";
    const scroller = document.querySelector(".content-wrap");
    if (scroller) scroller.scrollTop = 0;
  }

  function paintError(msg) {
    content.innerHTML =
      `<div style="padding:40px 0"><h1>未找到文档</h1>` +
      `<p style="color:var(--text-secondary)">${escapeHtml(msg)}</p>` +
      `<p><a href="#/">返回首页</a></p></div>`;
    tocEl.hidden = true;
    docnav.innerHTML = "";
  }

  /** 菜单点击时决定行为。 */
  function dispatch(item) {
    if (!item) return;
    switch (item.action) {
      case "link":
        if (item.url) window.open(item.url, "_blank", "noopener");
        return;
      case "callback":
        if (item.callback && window[item.callback]) {
          window[item.callback]();
        }
        return;
      case "doc":
        if (item.path) {
          const target = "#/" + item.path;
          if (location.hash !== target) location.hash = target;
          else load(item.path);
        }
        return;
      case "toggle":
      default:
        if (window.Sidebar) window.Sidebar.toggleByPath(item);
    }
  }

  function start() {
    if (started) return;
    started = true;

    window.addEventListener("hashchange", () => {
      const p = pathFromHash();
      // 非路由 hash（页内锚点）不重新加载文档，交给浏览器滚动
      if (!p) return;
      load(p);
    });

    const initial =
      pathFromHash() || (window.Site && window.Site.defaultDoc) || "index.md";
    load(initial);
  }

  window.Router = {
    load,
    dispatch,
    pathFromHash,
    start,
    currentPath: () => currentPath,
  };
})();
