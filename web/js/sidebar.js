/**
 * sidebar.js — 侧边栏：多级菜单渲染、折叠、选中高亮、移动端抽屉。
 * 依赖 /api/menu 的返回结构（见 model.MenuItem）。
 */
(function () {
  "use strict";

  const app = document.getElementById("app");
  const nav = document.getElementById("sidebarNav");
  const btnCollapse = document.getElementById("btnCollapse");
  const btnMobile = document.getElementById("btnMobileMenu");
  const overlay = document.getElementById("overlay");

  const KEY_COLLAPSE = "docsite:collapsed";

  const ICONS = {
    home: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>',
    book: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>',
    info: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>',
    download: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>',
    settings: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>',
    code: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>',
    heart: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/></svg>',
    doc: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>',
    chevron: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>',
  };

  function icon(name) {
    return (name && ICONS[name]) || ICONS.doc;
  }

  /** 递归渲染菜单树。 */
  function build(items) {
    const frag = document.createDocumentFragment();
    items.forEach((it) => {
      if (it.hidden) return;
      frag.appendChild(renderNode(it));
    });
    nav.appendChild(frag);
  }

  function renderNode(item) {
    const ul = document.createElement("div");
    ul.className = "menu";
    ul.dataset.id = item.id;

    const hasChildren = item.children && item.children.length > 0;
    const isDoc = item.action === "doc" && item.path;

    const row = document.createElement(item.action === "link" ? "a" : "button");
    row.className = "menu__row";
    if (item.action === "link" && item.url) {
      row.href = item.url;
      row.target = "_blank";
      row.rel = "noopener";
      row.classList.add("is-external");
    } else {
      row.type = "button";
    }
    row.dataset.path = item.path || "";
    row.dataset.action = item.action || "";
    row.setAttribute("role", "menuitem");

    const iconEl = document.createElement("span");
    iconEl.className = "menu__icon";
    iconEl.innerHTML = icon(item.icon);

    const titleEl = document.createElement("span");
    titleEl.className = "menu__title";
    titleEl.textContent = item.title;

    row.appendChild(iconEl);
    row.appendChild(titleEl);

    if (hasChildren) {
      const chev = document.createElement("span");
      chev.className = "menu__chevron";
      chev.innerHTML = ICONS.chevron;
      row.appendChild(chev);
    }

    ul.appendChild(row);

    if (hasChildren) {
      const kids = document.createElement("div");
      kids.className = "menu__children";
      const inner = document.createElement("div");
      inner.className = "menu__children-inner";
      item.children.forEach((c) => inner.appendChild(renderNode(c)));
      kids.appendChild(inner);
      ul.appendChild(kids);
      if (item.expanded) ul.classList.add("is-open");
      row.setAttribute("aria-expanded", ul.classList.contains("is-open") ? "true" : "false");

      row.addEventListener("click", (e) => {
        e.stopPropagation();
        // 父菜单：切换展开
        ul.classList.toggle("is-open");
        row.setAttribute("aria-expanded", ul.classList.contains("is-open") ? "true" : "false");
        rememberExpand(ul);
      });
    } else if (isDoc) {
      row.addEventListener("click", () => {
        if (window.Router) window.Router.dispatch(item);
        closeMobileNav();
      });
    } else if (item.action === "callback" || item.action === "link") {
      row.addEventListener("click", (e) => {
        if (item.action !== "link") {
          e.preventDefault();
          if (window.Router) window.Router.dispatch(item);
        }
        closeMobileNav();
      });
    }

    return ul;
  }

  /* ---------- 展开态记忆 ---------- */
  const KEY_EXPAND = "docsite:expanded";

  function rememberExpand(ul) {
    try {
      const map = JSON.parse(localStorage.getItem(KEY_EXPAND) || "{}");
      map[ul.dataset.id] = ul.classList.contains("is-open");
      localStorage.setItem(KEY_EXPAND, JSON.stringify(map));
    } catch {
      /* ignore */
    }
  }

  function restoreExpand() {
    try {
      const map = JSON.parse(localStorage.getItem(KEY_EXPAND) || "{}");
      nav.querySelectorAll(".menu").forEach((el) => {
        if (el.dataset.id in map) {
          el.classList.toggle("is-open", map[el.dataset.id]);
        }
      });
    } catch {
      /* ignore */
    }
  }

  /* ---------- 选中高亮 ---------- */
  function syncActive(path) {
    nav.querySelectorAll(".menu__row").forEach((r) => {
      r.classList.remove("is-active");
      r.removeAttribute("aria-current");
    });
    if (!path) return;
    const target = nav.querySelector(
      `.menu__row[data-path="${CSS.escape(path)}"]`
    );
    if (!target) return;
    target.classList.add("is-active");
    target.setAttribute("aria-current", "page");
    // 自动展开所有父级
    let p = target.parentElement;
    while (p && p !== nav) {
      if (p.classList && p.classList.contains("menu")) {
        p.classList.add("is-open");
        const row = p.querySelector(":scope > .menu__row");
        if (row) row.setAttribute("aria-expanded", "true");
      }
      p = p.parentElement;
    }
  }

  /* ---------- 折叠 / 抽屉 ---------- */
  function setCollapsed(v) {
    app.classList.toggle("is-collapsed", v);
    try {
      localStorage.setItem(KEY_COLLAPSE, v ? "1" : "0");
    } catch {
      /* ignore */
    }
  }

  function toggleCollapsed() {
    setCollapsed(!app.classList.contains("is-collapsed"));
  }

  function openMobileNav() {
    app.classList.add("nav-open");
    btnMobile.setAttribute("aria-expanded", "true");
    overlay.hidden = false;
  }

  function closeMobileNav() {
    app.classList.remove("nav-open");
    btnMobile.setAttribute("aria-expanded", "false");
    overlay.hidden = true;
  }

  function init() {
    // 折叠状态
    let collapsed = false;
    try {
      collapsed = localStorage.getItem(KEY_COLLAPSE) === "1";
    } catch {
      /* ignore */
    }
    app.classList.toggle("is-collapsed", collapsed);

    btnCollapse.addEventListener("click", toggleCollapsed);
    btnMobile.addEventListener("click", openMobileNav);
    overlay.addEventListener("click", closeMobileNav);

    document.addEventListener("keydown", (e) => {
      if (e.key === "Escape" && app.classList.contains("nav-open")) {
        closeMobileNav();
      }
    });
  }

  /* ---------- 启动 ---------- */
  (async function boot() {
    try {
      const res = await fetch("/api/menu");
      const data = await res.json();
      build(data.items || []);
      restoreExpand();
      // 初始选中
      const p =
        (window.Router && window.Router.pathFromHash()) ||
        (window.Site && window.Site.defaultDoc) ||
        "index.md";
      syncActive(p);
    } catch (err) {
      nav.innerHTML =
        '<div style="padding:16px;color:var(--text-tertiary);font-size:13px">菜单加载失败</div>';
    }
    init();
  })();

  window.Sidebar = {
    syncActive,
    toggleByPath(item) {
      const el = nav.querySelector(`.menu[data-id="${CSS.escape(item.id)}"]`);
      if (!el) return;
      el.classList.toggle("is-open");
      const row = el.querySelector(":scope > .menu__row");
      if (row) {
        row.setAttribute(
          "aria-expanded",
          el.classList.contains("is-open") ? "true" : "false"
        );
      }
    },
    closeMobileNav,
  };

  // 暴露给搜索模块
  window.Sidebar._nav = nav;
})();
