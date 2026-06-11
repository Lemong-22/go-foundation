import type { BlockView } from "@/lib/course-api/server";

export type LessonReaderBlock = BlockView & {
  kind: "practice" | "quiz" | "text" | "video" | "unsupported";
  label: string;
  number: string;
  positionLabel: string;
};

export type MarkdownBlock =
  | {
      level: 2 | 3 | 4;
      text: string;
      type: "heading";
    }
  | {
      text: string;
      type: "paragraph";
    }
  | {
      items: string[];
      type: "list";
    }
  | {
      code: string;
      language: string;
      type: "code";
    };

export type MarkdownInline =
  | {
      text: string;
      type: "text";
    }
  | {
      text: string;
      type: "code";
    }
  | {
      href: string;
      text: string;
      type: "link";
    };

export type VideoEmbed =
  | {
      kind: "iframe";
      providerLabel: string;
      src: string;
    }
  | {
      kind: "video";
      providerLabel: string;
      src: string;
    }
  | {
      href: string;
      kind: "link";
      providerLabel: string;
    }
  | {
      kind: "unavailable";
      providerLabel: string;
    };

export type DummyActivity =
  | {
      badge: "Sample data";
      description: string;
      kind: "quiz";
      options: string[];
      prompt: string;
      ref: string;
      title: string;
    }
  | {
      badge: "Sample data";
      checkpoints: string[];
      kind: "practice";
      prompt: string;
      ref: string;
      starterCode: string;
      title: string;
    };

const inlinePattern = /(`([^`]+)`|\[([^\]]+)\]\(((?:[^()]|\([^)]*\))+)\))/g;

export function buildLessonBlocks(blocks: BlockView[]): LessonReaderBlock[] {
  return blocks.map((block, index) => {
    const kind = normalizeBlockKind(block.Kind);

    return {
      ...block,
      kind,
      label: labelForBlock(block, kind),
      number: formatBlockNumber(index + 1),
      positionLabel: `Position ${block.Position}`,
    };
  });
}

export function formatBlockNumber(value: number) {
  return value.toString().padStart(2, "0");
}

export function buildVideoEmbed(block: Pick<BlockView, "VideoLocator" | "VideoProvider">): VideoEmbed {
  const provider = block.VideoProvider.trim().toLowerCase();
  const locator = block.VideoLocator.trim();

  if (!locator) {
    return {
      kind: "unavailable",
      providerLabel: providerLabel(provider),
    };
  }

  if (provider === "youtube") {
    return {
      kind: "iframe",
      providerLabel: "YouTube",
      src: `https://www.youtube-nocookie.com/embed/${encodeURIComponent(locator)}`,
    };
  }

  if (provider === "mux") {
    return {
      kind: "iframe",
      providerLabel: "Mux",
      src: `https://player.mux.com/${encodeURIComponent(locator)}`,
    };
  }

  if (provider === "url" && isSafeHttpUrl(locator)) {
    return {
      kind: "video",
      providerLabel: "URL",
      src: locator,
    };
  }

  if (isSafeHttpUrl(locator)) {
    return {
      href: locator,
      kind: "link",
      providerLabel: providerLabel(provider),
    };
  }

  return {
    kind: "unavailable",
    providerLabel: providerLabel(provider),
  };
}

export function buildDummyActivity(
  block: Pick<BlockView, "Kind" | "PracticeRef" | "QuizRef">,
): DummyActivity {
  const kind = block.Kind.trim().toLowerCase();

  if (kind === "practice") {
    return {
      badge: "Sample data",
      checkpoints: [
        "Read the prompt and identify the inputs.",
        "Sketch the steps in plain language.",
        "Try a small example before writing code.",
      ],
      kind: "practice",
      prompt:
        "Use the concept from this lesson to transform a small input and explain each step.",
      ref: block.PracticeRef.trim(),
      starterCode:
        "function practice(input) {\n  // Write a small experiment here.\n  return input;\n}",
      title: "Sample practice prompt",
    };
  }

  return {
    badge: "Sample data",
    description:
      "Use this placeholder to preview the learner interaction shape.",
    kind: "quiz",
    options: [
      "Identify the key term in the lesson.",
      "Match the concept to a short example.",
      "Explain what changes between two snippets.",
    ],
    prompt: "Which action best checks your understanding of this lesson?",
    ref: block.QuizRef.trim(),
    title: "Sample quiz prompt",
  };
}

export function parseMarkdown(markdown: string): MarkdownBlock[] {
  const lines = markdown.replace(/\r\n?/g, "\n").split("\n");
  const blocks: MarkdownBlock[] = [];
  let index = 0;

  while (index < lines.length) {
    const line = lines[index] ?? "";
    const trimmed = line.trim();

    if (!trimmed) {
      index += 1;
      continue;
    }

    if (trimmed.startsWith("```")) {
      const language = trimmed.slice(3).trim();
      const code: string[] = [];
      index += 1;

      while (index < lines.length && !(lines[index] ?? "").trim().startsWith("```")) {
        code.push(lines[index] ?? "");
        index += 1;
      }

      if (index < lines.length) {
        index += 1;
      }

      blocks.push({
        code: code.join("\n"),
        language,
        type: "code",
      });
      continue;
    }

    const heading = /^(#{1,3})\s+(.+)$/.exec(trimmed);
    if (heading) {
      blocks.push({
        level: (heading[1].length + 1) as 2 | 3 | 4,
        text: heading[2].trim(),
        type: "heading",
      });
      index += 1;
      continue;
    }

    if (/^[-*]\s+/.test(trimmed)) {
      const items: string[] = [];

      while (index < lines.length) {
        const item = /^[-*]\s+(.+)$/.exec((lines[index] ?? "").trim());
        if (!item) {
          break;
        }

        items.push(item[1].trim());
        index += 1;
      }

      blocks.push({
        items,
        type: "list",
      });
      continue;
    }

    const paragraph: string[] = [];
    while (index < lines.length) {
      const next = lines[index] ?? "";
      const nextTrimmed = next.trim();
      if (
        !nextTrimmed ||
        nextTrimmed.startsWith("```") ||
        /^(#{1,3})\s+/.test(nextTrimmed) ||
        /^[-*]\s+/.test(nextTrimmed)
      ) {
        break;
      }

      paragraph.push(nextTrimmed);
      index += 1;
    }

    blocks.push({
      text: paragraph.join(" "),
      type: "paragraph",
    });
  }

  return blocks;
}

export function parseInlineMarkdown(text: string): MarkdownInline[] {
  const nodes: MarkdownInline[] = [];
  let lastIndex = 0;

  for (const match of text.matchAll(inlinePattern)) {
    const start = match.index ?? 0;
    if (start > lastIndex) {
      nodes.push({
        text: text.slice(lastIndex, start),
        type: "text",
      });
    }

    if (match[2]) {
      nodes.push({
        text: match[2],
        type: "code",
      });
    } else {
      const linkText = match[3] ?? "";
      const href = match[4] ?? "";
      if (isSafeHttpUrl(href)) {
        nodes.push({
          href,
          text: linkText,
          type: "link",
        });
      } else {
        nodes.push({
          text: linkText,
          type: "text",
        });
      }
    }

    lastIndex = start + match[0].length;
  }

  if (lastIndex < text.length) {
    nodes.push({
      text: text.slice(lastIndex),
      type: "text",
    });
  }

  return nodes;
}

function normalizeBlockKind(kind: string): LessonReaderBlock["kind"] {
  const normalized = kind.trim().toLowerCase();
  if (
    normalized === "practice" ||
    normalized === "quiz" ||
    normalized === "text" ||
    normalized === "video"
  ) {
    return normalized;
  }

  return "unsupported";
}

function labelForBlock(
  block: BlockView,
  kind: LessonReaderBlock["kind"],
) {
  if (kind === "text") {
    return "Text";
  }

  if (kind === "video") {
    return providerLabel(block.VideoProvider);
  }

  if (kind === "quiz") {
    return "Sample quiz";
  }

  if (kind === "practice") {
    return "Sample practice";
  }

  return block.Kind || "Activity";
}

function providerLabel(provider: string) {
  const normalized = provider.trim().toLowerCase();

  switch (normalized) {
    case "youtube":
      return "YouTube";
    case "mux":
      return "Mux";
    case "url":
      return "URL";
    default:
      return normalized || "Video";
  }
}

function isSafeHttpUrl(value: string) {
  try {
    const url = new URL(value);
    return url.protocol === "https:" || url.protocol === "http:";
  } catch {
    return false;
  }
}
