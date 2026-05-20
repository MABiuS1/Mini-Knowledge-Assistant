import type { Citation } from "@/types/api";

export function CitationList({ citations }: { citations: Citation[] }) {
  if (citations.length === 0) {
    return null;
  }

  return (
    <div className="mt-3 space-y-2 border-t border-line pt-3">
      <p className="text-xs font-semibold uppercase text-muted">Citations</p>
      <div className="space-y-2">
        {citations.map((citation, index) => (
          <article
            key={citation.chunkId}
            className="rounded-md border border-line bg-white px-3 py-2"
          >
            <div className="flex flex-wrap items-center gap-2">
              <span className="rounded-full bg-brand px-2 py-0.5 text-xs font-semibold text-white">
                {index + 1}
              </span>
              <p className="min-w-0 flex-1 truncate text-xs font-semibold text-ink">
                {citation.fileName}
              </p>
              <span className="text-xs text-muted">
                Chunk {citation.chunkIndex + 1}
              </span>
            </div>
            <p className="mt-2 line-clamp-3 text-xs leading-5 text-muted">
              {citation.snippet}
            </p>
            <p className="mt-2 text-[11px] text-muted">
              Similarity {citation.similarity.toFixed(3)}
            </p>
          </article>
        ))}
      </div>
    </div>
  );
}
