/**
 * prettier-plugin-gonx
 *
 * Formats .gonx Single File Components:
 *   <template>  → HTML (via prettier's built-in html parser)
 *   <style>     → CSS  (via prettier's built-in css parser)
 *   <script>    → left untouched by prettier (handled by `framework fmt` / gofmt)
 *
 * Approach: extract each block, format it independently, stitch back together.
 * Gonx interpolation markers {{ }} are preserved as-is inside the HTML.
 */

"use strict";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Extract a named SFC block and its surrounding tags from the source. */
function extractBlock(src, tag) {
  const re = new RegExp(
    `(<${tag}(?:[^>]*)>)([\\s\\S]*?)(<\\/${tag}>)`,
    "i"
  );
  const m = src.match(re);
  if (!m) return null;
  return { open: m[1], content: m[2], close: m[3], full: m[0] };
}

/** Replace the content of a specific block in the source. */
function replaceBlock(src, tag, newContent) {
  const re = new RegExp(
    `(<${tag}(?:[^>]*)>)([\\s\\S]*?)(<\\/${tag}>)`,
    "i"
  );
  return src.replace(re, (_match, open, _old, close) => {
    return open + newContent + close;
  });
}

/**
 * Escape {{ }} so prettier's HTML parser treats them as plain text
 * (otherwise it may mangle them thinking they are template expressions).
 * We use a unique sentinel that prettier won't touch.
 */
const OPEN_SENTINEL = "\u200B\uFEFF\u200B"; // zero-width chars, invisible
const CLOSE_SENTINEL = "\u200C\uFEFF\u200C";

function escapeInterpolations(html) {
  return html
    .replace(/\{\{/g, OPEN_SENTINEL)
    .replace(/\}\}/g, CLOSE_SENTINEL);
}

function unescapeInterpolations(html) {
  return html
    .replace(new RegExp(OPEN_SENTINEL, "g"), "{{")
    .replace(new RegExp(CLOSE_SENTINEL, "g"), "}}");
}

// ---------------------------------------------------------------------------
// Prettier Plugin API
// ---------------------------------------------------------------------------

const languages = [
  {
    name: "Gonx",
    parsers: ["gonx"],
    extensions: [".gonx"],
    vscodeLanguageIds: ["gonx"],
  },
];

const parsers = {
  gonx: {
    astFormat: "gonx-ast",

    async parse(text, options) {
      return { type: "gonx-root", source: text, options };
    },

    locStart: () => 0,
    locEnd: (node) => node.source.length,
  },
};

const printers = {
  "gonx-ast": {
    async print(path, options, _print) {
      const node = path.getValue();
      const src = node.source;

      // We need access to the prettier instance to format sub-blocks.
      // prettier v3 passes it via options.plugins; prettier v2 is global.
      const prettier =
        options.__plugin_prettier_instance ||
        (typeof require !== "undefined" ? require("prettier") : null);

      if (!prettier) {
        // Fallback: return source unchanged
        return src;
      }

      let result = src;

      // --- Format <template> as HTML ---
      const tmpl = extractBlock(src, "template");
      if (tmpl) {
        try {
          const escaped = escapeInterpolations(tmpl.content);
          // Wrap in a dummy root so prettier can parse as a fragment
          const wrappedInput = `<div id="__gonx_root__">${escaped}</div>`;
          const formatted = await prettier.format(wrappedInput, {
            ...options,
            parser: "html",
            plugins: [], // avoid recursion
            htmlWhitespaceSensitivity: "css",
            printWidth: options.printWidth || 100,
          });
          // Unwrap the dummy root
          const innerMatch = formatted.match(
            /<div id="__gonx_root__">([\s\S]*?)<\/div>/
          );
          const inner = innerMatch
            ? innerMatch[1]
            : formatted
                .replace('<div id="__gonx_root__">', "")
                .replace("</div>", "");
          const unescaped = unescapeInterpolations(inner);
          result = replaceBlock(result, "template", "\n" + unescaped);
        } catch (_e) {
          // If formatting fails (e.g., invalid HTML), keep original
        }
      }

      // --- Format <style> as CSS ---
      const style = extractBlock(result, "style");
      if (style && style.content.trim()) {
        try {
          const formattedCss = await prettier.format(style.content, {
            ...options,
            parser: "css",
            plugins: [],
          });
          result = replaceBlock(result, "style", "\n" + formattedCss);
        } catch (_e) {
          // keep original
        }
      }

      // <script> is intentionally not touched — use `framework fmt` (gofmt)

      return result;
    },
  },
};

module.exports = { languages, parsers, printers };
