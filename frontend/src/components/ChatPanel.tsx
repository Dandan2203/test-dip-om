import { useState, useRef, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { X, Send, Sparkles, Undo2, Check } from "lucide-react";
import { AnimatePresence, motion } from "framer-motion";
import { useQueryClient } from "@tanstack/react-query";
import api from "@/lib/api";
import { useChatStore } from "@/store/chatStore";
import { Markdown } from "@/components/Markdown";
import type { ChatMessage, ChatAction, PendingAction } from "@/types";
import { cn } from "@/lib/utils";

const QUICK_PROMPTS = [
  "Скільки я витратив цього місяця?",
  "На що йде найбільше грошей?",
  "Чи зможу я накопичити на ціль?",
  "Дай пораду щодо бюджету",
];

const INTENT_LABELS: Record<string, string> = {
  FINANCIAL: "Фінанси",
  SCENARIO: "Сценарій",
  ADVICE: "Порада",
  GENERAL: "FinAgent",
  ACTION: "Дія",
};

const UNDO_TIMEOUT = 8000;
const MAX_MESSAGE_LEN = 500;

export function ChatPanel() {
  const { open, setOpen, width, setWidth } = useChatStore();
  const qc = useQueryClient();
  const navigate = useNavigate();

  function startResize(e: React.MouseEvent) {
    e.preventDefault();
    const onMove = (ev: MouseEvent) => setWidth(window.innerWidth - ev.clientX);
    const onUp = () => {
      document.removeEventListener("mousemove", onMove);
      document.removeEventListener("mouseup", onUp);
      document.body.style.userSelect = "";
    };
    document.body.style.userSelect = "none";
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  }

  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [undo, setUndo] = useState<ChatAction | null>(null);
  const [undoing, setUndoing] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);
  const undoTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, loading, undo]);

  useEffect(() => {
    const onResize = () => setWidth(useChatStore.getState().width);
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, [setWidth]);

  useEffect(() => () => {
    if (undoTimer.current) clearTimeout(undoTimer.current);
  }, []);

  useEffect(() => {
    if (!open || messages.length > 0) return;
    api
      .get<{ data: { messages: ChatMessage[] } }>("/chat/history")
      .then((res) => setMessages(res.data.data.messages ?? []))
      .catch((err: unknown) => {
        console.warn("Не вдалося завантажити історію чату", err);
      });
  }, [open, messages.length]);

  function showUndo(action: ChatAction | null) {
    if (undoTimer.current) clearTimeout(undoTimer.current);
    setUndo(action);
    if (action) undoTimer.current = setTimeout(() => setUndo(null), UNDO_TIMEOUT);
  }

  async function executeAction(pending: PendingAction): Promise<ChatAction | null> {
    const res = await api.post<{ data: { executed: boolean; action?: ChatAction } }>(
      "/actions/execute",
      pending,
    );
    const executed = res.data.data.action ?? null;
    if (!res.data.data.executed || !executed) {
      throw new Error("Дію не виконано");
    }
    qc.invalidateQueries({ queryKey: ["transactions"] });
    qc.invalidateQueries({ queryKey: ["goals"] });
    qc.invalidateQueries({ queryKey: ["dashboard"] });
    qc.invalidateQueries({ queryKey: ["by-category"] });
    qc.invalidateQueries({ queryKey: ["anomalies"] });
    if (executed) showUndo(executed);
    if (pending.actionType === "create_goal" && executed) {
      navigate(`/app/goals?highlight=${executed.entityId}`);
    }
    return executed;
  }

  async function send(text: string) {
    const message = text.trim();
    if (!message || loading) return;

    const history = messages
      .filter((m) => m.content?.trim())
      .slice(-6)
      .map((m) => ({ role: m.role, content: m.content }));

    setMessages((p) => [
      ...p,
      { role: "user", content: message, timestamp: new Date().toISOString() },
    ]);
    setInput("");
    setLoading(true);
    try {
      const res = await api.post<{
        data: { response: string; intent: string; pendingAction?: PendingAction | null };
      }>("/chat", { message, history }, { timeout: 60000 });
      const { response, intent, pendingAction } = res.data.data;
      const needsConfirm = !!pendingAction?.requiresConfirm;

      setMessages((p) => [
        ...p,
        {
          role: "assistant",
          content: response,
          intent,
          pendingAction: needsConfirm ? pendingAction : null,
          pendingStatus: needsConfirm ? "pending" : undefined,
          timestamp: new Date().toISOString(),
        },
      ]);

    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message ?? "Вибачте, ШІ-сервіс зараз недоступний.";
      setMessages((p) => [
        ...p,
        { role: "assistant", content: msg, timestamp: new Date().toISOString() },
      ]);
    } finally {
      setLoading(false);
    }
  }

  async function confirmPending(index: number) {
    const msg = messages[index];
    if (!msg?.pendingAction) return;
    setMessages((p) =>
      p.map((m, i) =>
        i === index ? { ...m, pendingStatus: "executing", pendingError: undefined } : m,
      ),
    );
    try {
      await executeAction(msg.pendingAction);
      setMessages((p) => p.map((m, i) => (i === index ? { ...m, pendingStatus: "done" } : m)));
    } catch {
      setMessages((p) =>
        p.map((m, i) =>
          i === index
            ? { ...m, pendingStatus: "pending", pendingError: "Не вдалося виконати дію." }
            : m,
        ),
      );
    }
  }

  function cancelPending(index: number) {
    setMessages((p) => p.map((m, i) => (i === index ? { ...m, pendingStatus: "cancelled" } : m)));
  }

  function setPendingDeadline(index: number, deadline: string) {
    setMessages((p) =>
      p.map((m, i) =>
        i === index && m.pendingAction
          ? { ...m, pendingAction: { ...m.pendingAction, deadline: deadline || null } }
          : m,
      ),
    );
  }

  async function handleUndo() {
    setUndoing(true);
    try {
      await api.post("/actions/undo");
      showUndo(null);
      qc.invalidateQueries({ queryKey: ["transactions"] });
      qc.invalidateQueries({ queryKey: ["goals"] });
      qc.invalidateQueries({ queryKey: ["dashboard"] });
      qc.invalidateQueries({ queryKey: ["by-category"] });
      qc.invalidateQueries({ queryKey: ["anomalies"] });
    } catch (err: unknown) {
      console.warn("Не вдалося скасувати дію", err);
      setMessages((p) => [
        ...p,
        {
          role: "assistant",
          content: "Не вдалося скасувати останню дію.",
          timestamp: new Date().toISOString(),
        },
      ]);
    } finally {
      setUndoing(false);
    }
  }

  return (
    <AnimatePresence>
      {open && (
        <>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setOpen(false)}
            className="fixed inset-0 z-40 bg-black/30 md:hidden"
          />
          <motion.aside
            initial={{ x: "100%" }}
            animate={{ x: 0 }}
            exit={{ x: "100%" }}
            transition={{ duration: 0.2, ease: "easeOut" }}
            style={{ width }}
            className="fixed right-0 top-0 z-50 flex h-full max-md:!w-full flex-col border-l border-border bg-card"
          >
            <div
              onMouseDown={startResize}
              className="absolute left-0 top-0 z-20 hidden h-full w-1.5 cursor-col-resize transition-colors hover:bg-primary/40 md:block"
              title="Перетягніть, щоб змінити ширину"
            />
            <div className="flex items-center justify-between border-b border-border px-4 py-3">
              <div className="flex items-center gap-2">
                <Sparkles size={18} className="text-primary" />
                <span className="font-semibold">ШІ-асистент</span>
              </div>
              <button
                onClick={() => setOpen(false)}
                className="rounded-lg p-1 text-muted-foreground hover:bg-muted"
              >
                <X size={18} />
              </button>
            </div>

            <div className="flex-1 space-y-3 overflow-y-auto p-4">
              {messages.length === 0 && (
                <div className="mt-6 space-y-3">
                  <p className="text-center text-sm text-muted-foreground">
                    Запитайте про ваші фінанси природною мовою
                  </p>
                  <div className="space-y-2">
                    {QUICK_PROMPTS.map((q) => (
                      <button
                        key={q}
                        onClick={() => send(q)}
                        className="w-full rounded-xl border border-border bg-background px-3 py-2 text-left text-sm text-foreground transition-colors hover:bg-muted"
                      >
                        {q}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {messages.map((m, i) => (
                <div key={i} className="space-y-2">
                  {(m.role === "user" || m.content?.trim()) && (
                    <div className={cn("flex", m.role === "user" ? "justify-end" : "justify-start")}>
                      <div
                        className={cn(
                          "max-w-[85%] break-words rounded-2xl px-3 py-2 text-sm",
                          m.role === "user"
                            ? "whitespace-pre-wrap bg-primary text-primary-foreground"
                            : "bg-muted text-foreground",
                        )}
                      >
                        {m.role === "assistant" && m.intent && INTENT_LABELS[m.intent] && (
                          <span className="mb-1 block text-xs text-muted-foreground">
                            {INTENT_LABELS[m.intent]}
                          </span>
                        )}
                        {m.role === "assistant" ? <Markdown content={m.content} /> : m.content}
                      </div>
                    </div>
                  )}

                  {m.role === "assistant" && m.pendingAction && (
                    <div className="flex justify-start">
                      <div className="w-[90%] rounded-2xl border border-border bg-background p-3">
                        <p className="text-sm font-medium text-foreground">
                          {m.pendingAction.confirmText}
                        </p>
                        {m.pendingError && (
                          <p className="mt-1.5 text-xs text-negative">{m.pendingError}</p>
                        )}

                        {m.pendingAction.actionType === "create_goal" &&
                          m.pendingStatus === "pending" && (
                            <label className="mt-2.5 block">
                              <span className="mb-1 block text-xs text-muted-foreground">
                                Дедлайн (необов'язково)
                              </span>
                              <input
                                type="date"
                                value={m.pendingAction.deadline ?? ""}
                                onChange={(e) => setPendingDeadline(i, e.target.value)}
                                className="w-full rounded-lg border border-input bg-background px-2.5 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                              />
                            </label>
                          )}

                        {m.pendingStatus === "pending" ? (
                          <div className="mt-2.5 flex gap-2">
                            <button
                              onClick={() => confirmPending(i)}
                              className="flex flex-1 items-center justify-center gap-1.5 rounded-lg bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
                            >
                              <Check size={14} /> Так
                            </button>
                            <button
                              onClick={() => cancelPending(i)}
                              className="flex flex-1 items-center justify-center gap-1.5 rounded-lg border border-border px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted"
                            >
                              <X size={14} /> Ні
                            </button>
                          </div>
                        ) : m.pendingStatus === "executing" ? (
                          <p className="mt-1.5 text-xs text-muted-foreground">Виконання...</p>
                        ) : m.pendingStatus === "cancelled" ? (
                          <p className="mt-1.5 text-xs text-muted-foreground">Скасовано</p>
                        ) : (
                          <p className="mt-1.5 flex items-center gap-1 text-xs text-positive">
                            <Check size={12} /> Виконано
                          </p>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              ))}

              {loading && (
                <div className="flex justify-start">
                  <div className="rounded-2xl bg-muted px-3 py-3 flex gap-1.5">
                    {[0, 1, 2].map((i) => (
                      <span
                        key={i}
                        className="h-1.5 w-1.5 rounded-full bg-muted-foreground animate-bounce"
                        style={{ animationDelay: `${i * 0.15}s` }}
                      />
                    ))}
                  </div>
                </div>
              )}

              <div ref={bottomRef} />
            </div>

            <AnimatePresence>
              {undo && (
                <motion.div
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: "auto", opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  className="shrink-0 overflow-hidden border-t border-border bg-muted/40"
                >
                  <div className="flex items-center justify-center gap-3 px-4 py-2">
                    <span className="text-xs text-muted-foreground">Дію виконано</span>
                    <button
                      onClick={handleUndo}
                      disabled={undoing}
                      className="flex items-center gap-1 text-xs font-medium text-primary hover:underline disabled:opacity-50"
                    >
                      <Undo2 size={12} />
                      {undoing ? "..." : "Відмінити"}
                    </button>
                    <button
                      onClick={() => showUndo(null)}
                      className="text-muted-foreground hover:text-foreground"
                      aria-label="Закрити"
                    >
                      <X size={13} />
                    </button>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>

            <div className="border-t border-border p-3">
              <div className="flex gap-2">
                <textarea
                  rows={1}
                  value={input}
                  maxLength={MAX_MESSAGE_LEN}
                  onChange={(e) => setInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && !e.shiftKey) {
                      e.preventDefault();
                      send(input);
                    }
                  }}
                  placeholder="Повідомлення..."
                  className="flex-1 resize-none rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                />
                <button
                  onClick={() => send(input)}
                  disabled={!input.trim() || loading}
                  className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-40"
                >
                  <Send size={16} />
                </button>
              </div>
              <div
                className={cn(
                  "mt-1 text-right text-[11px]",
                  input.length >= MAX_MESSAGE_LEN
                    ? "text-negative"
                    : input.length > MAX_MESSAGE_LEN * 0.9
                      ? "text-amber-500"
                      : "text-muted-foreground",
                )}
              >
                {input.length}/{MAX_MESSAGE_LEN}
              </div>
            </div>
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  );
}
