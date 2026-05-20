import { apiRequest } from "@/lib/api";
import type {
  ChatRequest,
  ChatResponse,
  ConversationDetailResponse,
  ConversationsResponse,
} from "@/types/api";

export function sendChatMessage(payload: ChatRequest): Promise<ChatResponse> {
  return apiRequest<ChatResponse>("/api/chat", {
    method: "POST",
    body: payload,
  });
}

export function listConversations(): Promise<ConversationsResponse> {
  return apiRequest<ConversationsResponse>("/api/chat/conversations");
}

export function loadConversation(
  conversationId: string,
): Promise<ConversationDetailResponse> {
  return apiRequest<ConversationDetailResponse>(
    `/api/chat/conversations/${conversationId}`,
  );
}
