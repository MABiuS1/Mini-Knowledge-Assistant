"use client";

import type { FormEvent } from "react";
import { useEffect, useRef, useState } from "react";
import { AppShell } from "@/components/app-shell";
import { AuthGuard } from "@/components/auth-guard";
import { CitationList } from "@/components/citation-list";
import { MarkdownContent } from "@/components/markdown-content";
import {
  listConversations,
  loadConversation,
  sendChatMessage,
  streamChatMessage,
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

const selectedConversationKey = "knowledge-assistant:selected-conversation";

export default function ChatPage() {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [documentsState, setDocumentsState] = useState<
    "loading" | "ready" | "error"
  >("loading");
  const [conversations, setConversations] = useState<ConversationSummary[]>([]);
  const [historyState, setHistoryState] = useState<
    "loading" | "ready" | "error"
  >("loading");
  const [selectedDocumentIds, setSelectedDocumentIds] = useState<string[]>([]);
  const [conversationId, setConversationId] = useState("");
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [draft, setDraft] = useState("");
  const [isSending, setIsSending] = useState(false);
  const [error, setError] = useState("");
  const [isDocumentPickerOpen, setIsDocumentPickerOpen] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    async function loadDocuments() {
      setDocumentsState("loading");
      try {
        const response = await listDocuments();
        setDocuments(response.documents);
        setSelectedDocumentIds((current) =>
          current.length > 0
            ? current
            : response.documents.map((document) => document.id),
        );
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
    const assistantDraftId = `assistant-${Date.now()}`;
    const assistantDraft: ChatMessage = {
      id: assistantDraftId,
      role: "assistant",
      content: "",
    };
    setMessages((current) => [...current, userMessage, assistantDraft]);

    try {
      const request = {
        message,
        conversationId: conversationId || undefined,
        documentIds: selectedDocumentIds,
      };

      let streamedContent = "";
      let response: Awaited<ReturnType<typeof sendChatMessage>>;
      try {
        response = await streamChatMessage(request, (content) => {
          streamedContent += content;
          setMessages((current) =>
            current.map((item) =>
              item.id === assistantDraftId
                ? { ...item, content: streamedContent }
                : item,
            ),
          );
        });
      } catch {
        response = await sendChatMessage(request);
      }

      setConversationId(response.conversationId);
      window.localStorage.setItem(
        selectedConversationKey,
        response.conversationId,
      );
      setMessages((current) =>
        current.map((item) =>
          item.id === assistantDraftId
            ? {
                id: response.assistantMessageId,
                role: "assistant",
                content: response.answer,
                usage: response.usage,
                citations: response.citations,
              }
            : item,
        ),
      );
      void refreshConversations();
    } catch (sendError) {
      setError(sendError instanceof Error ? sendError.message : "Chat failed.");
      setMessages((current) =>
        current.filter(
          (item) => item.id !== userMessage.id && item.id !== assistantDraftId,
        ),
      );
      setDraft(message);
    } finally {
      setIsSending(false);
    }
  }

  function handleNewConversation() {
    setConversationId("");
    setMessages([]);
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

  function handleToggleDocument(documentId: string) {
    setSelectedDocumentIds((current) =>
      current.includes(documentId)
        ? current.filter((id) => id !== documentId)
        : [...current, documentId],
    );
  }

  async function handleLoadConversation(nextConversationId: string) {
    setError("");
    try {
      const response = await loadConversation(nextConversationId);
      setConversationId(response.conversation.id);
      setMessages(toChatMessages(response.messages));
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
      <AppShell
        onNewChat={handleNewConversation}
        sidebarContent={
          <div>
            <p className="mb-2 flex items-center gap-3 px-2 text-sm text-ink">
              <span className="w-5 text-center text-lg leading-none">↺</span>
              Recent History
            </p>
            <ConversationList
              conversations={conversations}
              activeConversationId={conversationId}
              state={historyState}
              onSelect={(id) => void handleLoadConversation(id)}
            />
          </div>
        }
      >
        <section className="relative overflow-hidden rounded-lg">
          <div className="relative flex min-h-[620px] flex-col items-center px-2 pb-8 pt-16 text-center sm:min-h-[calc(100vh-2rem)] sm:px-4 sm:pb-12 sm:pt-12">
            {messages.length === 0 ? (
              <div className="relative z-10 flex flex-1 flex-col items-center justify-center pb-8">
                <span className="cosmic-orb !relative mb-10 h-20 w-20 opacity-95 sm:h-[102px] sm:w-[102px]" />
                <h1 className="mx-auto max-w-[520px] text-3xl font-semibold leading-[1.12] text-ink sm:text-[46px]">
                  Ready to Ask Your Documents?
                </h1>
              </div>
            ) : (
              <div className="relative z-10 mb-5 mt-2 max-h-[calc(100vh-280px)] w-full max-w-3xl flex-1 space-y-4 overflow-y-auto text-left sm:mb-6 sm:mt-8 sm:max-h-[calc(100vh-300px)] sm:space-y-5">
                {messages.map((message) => (
                  <MessageBubble key={message.id} message={message} />
                ))}
                <div ref={messagesEndRef} />
              </div>
            )}

            <form
              onSubmit={handleSubmit}
              className="prompt-box relative z-20 mb-4 w-full max-w-[680px] px-3 py-3 sm:mb-6 sm:px-4"
            >
              <textarea
                value={draft}
                onChange={(event) => setDraft(event.target.value)}
                placeholder="Ask about your documents..."
                rows={1}
                disabled={isSending}
                className="min-h-[48px] w-full resize-none bg-transparent px-1 py-1 text-base text-ink outline-none placeholder:text-muted/70 disabled:cursor-not-allowed disabled:opacity-60 sm:min-h-[56px] sm:px-2 sm:text-lg"
              />

              <div className="mt-1 border-t border-line pt-2 sm:pt-3">
                <div className="relative flex items-center justify-between gap-3">
                  <DocumentAttachPicker
                    documents={documents}
                    selectedDocumentIds={selectedDocumentIds}
                    state={documentsState}
                    isOpen={isDocumentPickerOpen}
                    onOpenChange={setIsDocumentPickerOpen}
                    onToggle={handleToggleDocument}
                  />

                  <button
                    type="submit"
                    disabled={isSending || draft.trim() === ""}
                    className="prompt-send flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-2xl leading-none disabled:cursor-not-allowed disabled:opacity-50 sm:h-11 sm:w-11"
                    aria-label="Send message"
                  >
                    {isSending ? "…" : "↑"}
                  </button>
                </div>
              </div>
            </form>

            {error ? (
              <div className="mt-4 w-full max-w-xl rounded-md border border-red-400/30 bg-red-500/10 px-3 py-2 text-sm text-red-200">
                {error}
              </div>
            ) : null}
          </div>
          <div className="absolute bottom-3 left-4 right-4 h-px bg-line" />
        </section>
      </AppShell>
    </AuthGuard>
  );
}

function DocumentAttachPicker({
  documents,
  selectedDocumentIds,
  state,
  isOpen,
  onOpenChange,
  onToggle,
}: {
  documents: Document[];
  selectedDocumentIds: string[];
  state: "loading" | "ready" | "error";
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  onToggle: (documentId: string) => void;
}) {
  const selectedDocuments = documents.filter((document) =>
    selectedDocumentIds.includes(document.id),
  );
  const summary =
    selectedDocuments.length === 0
      ? "No documents selected"
      : selectedDocuments.length === documents.length
        ? "All documents"
        : selectedDocuments.length === 1
          ? selectedDocuments[0].originalName
          : `${selectedDocuments.length} documents selected`;

  return (
    <div className="relative text-left">
      <button
        type="button"
        onClick={() => onOpenChange(!isOpen)}
        className="flex h-12 min-w-0 items-center gap-3 rounded-full px-3 text-muted hover:bg-white/10 hover:text-ink"
        aria-expanded={isOpen}
        aria-label="Select document context"
      >
        <span className="flex h-9 w-9 items-center justify-center rounded-full border border-line text-2xl leading-none sm:h-8 sm:w-8">
          +
        </span>
        <span className="max-w-[130px] truncate text-sm sm:max-w-[240px]">
          {summary}
        </span>
      </button>

      {isOpen ? (
        <div className="absolute bottom-[calc(100%+0.75rem)] left-0 z-30 w-[calc(100vw-2rem)] max-w-[420px] rounded-lg border border-line bg-[#111329]/95 p-3 shadow-[0_18px_60px_rgba(0,0,0,0.45)] backdrop-blur-2xl sm:w-[420px]">
          <div className="mb-2 flex items-center justify-between gap-3">
            <p className="text-xs font-semibold uppercase text-muted">
              Select uploaded files
            </p>
            <span className="text-xs text-muted">
              {selectedDocumentIds.length}/{documents.length}
            </span>
          </div>

          {state === "loading" ? (
            <p className="rounded-md px-3 py-2 text-sm text-muted">
              Loading documents...
            </p>
          ) : null}

          {state === "error" ? (
            <p className="rounded-md border border-red-400/30 bg-red-500/10 px-3 py-2 text-sm text-red-200">
              Documents unavailable.
            </p>
          ) : null}

          {state === "ready" && documents.length === 0 ? (
            <p className="rounded-md px-3 py-2 text-sm text-muted">
              Upload documents first to ask grounded questions.
            </p>
          ) : null}

          {state === "ready" && documents.length > 0 ? (
            <div className="max-h-48 overflow-y-auto pr-1">
              {documents.map((document) => {
                const selected = selectedDocumentIds.includes(document.id);
                return (
                  <label
                    key={document.id}
                    className="flex cursor-pointer items-center gap-3 rounded-md px-3 py-2 text-sm text-ink hover:bg-white/10"
                  >
                    <input
                      type="checkbox"
                      checked={selected}
                      onChange={() => onToggle(document.id)}
                      className="h-4 w-4 shrink-0 accent-brand"
                    />
                    <span className="min-w-0 flex-1 truncate">
                      {document.originalName}
                    </span>
                <span className="hidden shrink-0 text-xs text-muted sm:inline">
                  {document.chunkCount} chunks
                </span>
                  </label>
                );
              })}
            </div>
          ) : null}
        </div>
      ) : null}
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
    return <p className="px-2 py-3 text-sm text-muted">Loading history...</p>;
  }

  if (state === "error") {
    return <p className="px-2 py-3 text-sm text-red-300">History unavailable.</p>;
  }

  if (conversations.length === 0) {
    return <p className="px-2 py-3 text-sm text-muted">No conversations yet.</p>;
  }

  const visibleConversations = conversations.slice(0, 3);

  return (
    <div className="space-y-1">
      {visibleConversations.map((conversation, index) => {
        const active = conversation.id === activeConversationId;
        return (
          <button
            key={conversation.id}
            type="button"
            onClick={() => onSelect(conversation.id)}
            className={
              active
                ? "glass-active flex w-full items-center gap-3 rounded-md px-2 py-2 text-left text-sm"
                : "flex w-full items-center gap-3 rounded-md px-2 py-2 text-left text-sm text-ink hover:bg-white/10"
            }
          >
            <span className="w-5 shrink-0 text-center text-lg leading-none text-muted">
              {index === 0 ? "◔" : "□"}
            </span>
            <span className="block min-w-0 truncate">
              {conversation.title}
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
            ? "max-w-[80%] rounded-lg border border-white/20 bg-brand/70 px-4 py-3 text-sm text-white shadow-sm shadow-brand/20"
            : "glass-soft max-w-[80%] rounded-lg px-4 py-3 text-sm text-ink"
        }
      >
        {isUser ? (
          <p className="whitespace-pre-wrap leading-6">{message.content}</p>
        ) : (
          <MarkdownContent content={message.content || "Thinking..."} />
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
