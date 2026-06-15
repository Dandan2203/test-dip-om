import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2, PlusCircle, Target } from "lucide-react";
import api from "@/lib/api";
import type { Goal } from "@/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { cn, formatMoney } from "@/lib/utils";

interface FormData {
  title: string;
  targetAmount: string;
  deadline: string;
}

const emptyForm: FormData = { title: "", targetAmount: "", deadline: "" };

export function GoalsPage() {
  const qc = useQueryClient();
  const [modalOpen, setModalOpen] = useState(false);
  const [form, setForm] = useState<FormData>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [contributeId, setContributeId] = useState<number | null>(null);
  const [contributeAmount, setContributeAmount] = useState("");
  const [contributing, setContributing] = useState(false);

  const { data: goals = [], isLoading } = useQuery({
    queryKey: ["goals"],
    queryFn: () => api.get<{ data: Goal[] }>("/goals").then((r) => r.data.data),
  });

  // Підсвічування цілі, на яку перейшли з чату (?highlight=ID).
  // flashId виводимо з URL під час рендера; через 2.5 с прибираємо параметр,
  // тож CSS-анімація програється один раз і не повторюється при оновленні.
  const [searchParams, setSearchParams] = useSearchParams();
  const flashId = Number(searchParams.get("highlight")) || null;

  useEffect(() => {
    if (!flashId) return;
    document
      .querySelector(`[data-goal-id="${flashId}"]`)
      ?.scrollIntoView({ behavior: "smooth", block: "center" });
    const t = setTimeout(() => {
      searchParams.delete("highlight");
      setSearchParams(searchParams, { replace: true });
    }, 2500);
    return () => clearTimeout(t);
  }, [flashId, searchParams, setSearchParams]);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    try {
      await api.post("/goals", {
        title: form.title,
        targetAmount: parseFloat(form.targetAmount),
        deadline: form.deadline || null,
      });
      setModalOpen(false);
      setForm(emptyForm);
      qc.invalidateQueries({ queryKey: ["goals"] });
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message ?? "Помилка";
      setError(msg);
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(id: number) {
    if (!confirm("Видалити ціль?")) return;
    await api.delete(`/goals/${id}`);
    qc.invalidateQueries({ queryKey: ["goals"] });
  }

  async function handleContribute(e: React.FormEvent) {
    e.preventDefault();
    if (!contributeId) return;
    setContributing(true);
    try {
      await api.post(`/goals/${contributeId}/contribute`, {
        amount: parseFloat(contributeAmount),
      });
      setContributeId(null);
      setContributeAmount("");
      qc.invalidateQueries({ queryKey: ["goals"] });
    } finally {
      setContributing(false);
    }
  }

  if (isLoading) return <p className="text-muted-foreground">Завантаження...</p>;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Фінансові цілі</h1>
        <Button
          size="sm"
          onClick={() => {
            setForm(emptyForm);
            setError("");
            setModalOpen(true);
          }}
        >
          <Plus size={15} /> Нова ціль
        </Button>
      </div>

      {goals.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-border bg-card px-6 py-16 text-center">
          <Target size={40} className="mx-auto mb-3 text-muted-foreground/40" />
          <p className="font-medium">Цілей ще немає</p>
          <p className="mt-1 text-sm text-muted-foreground">Додайте першу фінансову ціль.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {goals.map((g) => {
            const pct = Math.min(100, Math.round((g.currentAmount / g.targetAmount) * 100));
            const done = pct >= 100;
            return (
              <div
                key={g.id}
                data-goal-id={g.id}
                className={cn(
                  "rounded-2xl border border-border bg-card p-5",
                  flashId === g.id && "flash-highlight",
                )}
              >
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h3 className="font-medium">{g.title}</h3>
                    {g.deadline && (
                      <p className="text-xs text-muted-foreground mt-0.5">до {g.deadline}</p>
                    )}
                  </div>
                  <div className="flex gap-1">
                    <button
                      onClick={() => {
                        setContributeId(g.id);
                        setContributeAmount("");
                      }}
                      className="p-1 text-muted-foreground hover:text-primary transition-colors"
                      title="Поповнити"
                      disabled={done}
                    >
                      <PlusCircle size={15} />
                    </button>
                    <button
                      onClick={() => handleDelete(g.id)}
                      className="p-1 text-muted-foreground hover:text-negative transition-colors"
                    >
                      <Trash2 size={15} />
                    </button>
                  </div>
                </div>

                <div className="w-full rounded-full h-2 bg-muted mb-2.5 overflow-hidden">
                  <div
                    className={`h-2 rounded-full transition-all ${done ? "bg-positive" : "bg-primary"}`}
                    style={{ width: `${pct}%` }}
                  />
                </div>
                <div className="flex justify-between text-xs text-muted-foreground">
                  <span>{formatMoney(g.currentAmount)}</span>
                  <span className={done ? "text-positive font-medium" : ""}>
                    {pct}% з {formatMoney(g.targetAmount)}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}

      <Dialog open={modalOpen} onOpenChange={setModalOpen}>
        <DialogContent title="Нова ціль">
          <form onSubmit={handleSave} className="space-y-3 pt-1">
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1">Назва</label>
              <Input
                type="text"
                required
                value={form.title}
                onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
                placeholder="Наприклад: Відпустка"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1">
                Цільова сума (₴)
              </label>
              <Input
                type="number"
                required
                min="1"
                step="0.01"
                value={form.targetAmount}
                onChange={(e) => setForm((f) => ({ ...f, targetAmount: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1">
                Дедлайн (необов'язково)
              </label>
              <Input
                type="date"
                value={form.deadline}
                onChange={(e) => setForm((f) => ({ ...f, deadline: e.target.value }))}
              />
            </div>
            {error && <p className="text-sm text-negative">{error}</p>}
            <div className="flex justify-end gap-2 pt-1">
              <Button type="button" variant="outline" onClick={() => setModalOpen(false)}>
                Скасувати
              </Button>
              <Button type="submit" disabled={saving}>
                {saving ? "Збереження..." : "Зберегти"}
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={contributeId !== null} onOpenChange={() => setContributeId(null)}>
        <DialogContent title="Поповнити ціль">
          <form onSubmit={handleContribute} className="space-y-3 pt-1">
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1">Сума (₴)</label>
              <Input
                type="number"
                required
                min="0.01"
                step="0.01"
                value={contributeAmount}
                onChange={(e) => setContributeAmount(e.target.value)}
                autoFocus
              />
            </div>
            <div className="flex justify-end gap-2 pt-1">
              <Button type="button" variant="outline" onClick={() => setContributeId(null)}>
                Скасувати
              </Button>
              <Button type="submit" disabled={contributing}>
                {contributing ? "..." : "Поповнити"}
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
