import { apiRequest, apiURL } from "@/lib/api";
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

export async function streamChatMessage(
  payload: ChatRequest,
  onDelta: (content: string) => void,
): Promise<ChatResponse> {
  const response = await fetch(`${apiURL()}/api/chat/stream`, {
    method: "POST",
    body: JSON.stringify(payload),
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
  });

  if (!response.ok || !response.body) {
    throw new Error(`Streaming request failed with ${response.status}`);
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let doneResponse: ChatResponse | null = null;

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }

    buffer += decoder.decode(value, { stream: true });
    const parsed = parseSSEBuffer(buffer);
    buffer = parsed.remainder;

    for (const event of parsed.events) {
      if (event.event === "delta") {
        const payload = JSON.parse(event.data) as { content?: string };
        if (payload.content) {
          onDelta(payload.content);
        }
      }

      if (event.event === "done") {
        doneResponse = JSON.parse(event.data) as ChatResponse;
      }
    }
  }

  if (!doneResponse) {
    throw new Error("Streaming response ended without completion metadata");
  }

  return doneResponse;
}

function parseSSEBuffer(buffer: string): {
  events: Array<{ event: string; data: string }>;
  remainder: string;
} {
  const parts = buffer.split("\n\n");
  const remainder = parts.pop() ?? "";
  const events = parts
    .map((part) => {
      const lines = part.split("\n");
      const event = lines
        .find((line) => line.startsWith("event:"))
        ?.replace("event:", "")
        .trim();
      const data = lines
        .filter((line) => line.startsWith("data:"))
        .map((line) => line.replace("data:", "").trim())
        .join("\n");

      if (!event || !data) {
        return null;
      }

      return { event, data };
    })
    .filter((event): event is { event: string; data: string } => event !== null);

  return { events, remainder };
}
