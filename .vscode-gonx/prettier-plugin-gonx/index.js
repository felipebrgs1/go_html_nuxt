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

/**
 * Extract a named SFC block using index-based matching to avoid the lazy-regex
 * pitfall where `<template…>([\s\S]*?)</template>` stops at the first
 * `</template>` that Prettier may inject inside the HTML content.
 *
 * Strategy: find the opening tag, then scan forward counting nested opens to
 * locate the correct closing tag.
 */
function extractBlock(src, tag) {
  const openRe = new RegExp(`<${tag}(?:[^>]*)>`, "i");
  const openMatch = src.match(openRe);
  if (!openMatch) return null;

  const openStart = openMatch.index;
  const openEnd = openStart + openMatch[0].length;

  // Scan for matching close tag, skipping nested occurrences of the same tag
  const closeTag = `</${tag}>`;
  const nestedOpen = new RegExp(`<${tag}(?:[^>]*)>`, "gi");
  let depth = 1;
  let searchFrom = openEnd;
  let closeStart = -1;

  while (depth > 0) {
    const nextClose = src.indexOf(closeTag, searchFrom);
    if (nextClose === -1) return null; // malformed

    // Count any nested opens between searchFrom and nextClose
    nestedOpen.lastIndex = searchFrom;
    let m;
    while ((m = nestedOpen.exec(src)) !== null && m.index < nextClose) {
      depth++;
    }

    depth--;
    if (depth === 0) {
      closeStart = nextClose;
    } else {
      searchFrom = nextClose + closeTag.length;
    }
  }

  if (closeStart === -1) return null;

  return {
    open: openMatch[0],
    content: src.slice(openEnd, closeStart),
    close: closeTag,
    full: src.slice(openStart, closeStart + closeTag.length),
    openStart,
    openEnd,
    closeStart,
  };
}

/** Replace the content of a specific block in the source (index-aware). */
function replaceBlock(src, tag, newContent) {
  const block = extractBlock(src, tag);
  if (!block) return src;
  return (
    src.slice(0, block.openEnd) +
    newContent +
    src.slice(block.closeStart)
  );
}

/**
 * Remove the outermost wrapper div that we added before formatting.
 * We do this by stripping the first opening tag and the LAST closing </div>,
 * so nested </div> tags inside the content are preserved correctly.
 */
function unwrapDiv(formatted, id) {
  const openTag = `<div id="${id}">`;
  const closeTag = `</div>`;

  const openIdx = formatted.indexOf(openTag);
  if (openIdx === -1) return formatted;

  const contentStart = openIdx + openTag.length;
  const closeIdx = formatted.lastIndexOf(closeTag);
  if (closeIdx === -1 || closeIdx < contentStart) return formatted;

  return formatted.slice(contentStart, closeIdx);
}

/**
 * Escape {{ }} so prettier's HTML parser treats them as plain text
 * (otherwise it may mangle them thinking they are template expressions).
 * We use a unique sentinel that prettier won't touch.
 */
const OPEN_SENTINEL = "___GONX_OPEN___";
const CLOSE_SENTINEL = "___GONX_CLOSE___";

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

const WRAPPER_ID = "__gonx_root__";

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
          const wrappedInput = `<div id="${WRAPPER_ID}">${escaped}</div>`;
          const formatted = await prettier.format(wrappedInput, {
            ...options,
            parser: "html",
            plugins: [], // avoid recursion
            htmlWhitespaceSensitivity: "css",
            printWidth: options.printWidth || 100,
          });

          // Unwrap using last-index approach to preserve all inner </div> tags
          const inner = unwrapDiv(formatted, WRAPPER_ID);
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
