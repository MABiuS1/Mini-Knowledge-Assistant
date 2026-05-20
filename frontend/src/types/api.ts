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

export type Usage = {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
};

export type Citation = {
  documentId: string;
  fileName: string;
  chunkId: string;
  chunkIndex: number;
  snippet: string;
  similarity: number;
};

export type ConversationSummary = {
  id: string;
  title: string;
  createdAt: string;
  updatedAt: string;
};

export type ConversationMessage = {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  usage: Usage;
  createdAt: string;
};

export type ChatRequest = {
  message: string;
  conversationId?: string;
  documentIds?: string[];
};

export type ChatResponse = {
  answer: string;
  conversationId: string;
  usage: Usage;
  sessionTotalUsage: Usage;
  assistantMessageId: string;
  citations: Citation[];
};

export type ConversationsResponse = {
  conversations: ConversationSummary[];
};

export type ConversationDetailResponse = {
  conversation: ConversationSummary;
  messages: ConversationMessage[];
  sessionTotalUsage: Usage;
};

export type ApiError = {
  error: {
    message: string;
  };
};
