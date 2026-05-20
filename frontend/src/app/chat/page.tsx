"use client";

import type { FormEvent } from "react";
import { useEffect, useMemo, useRef, useState } from "react";
import { AppShell } from "@/components/app-shell";
import { AuthGuard } from "@/components/auth-guard";
import { CitationList } from "@/components/citation-list";
import { MarkdownContent } from "@/components/markdown-content";
import {
  listConversations,
  loadConversation,
  sendChatMessage,
} from "@/lib/chat-api";
import { listDocuments } from "@/lib/documents-api";
import type {
  Citation,
  ConversationMessage,
  ConversationSummary,
  Document,
  Usage,
} from "@/types/api";

type ChatMessage = {
  id: string;
  role: "user" | "assistant";
  content: string;
  usage?: Usage;
  citations?: Citation[];
};

const emptyUsage: Usage = {
  promptTokens: 0,
  completionTokens: 0,
  totalTokens: 0,
};

const selectedConversationKey = "knowledge-assistant:selected-conversation";

export default function ChatPage() {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [conversations, setConversations] = useState<ConversationSummary[]>([]);
  const [historyState, setHistoryState] = useState<
    "loading" | "ready" | "error"
  >("loading");
  const [selectedDocumentIds, setSelectedDocumentIds] = useState<string[]>([]);
  const [documentsState, setDocumentsState] = useState<
    "loading" | "ready" | "error"
  >("loading");
  const [conversationId, setConversationId] = useState("");
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [draft, setDraft] = useState("");
  const [isSending, setIsSending] = useState(false);
  const [error, setError] = useState("");
  const [sessionUsage, setSessionUsage] = useState<Usage>(emptyUsage);
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const selectedDocuments = useMemo(
    () =>
      documents.filter((document) =>
        selectedDocumentIds.includes(document.id),
      ),
    [documents, selectedDocumentIds],
  );

  useEffect(() => {
    async function loadDocuments() {
      setDocumentsState("loading");
      try {
        const response = await listDocuments();
        setDocuments(response.documents);
        setDocumentsState("ready");
      } catch (loadError) {
        setError(
          loadError instanceof Error
            ? loadError.message
            : "Unable to load documents.",
        );
        setDocumentsState("error");
      }
    }

    void loadDocuments();
  }, []);

  useEffect(() => {
    async function initializeHistory() {
      const history = await refreshConversations();
      const savedConversationId = window.localStorage.getItem(
        selectedConversationKey,
      );

      if (
        savedConversationId &&
        history.some((conversation) => conversation.id === savedConversationId)
      ) {
        await handleLoadConversation(savedConversationId);
      }
    }

    void initializeHistory();
  }, []);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, isSending]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const message = draft.trim();
    if (!message || isSending) {
      return;
    }

    setDraft("");
    setError("");
    setIsSending(true);

    const userMessage: ChatMessage = {
      id: `user-${Date.now()}`,
      role: "user",
      content: message,
    };
    setMessages((current) => [...current, userMessage]);

    try {
      const response = await sendChatMessage({
        message,
        conversationId: conversationId || undefined,
        documentIds: selectedDocumentIds,
      });

      setConversationId(response.conversationId);
      window.localStorage.setItem(
        selectedConversationKey,
        response.conversationId,
      );
      setSessionUsage(response.sessionTotalUsage);
      setMessages((current) => [
        ...current,
        {
          id: response.assistantMessageId,
          role: "assistant",
          content: response.answer,
          usage: response.usage,
          citations: response.citations,
        },
      ]);
      void refreshConversations();
    } catch (sendError) {
      setError(sendError instanceof Error ? sendError.message : "Chat failed.");
      setMessages((current) => current.filter((item) => item.id !== userMessage.id));
      setDraft(message);
    } finally {
      setIsSending(false);
    }
  }

  function handleToggleDocument(documentId: string) {
    setSelectedDocumentIds((current) =>
      current.includes(documentId)
        ? current.filter((id) => id !== documentId)
        : [...current, documentId],
    );
  }

  function handleNewConversation() {
    setConversationId("");
    setMessages([]);
    setSessionUsage(emptyUsage);
    setError("");
    window.localStorage.removeItem(selectedConversationKey);
  }

  async function refreshConversations(): Promise<ConversationSummary[]> {
    setHistoryState("loading");
    try {
      const response = await listConversations();
      setConversations(response.conversations);
      setHistoryState("ready");
      return response.conversations;
    } catch (historyError) {
      setError(
        historyError instanceof Error
          ? historyError.message
          : "Unable to load conversations.",
      );
      setHistoryState("error");
      return [];
    }
  }

  async function handleLoadConversation(nextConversationId: string) {
    setError("");
    try {
      const response = await loadConversation(nextConversationId);
      setConversationId(response.conversation.id);
      setMessages(toChatMessages(response.messages));
      setSessionUsage(response.sessionTotalUsage);
      window.localStorage.setItem(
        selectedConversationKey,
        response.conversation.id,
      );
    } catch (loadError) {
      setError(
        loadError instanceof Error
          ? loadError.message
          : "Unable to load conversation.",
      );
      window.localStorage.removeItem(selectedConversationKey);
    }
  }

  return (
    <AuthGuard>
      <AppShell>
        <div className="grid min-h-[calc(100vh-9rem)] gap-6 lg:grid-cols-[280px_1fr]">
          <aside className="rounded-lg border border-line bg-white shadow-sm">
            <div className="border-b border-line px-5 py-4">
              <div className="flex items-center justify-between gap-3">
                <h1 className="text-base font-semibold text-ink">History</h1>
                <button
                  type="button"
                  onClick={handleNewConversation}
                  className="rounded-md border border-line px-2.5 py-1.5 text-xs font-medium text-ink hover:border-brand"
                >
                  New
                </button>
              </div>
              <div className="mt-3 max-h-[220px] overflow-y-auto">
                <ConversationList
                  conversations={conversations}
                  activeConversationId={conversationId}
                  state={historyState}
                  onSelect={(id) => void handleLoadConversation(id)}
                />
              </div>
            </div>

            <div className="border-b border-line px-5 py-4">
              <h1 className="text-base font-semibold text-ink">Context</h1>
              <p className="mt-1 text-sm text-muted">
                Select documents for grounded answers.
              </p>
            </div>

            <div className="max-h-[360px] overflow-y-auto p-3">
              <DocumentSelector
                documents={documents}
                selectedDocumentIds={selectedDocumentIds}
                state={documentsState}
                onToggle={handleToggleDocument}
              />
            </div>

            <div className="border-t border-line p-5">
              <p className="text-xs font-semibold uppercase text-muted">
                Session tokens
              </p>
              <UsageGrid usage={sessionUsage} />
            </div>
          </aside>

          <section className="flex min-h-0 flex-col rounded-lg border border-line bg-white shadow-sm">
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line px-6 py-4">
              <div>
                <h2 className="text-base font-semibold text-ink">Chat</h2>
                <p className="mt-1 text-sm text-muted">
                  {selectedDocuments.length > 0
                    ? `${selectedDocuments.length} document selected`
                    : "No document context selected"}
                </p>
              </div>
              {conversationId ? (
                <p className="max-w-[260px] truncate text-xs text-muted">
                  {conversationId}
                </p>
              ) : null}
            </div>

            <div className="min-h-0 flex-1 overflow-y-auto px-6 py-5">
              {messages.length === 0 ? (
                <div className="flex h-full min-h-[280px] items-center justify-center text-center">
                  <div>
                    <p className="text-base font-semibold text-ink">
                      Ask about your uploaded knowledge base.
                    </p>
                    <p className="mt-2 max-w-md text-sm text-muted">
                      Choose documents on the left, then send a question.
                    </p>
                  </div>
                </div>
              ) : (
                <div className="space-y-5">
                  {messages.map((message) => (
                    <MessageBubble key={message.id} message={message} />
                  ))}
                  {isSending ? (
                    <div className="max-w-[80%] rounded-lg border border-line bg-surface px-4 py-3 text-sm text-muted">
                      Thinking...
                    </div>
                  ) : null}
                  <div ref={messagesEndRef} />
                </div>
              )}
            </div>

            {error ? (
              <div className="mx-6 mb-3 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                {error}
              </div>
            ) : null}

            <form onSubmit={handleSubmit} className="border-t border-line p-4">
              <div className="flex gap-3">
                <textarea
                  value={draft}
                  onChange={(event) => setDraft(event.target.value)}
                  placeholder="Ask a question..."
                  rows={3}
                  disabled={isSending}
                  className="min-h-[76px] flex-1 resize-none rounded-md border border-line px-3 py-2 text-sm text-ink outline-none focus:border-brand disabled:cursor-not-allowed disabled:bg-surface"
                />
                <button
                  type="submit"
                  disabled={isSending || draft.trim() === ""}
                  className="self-end rounded-md bg-brand px-5 py-2.5 text-sm font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {isSending ? "Sending..." : "Send"}
                </button>
              </div>
            </form>
          </section>
        </div>
      </AppShell>
    </AuthGuard>
  );
}

function DocumentSelector({
  documents,
  selectedDocumentIds,
  state,
  onToggle,
}: {
  documents: Document[];
  selectedDocumentIds: string[];
  state: "loading" | "ready" | "error";
  onToggle: (documentId: string) => void;
}) {
  if (state === "loading") {
    return <p className="px-2 py-3 text-sm text-muted">Loading documents...</p>;
  }

  if (state === "error") {
    return <p className="px-2 py-3 text-sm text-red-700">Documents unavailable.</p>;
  }

  if (documents.length === 0) {
    return <p className="px-2 py-3 text-sm text-muted">Upload documents first.</p>;
  }

  return (
    <div className="space-y-2">
      {documents.map((document) => (
        <label
          key={document.id}
          className="flex cursor-pointer gap-3 rounded-md border border-transparent px-2 py-2 hover:border-line hover:bg-surface"
        >
          <input
            type="checkbox"
            checked={selectedDocumentIds.includes(document.id)}
            onChange={() => onToggle(document.id)}
            className="mt-1 h-4 w-4 rounded border-line text-brand"
          />
          <span className="min-w-0">
            <span className="block truncate text-sm font-medium text-ink">
              {document.originalName}
            </span>
            <span className="mt-1 block text-xs text-muted">
              {document.chunkCount} chunks
            </span>
          </span>
        </label>
      ))}
    </div>
  );
}

function ConversationList({
  conversations,
  activeConversationId,
  state,
  onSelect,
}: {
  conversations: ConversationSummary[];
  activeConversationId: string;
  state: "loading" | "ready" | "error";
  onSelect: (conversationId: string) => void;
}) {
  if (state === "loading") {
    return <p className="py-3 text-sm text-muted">Loading conversations...</p>;
  }

  if (state === "error") {
    return <p className="py-3 text-sm text-red-700">History unavailable.</p>;
  }

  if (conversations.length === 0) {
    return <p className="py-3 text-sm text-muted">No conversations yet.</p>;
  }

  return (
    <div className="space-y-2">
      {conversations.map((conversation) => {
        const active = conversation.id === activeConversationId;
        return (
          <button
            key={conversation.id}
            type="button"
            onClick={() => onSelect(conversation.id)}
            className={
              active
                ? "w-full rounded-md bg-brand px-3 py-2 text-left text-sm text-white"
                : "w-full rounded-md border border-transparent px-3 py-2 text-left text-sm text-ink hover:border-line hover:bg-surface"
            }
          >
            <span className="block truncate font-medium">
              {conversation.title}
            </span>
            <span
              className={
                active
                  ? "mt-1 block text-xs text-blue-100"
                  : "mt-1 block text-xs text-muted"
              }
            >
              {formatConversationDate(conversation.updatedAt)}
            </span>
          </button>
        );
      })}
    </div>
  );
}

function MessageBubble({ message }: { message: ChatMessage }) {
  const isUser = message.role === "user";

  return (
    <article className={isUser ? "flex justify-end" : "flex justify-start"}>
      <div
        className={
          isUser
            ? "max-w-[80%] rounded-lg bg-brand px-4 py-3 text-sm text-white"
            : "max-w-[80%] rounded-lg border border-line bg-surface px-4 py-3 text-sm text-ink"
        }
      >
        {isUser ? (
          <p className="whitespace-pre-wrap leading-6">{message.content}</p>
        ) : (
          <MarkdownContent content={message.content} />
        )}
        {message.usage ? (
          <div className="mt-3 border-t border-line pt-3">
            <UsageInline usage={message.usage} />
            <CitationList citations={message.citations ?? []} />
          </div>
        ) : null}
      </div>
    </article>
  );
}

function UsageGrid({ usage }: { usage: Usage }) {
  return (
    <div className="mt-3 grid grid-cols-3 gap-2">
      <UsageBox label="Prompt" value={usage.promptTokens} />
      <UsageBox label="Answer" value={usage.completionTokens} />
      <UsageBox label="Total" value={usage.totalTokens} />
    </div>
  );
}

function UsageBox({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-md border border-line bg-surface px-2 py-2">
      <p className="text-[11px] font-medium text-muted">{label}</p>
      <p className="mt-1 text-sm font-semibold text-ink">{value}</p>
    </div>
  );
}

function UsageInline({ usage }: { usage: Usage }) {
  return (
    <p className="text-xs text-muted">
      Tokens: {usage.totalTokens} total, {usage.promptTokens} prompt,{" "}
      {usage.completionTokens} answer
    </p>
  );
}

function toChatMessages(messages: ConversationMessage[]): ChatMessage[] {
  return messages
    .filter(isChatRoleMessage)
    .map((message) => ({
      id: message.id,
      role: message.role,
      content: message.content,
      usage:
        message.role === "assistant" && message.usage.totalTokens > 0
          ? message.usage
          : undefined,
    }));
}

function isChatRoleMessage(
  message: ConversationMessage,
): message is ConversationMessage & { role: "user" | "assistant" } {
  return message.role === "user" || message.role === "assistant";
}

function formatConversationDate(value: string): string {
  return new Intl.DateTimeFormat("th-TH", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(new Date(value));
}
