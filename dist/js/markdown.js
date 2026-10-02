/**
 * markdown.js — 服务端已渲染 HTML；这里做客户端增强：
 *  - 为标题补 id（与服务端 TOC slug 保持一致，确保锚点可点）
 *  - 为服务端生成的复制按钮绑定复制行为
 */
(function () {
  "use strict";

  function slugify(s) {
    s = String(s).trim().toLowerCase();
    s = s.replace(/[^\p{L}\p{N}-]+/gu, "-");
    s = s.replace(/^-+|-+$/g, "");
    return s || "section";
  }

  function anchorHeadings(root) {
    const seen = {};
    root.querySelectorAll("h1, h2, h3, h4, h5, h6").forEach((h) => {
      // 服务端渲染时已写入唯一 id，直接沿用，避免与 TOC 锚点不一致
      if (h.id) {
        seen[h.id] = 0;
        return;
      }
      let id = slugify(h.textContent);
      if (seen[id]) {
        seen[id] += 1;
        id = id + "-" + seen[id];
      } else {
        seen[id] = 0;
      }
      h.id = id;
    });
  }

  function bindCopyButtons(root) {
    root.querySelectorAll(".codeblock__copy").forEach((btn) => {
      if (btn._bound) return;
      btn._bound = true;
      btn.addEventListener("click", async () => {
        const code = btn.closest(".codeblock")?.querySelector("code");
        if (!code) return;
        const text = code.innerText;
        try {
          await navigator.clipboard.writeText(text);
        } catch {
          const ta = document.createElement("textarea");
          ta.value = text;
          document.body.appendChild(ta);
          ta.select();
          document.execCommand("copy");
          ta.remove();
        }
        const label = btn.querySelector("span:last-child") || btn;
        const old = label.textContent;
        label.textContent = "已复制";
        btn.classList.add("is-copied");
        setTimeout(() => {
          label.textContent = old;
          btn.classList.remove("is-copied");
        }, 1500);
      });
    });
  }

  function enhance(root) {
    anchorHeadings(root);
    bindCopyButtons(root);
  }

  window.MarkdownUI = { enhance: enhance, slugify: slugify };
})();
