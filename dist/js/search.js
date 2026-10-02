/**
 * search.js — 前端全文搜索（调用 /api/search），带键盘导航。
 */
(function () {
  "use strict";

  const input = document.getElementById("searchInput");
  const box = document.getElementById("searchResults");
  if (!input || !box) return;

  let timer = null;
  let hits = [];
  let focused = -1;

  function esc(s) {
    return String(s).replace(/[&<>"']/g, (c) => ({
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#39;",
    }[c]));
  }

  function render(list, q) {
    hits = list;
    focused = -1;
    if (!list || !list.length) {
      box.innerHTML = `<div class="search-empty">未找到与 “${esc(q)}” 相关的文档</div>`;
      box.hidden = false;
      return;
    }
    box.innerHTML =
      list
        .map(
          (h, i) =>
            `<button class="search-hit" data-i="${i}" data-path="${esc(h.path)}">
              <span class="search-hit__title">${esc(h.title)}</span>
              <span class="search-hit__path">${esc(h.path)}</span>
              ${h.snippet ? `<span class="search-hit__snippet">${esc(h.snippet)}</span>` : ""}
            </button>`
        )
        .join("");
    box.hidden = false;

    box.querySelectorAll(".search-hit").forEach((btn) => {
      btn.addEventListener("click", () => go(btn.dataset.path));
    });
  }

  function go(path) {
    location.hash = "/" + path;
    hide();
    input.blur();
    if (window.Sidebar) window.Sidebar.closeMobileNav();
  }

  function hide() {
    box.hidden = true;
    box.innerHTML = "";
    hits = [];
    focused = -1;
  }

  async function run() {
    const q = input.value.trim();
    if (q.length < 2) {
      hide();
      return;
    }
    try {
      const results = await window.API.search(q);
      render(results, q);
    } catch {
      box.innerHTML = '<div class="search-empty">搜索索引载入失败</div>';
      box.hidden = false;
    }
  }

  input.addEventListener("input", () => {
    clearTimeout(timer);
    timer = setTimeout(run, 200);
  });

  input.addEventListener("focus", () => {
    if (input.value.trim().length >= 2 && hits.length) box.hidden = false;
  });

  document.addEventListener("click", (e) => {
    if (!box.contains(e.target) && e.target !== input) hide();
  });

  input.addEventListener("keydown", (e) => {
    const items = Array.from(box.querySelectorAll(".search-hit"));
    if (e.key === "ArrowDown") {
      e.preventDefault();
      focused = Math.min(focused + 1, items.length - 1);
      mark(items);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      focused = Math.max(focused - 1, 0);
      mark(items);
    } else if (e.key === "Enter") {
      e.preventDefault();
      const i = focused >= 0 ? focused : 0;
      if (hits[i]) go(hits[i].path);
    } else if (e.key === "Escape") {
      hide();
      input.blur();
    }
  });

  function mark(items) {
    items.forEach((el, i) => {
      el.classList.toggle("is-focused", i === focused);
      if (i === focused) el.scrollIntoView({ block: "nearest" });
    });
  }

  // 快捷键： "/" 聚焦搜索框
  document.addEventListener("keydown", (e) => {
    if (
      e.key === "/" &&
      document.activeElement !== input &&
      !/INPUT|TEXTAREA/.test(document.activeElement.tagName)
    ) {
      e.preventDefault();
      input.focus();
    }
  });
})();
