/**
 * api.js — 统一数据访问层。
 *
 * 服务端（Go 单二进制）与静态托管（GitHub Pages）都提供完全相同的
 * `api/*.json` 路径，因此同一份前端在两种模式下行为一致：
 *   - 服务端：由 internal/handler 实时生成
 *   - 静态：由 cmd/build 预生成为文件
 *
 * 路径一律使用相对地址，以便部署在子路径（如 /docsite/）下也能工作。
 */
(function () {
  "use strict";

  const BASE = "api/";
  const cache = { config: null, menu: null, docs: null, index: null };

  async function getJSON(name) {
    const res = await fetch(BASE + name, { cache: "no-cache" });
    if (!res.ok) throw new Error(name + " -> HTTP " + res.status);
    return res.json();
  }

  async function config() {
    if (!cache.config) cache.config = await getJSON("config.json");
    return cache.config;
  }

  async function menu() {
    if (!cache.menu) cache.menu = await getJSON("menu.json");
    return cache.menu;
  }

  /** 返回 path -> 文档 的 Map（一次性载入全部文档）。 */
  async function docs() {
    if (!cache.docs) {
      const bundle = await getJSON("docs.json");
      const map = new Map();
      (bundle.docs || []).forEach((d) => map.set(d.path, d));
      cache.docs = map;
    }
    return cache.docs;
  }

  async function doc(path) {
    const map = await docs();
    if (map.has(path)) return map.get(path);
    // 兼容省略 .md 后缀的写法
    const withExt = path.endsWith(".md") ? path : path + ".md";
    return map.get(withExt) || null;
  }

  async function searchIndex() {
    if (!cache.index) {
      const idx = await getJSON("search-index.json");
      cache.index = idx.items || [];
    }
    return cache.index;
  }

  /**
   * 客户端全文搜索，匹配规则与服务端 service.Search 保持一致
   * （不区分大小写的子串匹配，最多 20 条，按 path 排序）。
   */
  async function search(q) {
    q = String(q || "").trim().toLowerCase();
    if (q.length < 2) return [];

    const items = await searchIndex();
    const hits = [];
    for (const it of items) {
      const text = it.text || "";
      const i = text.toLowerCase().indexOf(q);
      if (i < 0) continue;
      const start = Math.max(0, i - 60);
      const end = Math.min(text.length, i + q.length + 80);
      hits.push({
        path: it.path,
        title: it.title,
        snippet: text.slice(start, end).replace(/\s+/g, " ").trim(),
      });
    }
    hits.sort((a, b) => (a.path < b.path ? -1 : a.path > b.path ? 1 : 0));
    return hits.slice(0, 20);
  }

  window.API = { config, menu, docs, doc, searchIndex, search };
})();
