"use client";

import type { FormEvent } from "react";
import { useEffect, useMemo, useState } from "react";
import { AppShell } from "@/components/app-shell";
import { AuthGuard } from "@/components/auth-guard";
import { listDocuments, uploadDocument } from "@/lib/documents-api";
import type { Document } from "@/types/api";

const allowedExtensions = new Set(["pdf", "txt"]);
const maxUploadBytes = readMaxUploadBytes();

type UploadState = "idle" | "uploading" | "success" | "error";

export default function UploadPage() {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploadState, setUploadState] = useState<UploadState>("idle");
  const [listState, setListState] = useState<"loading" | "ready" | "error">(
    "loading",
  );
  const [message, setMessage] = useState("");

  const selectedFileError = useMemo(() => {
    if (!selectedFile) {
      return "";
    }

    return validateFile(selectedFile);
  }, [selectedFile]);

  async function refreshDocuments() {
    setListState("loading");
    try {
      const response = await listDocuments();
      setDocuments(response.documents);
      setListState("ready");
    } catch (error) {
      setMessage(
        error instanceof Error ? error.message : "Unable to load documents",
      );
      setListState("error");
    }
  }

  useEffect(() => {
    void refreshDocuments();
  }, []);

  async function handleUpload(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setMessage("");

    if (!selectedFile) {
      setUploadState("error");
      setMessage("Select a PDF or TXT file.");
      return;
    }

    const validationError = validateFile(selectedFile);
    if (validationError) {
      setUploadState("error");
      setMessage(validationError);
      return;
    }

    setUploadState("uploading");
    try {
      const response = await uploadDocument(selectedFile);
      setDocuments((current) => [response.document, ...current]);
      setSelectedFile(null);
      setUploadState("success");
      setMessage(`${response.document.originalName} uploaded.`);
    } catch (error) {
      setUploadState("error");
      setMessage(error instanceof Error ? error.message : "Upload failed.");
    }
  }

  return (
    <AuthGuard>
      <AppShell>
        <div className="grid gap-4 xl:grid-cols-[1fr_280px] xl:gap-5">
          <div className="space-y-4 xl:space-y-5">
            <div>
              <h1 className="text-2xl font-semibold text-ink sm:text-3xl">
                Document Management
              </h1>
              <p className="mt-2 text-sm text-muted">
                Upload, index, and manage document context for Knowledge AI.
              </p>
            </div>

            <section className="glass-panel rounded-lg p-4 sm:p-5">
              <form onSubmit={handleUpload} className="space-y-4">
                <label className="drop-zone flex min-h-[160px] cursor-pointer flex-col items-center justify-center rounded-lg px-4 py-6 text-center sm:min-h-[190px] sm:px-6 sm:py-8">
                  <input
                    type="file"
                    accept=".pdf,.txt,application/pdf,text/plain"
                    disabled={uploadState === "uploading"}
                    onChange={(event) => {
                      setMessage("");
                      setUploadState("idle");
                      setSelectedFile(event.target.files?.[0] ?? null);
                      event.target.value = "";
                    }}
                    className="sr-only"
                  />
                  <span className="text-4xl leading-none text-purple-300 sm:text-5xl">
                    ⇧
                  </span>
                  <span className="mt-3 text-sm font-medium text-ink sm:text-base">
                    Drop upload files here
                  </span>
                  <span className="glow-button mt-4 rounded-md px-4 py-2 text-sm font-semibold">
                    Browse Local Files
                  </span>
                  <span className="mt-3 text-xs text-muted">
                    PDF and TXT, up to {formatBytes(maxUploadBytes)}
                  </span>
                </label>

                {selectedFile ? (
                  <div className="glass-soft rounded-lg p-4">
                    <div className="flex items-center gap-4">
                      <div className="file-glass-icon flex h-14 w-14 shrink-0 items-center justify-center rounded-lg text-[11px] font-semibold uppercase text-ink">
                        {selectedFile.name.split(".").pop() ?? "file"}
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-semibold text-ink">
                          {selectedFile.name}
                        </p>
                        <p className="mt-1 text-xs text-muted">
                          {formatBytes(selectedFile.size)}
                        </p>
                      </div>
                    </div>
                    {uploadState === "uploading" ? (
                      <div className="mt-4">
                        <div className="flex items-center justify-between text-xs font-medium text-muted">
                          <span>Uploading...</span>
                          <span>Processing</span>
                        </div>
                        <div className="glass-progress mt-2">
                          <span className="w-3/4" />
                        </div>
                      </div>
                    ) : null}
                  </div>
                ) : null}

                {selectedFileError ? (
                  <p className="text-sm font-medium text-red-300">
                    {selectedFileError}
                  </p>
                ) : null}

                <button
                  type="submit"
                  disabled={
                    uploadState === "uploading" ||
                    !selectedFile ||
                    Boolean(selectedFileError)
                  }
                  className="glow-button inline-flex w-full items-center justify-center rounded-md px-4 py-2.5 text-sm font-semibold disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {uploadState === "uploading" ? "Uploading..." : "Upload"}
                </button>
              </form>

              {message ? (
                <p
                  className={
                    uploadState === "error" || listState === "error"
                      ? "mt-4 rounded-md border border-red-400/30 bg-red-500/10 px-3 py-2 text-sm text-red-200"
                      : "mt-4 rounded-md border border-emerald-300/30 bg-emerald-400/10 px-3 py-2 text-sm text-emerald-100"
                  }
                >
                  {message}
                </p>
              ) : null}
            </section>

            <section className="glass-panel rounded-lg">
              <div className="flex items-center justify-between border-b border-line px-6 py-4">
                <h2 className="text-xl font-semibold text-ink">
                  Recent Documents
                </h2>
                <button
                  type="button"
                  onClick={() => void refreshDocuments()}
                  disabled={listState === "loading"}
                  className="ghost-button rounded-md px-3 py-2 text-sm font-medium disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {listState === "loading" ? "Loading..." : "Filter"}
                </button>
              </div>

              <DocumentTable documents={documents} listState={listState} />
            </section>
          </div>

          <aside className="grid gap-4 sm:grid-cols-2 xl:block xl:space-y-5">
            <StatusPanel
              title="Queue Processing"
              rows={[
                ["Indexed Processing", documents.length, documents.length || 1],
                ["Parsing Processing", uploadState === "uploading" ? 1 : 0, 1],
                [
                  "Queued",
                  documents.filter((item) => item.status !== "ready").length,
                  documents.length || 1,
                ],
              ]}
            />
            <StatusPanel
              title="System Health"
              rows={[
                ["Index Progress", 100, 100],
                ["System Status", 85, 100],
                ["System Health", 100, 100],
              ]}
            />
          </aside>
        </div>
      </AppShell>
    </AuthGuard>
  );
}

function StatusPanel({
  title,
  rows,
}: {
  title: string;
  rows: Array<[string, number, number]>;
}) {
  return (
    <section className="glass-panel rounded-lg p-5">
      <div className="flex items-center justify-between">
        <h2 className="text-xs font-semibold uppercase text-ink">{title}</h2>
        <span className="h-3 w-3 rounded-full bg-purple-300 shadow-[0_0_18px_rgba(216,132,255,0.9)]" />
      </div>
      <div className="mt-5 space-y-4">
        {rows.map(([label, value, total]) => {
          const percent = Math.min(100, Math.round((value / total) * 100));
          return (
            <div key={label}>
              <div className="flex items-center justify-between gap-3 text-sm">
                <span className="text-ink">{label}</span>
                <span className="text-muted">
                  {value}/{total}
                </span>
              </div>
              <div className="glass-progress mt-2 h-2">
                <span style={{ width: `${percent}%` }} />
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
}

function DocumentTable({
  documents,
  listState,
}: {
  documents: Document[];
  listState: "loading" | "ready" | "error";
}) {
  if (listState === "loading") {
    return <p className="px-6 py-8 text-sm text-muted">Loading documents...</p>;
  }

  if (listState === "error") {
    return (
      <p className="px-6 py-8 text-sm text-red-300">
        Documents unavailable.
      </p>
    );
  }

  if (documents.length === 0) {
    return <p className="px-6 py-8 text-sm text-muted">No documents yet.</p>;
  }

  return (
    <>
      <div className="space-y-3 p-4 md:hidden">
        {documents.map((document) => (
          <article key={document.id} className="glass-soft rounded-lg p-4">
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <p className="truncate font-medium text-ink">
                  {document.originalName}
                </p>
                <p className="mt-1 text-xs text-muted">
                  {formatBytes(document.sizeBytes)} · {document.chunkCount} chunks
                </p>
              </div>
              <span className="shrink-0 rounded-full border border-emerald-300/30 bg-emerald-400/10 px-2.5 py-1 text-xs font-medium text-emerald-100">
                {document.status}
              </span>
            </div>
            <p className="mt-3 truncate text-xs text-muted">{document.id}</p>
            <p className="mt-2 text-xs text-muted">
              Created {formatDate(document.createdAt)}
            </p>
          </article>
        ))}
      </div>

      <div className="hidden overflow-x-auto md:block">
      <table className="min-w-full divide-y divide-line text-left text-sm">
        <thead className="bg-white/40 text-xs uppercase text-muted">
          <tr>
            <th scope="col" className="px-6 py-3 font-semibold">
              Name
            </th>
            <th scope="col" className="px-4 py-3 font-semibold">
              Status
            </th>
            <th scope="col" className="px-4 py-3 font-semibold">
              Chunks
            </th>
            <th scope="col" className="px-4 py-3 font-semibold">
              Size
            </th>
            <th scope="col" className="px-4 py-3 font-semibold">
              Created
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-line">
          {documents.map((document) => (
            <tr key={document.id} className="align-top transition hover:bg-white/30">
              <td className="max-w-[280px] px-6 py-4">
                <p className="truncate font-medium text-ink">{document.originalName}</p>
                <p className="mt-1 truncate text-xs text-muted">{document.id}</p>
              </td>
              <td className="px-4 py-4">
                <span className="rounded-full border border-emerald-300/30 bg-emerald-400/10 px-2.5 py-1 text-xs font-medium text-emerald-100">
                  {document.status}
                </span>
              </td>
              <td className="px-4 py-4 text-muted">{document.chunkCount}</td>
              <td className="px-4 py-4 text-muted">{formatBytes(document.sizeBytes)}</td>
              <td className="whitespace-nowrap px-4 py-4 text-muted">
                {formatDate(document.createdAt)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      </div>
    </>
  );
}

function validateFile(file: File): string {
  const extension = file.name.split(".").pop()?.toLowerCase() ?? "";
  if (!allowedExtensions.has(extension)) {
    return "Only PDF and TXT files are allowed.";
  }

  if (file.size > maxUploadBytes) {
    return `File must be ${formatBytes(maxUploadBytes)} or smaller.`;
  }

  return "";
}

function readMaxUploadBytes(): number {
  const value = Number(process.env.NEXT_PUBLIC_MAX_UPLOAD_BYTES);
  if (!Number.isFinite(value) || value <= 0) {
    throw new Error("NEXT_PUBLIC_MAX_UPLOAD_BYTES is not configured");
  }
  return value;
}

function formatBytes(bytes: number): string {
  const units = ["B", "KB", "MB", "GB"];
  let value = bytes;
  let unitIndex = 0;

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }

  return `${value >= 10 || unitIndex === 0 ? value.toFixed(0) : value.toFixed(1)} ${units[unitIndex]}`;
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat("th-TH", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}
