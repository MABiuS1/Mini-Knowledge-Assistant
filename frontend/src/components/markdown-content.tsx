import type { ReactNode } from "react";

type MarkdownBlock =
  | { type: "code"; content: string }
  | { type: "heading"; level: 2 | 3; content: string }
  | { type: "list"; items: string[]; ordered: boolean }
  | { type: "paragraph"; content: string };

export function MarkdownContent({ content }: { content: string }) {
  const blocks = parseMarkdown(content);

  return (
    <div className="space-y-3 leading-6">
      {blocks.map((block, index) => (
        <MarkdownBlockView block={block} key={`${block.type}-${index}`} />
      ))}
    </div>
  );
}

function MarkdownBlockView({ block }: { block: MarkdownBlock }) {
  switch (block.type) {
    case "code":
      return (
        <pre className="overflow-x-auto rounded-md border border-line bg-black/40 px-3 py-2 text-xs text-white">
          <code>{block.content}</code>
        </pre>
      );
    case "heading":
      if (block.level === 2) {
        return (
          <h2 className="text-base font-semibold text-ink">
            {renderInlineMarkdown(block.content)}
          </h2>
        );
      }
      return (
        <h3 className="text-sm font-semibold text-ink">
          {renderInlineMarkdown(block.content)}
        </h3>
      );
    case "list": {
      const ListTag = block.ordered ? "ol" : "ul";
      return (
        <ListTag
          className={
            block.ordered
              ? "list-decimal space-y-1 pl-5"
              : "list-disc space-y-1 pl-5"
          }
        >
          {block.items.map((item, index) => (
            <li key={`${item}-${index}`}>{renderInlineMarkdown(item)}</li>
          ))}
        </ListTag>
      );
    }
    case "paragraph":
      return <p>{renderInlineMarkdown(block.content)}</p>;
  }
}

function parseMarkdown(content: string): MarkdownBlock[] {
  const lines = content.replace(/\r\n/g, "\n").split("\n");
  const blocks: MarkdownBlock[] = [];
  let index = 0;

  while (index < lines.length) {
    const line = lines[index];
    const trimmed = line.trim();

    if (trimmed === "") {
      index += 1;
      continue;
    }

    if (trimmed.startsWith("```")) {
      const codeLines: string[] = [];
      index += 1;
      while (index < lines.length && !lines[index].trim().startsWith("```")) {
        codeLines.push(lines[index]);
        index += 1;
      }
      blocks.push({ type: "code", content: codeLines.join("\n") });
      index += 1;
      continue;
    }

    const heading = trimmed.match(/^(#{1,3})\s+(.+)$/);
    if (heading) {
      blocks.push({
        type: "heading",
        level: heading[1].length <= 2 ? 2 : 3,
        content: heading[2],
      });
      index += 1;
      continue;
    }

    const unorderedItems = collectListItems(lines, index, "unordered");
    if (unorderedItems.items.length > 0) {
      blocks.push({
        type: "list",
        ordered: false,
        items: unorderedItems.items,
      });
      index = unorderedItems.nextIndex;
      continue;
    }

    const orderedItems = collectListItems(lines, index, "ordered");
    if (orderedItems.items.length > 0) {
      blocks.push({
        type: "list",
        ordered: true,
        items: orderedItems.items,
      });
      index = orderedItems.nextIndex;
      continue;
    }

    const paragraphLines: string[] = [];
    while (index < lines.length && lines[index].trim() !== "") {
      const paragraphLine = lines[index].trim();
      if (
        paragraphLine.startsWith("```") ||
        paragraphLine.match(/^(#{1,3})\s+(.+)$/) ||
        paragraphLine.match(/^[-*]\s+(.+)$/) ||
        paragraphLine.match(/^\d+\.\s+(.+)$/)
      ) {
        break;
      }

      paragraphLines.push(paragraphLine);
      index += 1;
    }

    if (paragraphLines.length > 0) {
      blocks.push({ type: "paragraph", content: paragraphLines.join(" ") });
      continue;
    }

    index += 1;
  }

  return blocks;
}

function collectListItems(
  lines: string[],
  startIndex: number,
  type: "ordered" | "unordered",
): { items: string[]; nextIndex: number } {
  const matcher =
    type === "ordered" ? /^\d+\.\s+(.+)$/ : /^[-*]\s+(.+)$/;
  const items: string[] = [];
  let index = startIndex;

  while (index < lines.length) {
    const match = lines[index].trim().match(matcher);
    if (!match) {
      break;
    }

    items.push(match[1]);
    index += 1;
  }

  return { items, nextIndex: index };
}

function renderInlineMarkdown(content: string): ReactNode[] {
  const parts: ReactNode[] = [];
  const pattern = /(`[^`]+`|\*\*[^*]+\*\*)/g;
  let lastIndex = 0;
  let match: RegExpExecArray | null;

  while ((match = pattern.exec(content)) !== null) {
    if (match.index > lastIndex) {
      parts.push(content.slice(lastIndex, match.index));
    }

    const token = match[0];
    if (token.startsWith("`")) {
      parts.push(
        <code
          className="rounded border border-line bg-white/20 px-1.5 py-0.5 text-xs text-ink"
          key={`${token}-${match.index}`}
        >
          {token.slice(1, -1)}
        </code>,
      );
    } else {
      parts.push(
        <strong className="font-semibold" key={`${token}-${match.index}`}>
          {token.slice(2, -2)}
        </strong>,
      );
    }

    lastIndex = pattern.lastIndex;
  }

  if (lastIndex < content.length) {
    parts.push(content.slice(lastIndex));
  }

  return parts;
}
