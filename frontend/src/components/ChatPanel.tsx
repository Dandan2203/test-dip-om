import { useState, useRef, useEffect } from "react";
import { X, Send, Sparkles, Undo2, Check } from "lucide-react";
import { AnimatePresence, motion } from "framer-motion";
import { useQueryClient } from "@tanstack/react-query";
import api from "@/lib/api";
import { useChatStore } from "@/store/chatStore";
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
  GENERAL: "Загальне",
  ACTION: "Дія",
};

export function ChatPanel() {
  const { open, setOpen, width, setWidth } = useChatStore();
  const qc = useQueryClient();

  // Перетягування лівого краю змінює ширину панелі (clamp у сторі).
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
  const [undoing, setUndoing] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, loading]);

  const lastAction = [...messages].reverse().find((m) => m.role === "assistant" && m.action)?.action;

  async function send(text: string) {
    const message = text.trim();
    if (!message || loading) return;
    setMessages((p) => [
      ...p,
      { role: "user", content: message, timestamp: new Date().toISOString() },
    ]);
    setInput("");
    setLoading(true);
    try {
      const res = await api.post<{
        data: { response: string; intent: string; pendingAction?: PendingAction | null };
      }>("/chat", { message });
      const { response, intent, pendingAction } = res.data.data;
      setMessages((p) => [
        ...p,
        {
          role: "assistant",
          content: response,
          intent,
          pendingAction: pendingAction ?? null,
          pendingStatus: pendingAction ? "pending" : undefined,
          timestamp: new Date().toISOString(),
        },
      ]);
    } catch {
      setMessages((p) => [
        ...p,
        {
          role: "assistant",
          content: "Вибачте, ШІ-сервіс зараз недоступний.",
          timestamp: new Date().toISOString(),
        },
      ]);
    } finally {
      setLoading(false);
    }
  }

  async function confirmPending(index: number) {
    const msg = messages[index];
    if (!msg?.pendingAction) return;
    setMessages((p) => p.map((m, i) => (i === index ? { ...m, pendingStatus: "done" } : m)));
    try {
      const res = await api.post<{ data: { executed: boolean; action?: ChatAction } }>(
        "/actions/execute",
        msg.pendingAction,
      );
      const executed = res.data.data.action ?? null;
      // Прикріплюємо виконану дію до повідомлення — щоб увімкнувся банер «Відмінити».
      setMessages((p) => p.map((m, i) => (i === index ? { ...m, action: executed } : m)));
      qc.invalidateQueries({ queryKey: ["transactions"] });
      qc.invalidateQueries({ queryKey: ["goals"] });
      qc.invalidateQueries({ queryKey: ["dashboard"] });
    } catch {
      setMessages((p) => p.map((m, i) => (i === index ? { ...m, pendingStatus: "pending" } : m)));
    }
  }

  function cancelPending(index: number) {
    setMessages((p) => p.map((m, i) => (i === index ? { ...m, pendingStatus: "cancelled" } : m)));
  }

  async function handleUndo() {
    setUndoing(true);
    try {
      await api.post("/actions/undo");
      setMessages((p) =>
        p.map((m) =>
          m === [...p].reverse().find((x) => x.role === "assistant" && x.action)
            ? { ...m, action: null }
            : m,
        ),
      );
      qc.invalidateQueries({ queryKey: ["transactions"] });
      qc.invalidateQueries({ queryKey: ["goals"] });
    } catch {
      /* ignore */
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
            {/* Ручка зміни ширини (десктоп) */}
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
                  <div className={cn("flex", m.role === "user" ? "justify-end" : "justify-start")}>
                    <div
                      className={cn(
                        "max-w-[85%] whitespace-pre-wrap break-words rounded-2xl px-3 py-2 text-sm",
                        m.role === "user"
                          ? "bg-primary text-primary-foreground"
                          : "bg-muted text-foreground",
                      )}
                    >
                      {m.role === "assistant" && m.intent && INTENT_LABELS[m.intent] && (
                        <span className="mb-1 block text-xs text-muted-foreground">
                          {INTENT_LABELS[m.intent]}
                        </span>
                      )}
                      {m.content}
                    </div>
                  </div>

                  {/* Підтвердження дії: так / ні */}
                  {m.role === "assistant" && m.pendingAction && (
                    <div className="flex justify-start">
                      <div className="w-[90%] rounded-2xl border border-border bg-background p-3">
                        <p className="text-sm font-medium text-foreground">
                          {m.pendingAction.confirmText}
                        </p>
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

            {/* Undo-банер */}
            <AnimatePresence>
              {lastAction && (
                <motion.div
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: "auto", opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  className="overflow-hidden border-t border-amber-200/60 dark:border-amber-800/40 bg-amber-50/80 dark:bg-amber-900/20"
                >
                  <div className="flex items-center justify-between px-4 py-2.5">
                    <span className="text-xs text-amber-700 dark:text-amber-400">
                      Відмінити останню дію?
                    </span>
                    <button
                      onClick={handleUndo}
                      disabled={undoing}
                      className="flex items-center gap-1.5 rounded-lg bg-amber-100 dark:bg-amber-800/40 px-2.5 py-1 text-xs font-medium text-amber-700 dark:text-amber-400 hover:bg-amber-200 dark:hover:bg-amber-800/60 transition-colors disabled:opacity-50"
                    >
                      <Undo2 size={12} />
                      {undoing ? "..." : "Відмінити"}
                    </button>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>

            <div className="flex gap-2 border-t border-border p-3">
              <textarea
                rows={1}
                value={input}
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
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  );
}
