export type User = {
  id: string;
  username: string;
};

export type AuthResponse = {
  user: User;
};

export type LoginRequest = {
  username: string;
  password: string;
};

export type Document = {
  id: string;
  originalName: string;
  storedName: string;
  mimeType: string;
  sizeBytes: number;
  status: string;
  chunkCount: number;
  createdAt: string;
};

export type DocumentsResponse = {
  documents: Document[];
};

export type DocumentResponse = {
  document: Document;
};

export type ApiError = {
  error: {
    message: string;
  };
};
