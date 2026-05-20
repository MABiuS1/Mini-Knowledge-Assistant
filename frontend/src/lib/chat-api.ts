import { apiRequest } from "@/lib/api";
import type { ChatRequest, ChatResponse } from "@/types/api";

export function sendChatMessage(payload: ChatRequest): Promise<ChatResponse> {
  return apiRequest<ChatResponse>("/api/chat", {
    method: "POST",
    body: payload,
  });
}
