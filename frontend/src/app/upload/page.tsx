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
      setMessage(error instanceof Error ? error.message : "Unable to load documents");
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
        <div className="grid gap-6 lg:grid-cols-[minmax(280px,360px)_1fr]">
          <section className="rounded-lg border border-line bg-white p-6 shadow-sm">
            <div>
              <h1 className="text-xl font-semibold text-ink">Documents</h1>
              <p className="mt-1 text-sm text-muted">PDF and TXT, up to {formatBytes(maxUploadBytes)}</p>
            </div>

            <form onSubmit={handleUpload} className="mt-6 space-y-4">
              <label className="block">
                <span className="text-sm font-medium text-ink">File</span>
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
                  className="mt-2 block w-full text-sm text-muted file:mr-4 file:rounded-md file:border-0 file:bg-brand file:px-3 file:py-2 file:text-sm file:font-medium file:text-white hover:file:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
                />
              </label>

              {selectedFile ? (
                <div className="rounded-md border border-line bg-surface px-3 py-2">
                  <p className="truncate text-sm font-medium text-ink">
                    {selectedFile.name}
                  </p>
                  <p className="mt-1 text-xs text-muted">
                    {formatBytes(selectedFile.size)}
                  </p>
                </div>
              ) : null}

              {selectedFileError ? (
                <p className="text-sm font-medium text-red-700">{selectedFileError}</p>
              ) : null}

              <button
                type="submit"
                disabled={
                  uploadState === "uploading" || !selectedFile || Boolean(selectedFileError)
                }
                className="inline-flex w-full items-center justify-center rounded-md bg-brand px-4 py-2.5 text-sm font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {uploadState === "uploading" ? "Uploading..." : "Upload"}
              </button>
            </form>

            {message ? (
              <p
                className={
                  uploadState === "error" || listState === "error"
                    ? "mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700"
                    : "mt-4 rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-700"
                }
              >
                {message}
              </p>
            ) : null}
          </section>

          <section className="rounded-lg border border-line bg-white shadow-sm">
            <div className="flex items-center justify-between border-b border-line px-6 py-4">
              <h2 className="text-base font-semibold text-ink">Library</h2>
              <button
                type="button"
                onClick={() => void refreshDocuments()}
                disabled={listState === "loading"}
                className="rounded-md border border-line px-3 py-2 text-sm font-medium text-ink hover:border-brand disabled:cursor-not-allowed disabled:opacity-60"
              >
                {listState === "loading" ? "Loading..." : "Refresh"}
              </button>
            </div>

            <DocumentTable documents={documents} listState={listState} />
          </section>
        </div>
      </AppShell>
    </AuthGuard>
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
    return <p className="px-6 py-8 text-sm text-red-700">Documents unavailable.</p>;
  }

  if (documents.length === 0) {
    return <p className="px-6 py-8 text-sm text-muted">No documents yet.</p>;
  }

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-line text-left text-sm">
        <thead className="bg-surface text-xs uppercase text-muted">
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
            <tr key={document.id} className="align-top">
              <td className="max-w-[280px] px-6 py-4">
                <p className="truncate font-medium text-ink">{document.originalName}</p>
                <p className="mt-1 truncate text-xs text-muted">{document.id}</p>
              </td>
              <td className="px-4 py-4">
                <span className="rounded-full bg-green-50 px-2.5 py-1 text-xs font-medium text-green-700">
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
