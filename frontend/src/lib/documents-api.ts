import { apiFormRequest, apiRequest } from "@/lib/api";
import type { DocumentResponse, DocumentsResponse } from "@/types/api";

export function listDocuments(): Promise<DocumentsResponse> {
  return apiRequest<DocumentsResponse>("/api/documents");
}

export function uploadDocument(file: File): Promise<DocumentResponse> {
  const formData = new FormData();
  formData.set("file", file);

  return apiFormRequest<DocumentResponse>("/api/documents/upload", formData, {
    method: "POST",
  });
}
