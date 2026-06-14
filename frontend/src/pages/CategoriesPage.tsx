import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2 } from "lucide-react";
import api from "@/lib/api";
import type { Category } from "@/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const selCls =
  "border border-input rounded-lg px-3 py-2 text-sm bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary";

export function CategoriesPage() {
  const qc = useQueryClient();
  const [newName, setNewName] = useState("");
  const [newType, setNewType] = useState<"income" | "expense">("expense");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const { data: categories = [], isLoading } = useQuery({
    queryKey: ["categories"],
    queryFn: () => api.get<{ data: Category[] }>("/categories").then((r) => r.data.data),
  });

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    try {
      await api.post("/categories", { name: newName, type: newType });
      setNewName("");
      qc.invalidateQueries({ queryKey: ["categories"] });
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
    if (!confirm("Видалити категорію?")) return;
    await api.delete(`/categories/${id}`);
    qc.invalidateQueries({ queryKey: ["categories"] });
  }

  const byType = (type: "income" | "expense") => categories.filter((c) => c.type === type);

  if (isLoading) return <p className="text-muted-foreground">Завантаження...</p>;

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-semibold">Категорії</h1>

      <form
        onSubmit={handleCreate}
        className="flex flex-wrap gap-3 items-end rounded-xl border border-border bg-card px-4 py-4"
      >
        <div className="flex-1 min-w-40">
          <label className="block text-xs font-medium text-muted-foreground mb-1">Назва</label>
          <Input
            type="text"
            required
            maxLength={100}
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="Нова категорія"
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-muted-foreground mb-1">Тип</label>
          <select
            value={newType}
            onChange={(e) => setNewType(e.target.value as "income" | "expense")}
            className={selCls}
          >
            <option value="expense">Витрата</option>
            <option value="income">Дохід</option>
          </select>
        </div>
        <Button type="submit" disabled={saving}>
          <Plus size={15} /> {saving ? "..." : "Додати"}
        </Button>
        {error && <p className="w-full text-sm text-negative">{error}</p>}
      </form>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {(["income", "expense"] as const).map((type) => (
          <div key={type} className="rounded-xl border border-border bg-card overflow-hidden">
            <div
              className={`px-4 py-3 text-sm font-medium border-b border-border ${
                type === "income"
                  ? "text-positive bg-positive/10"
                  : "text-negative bg-negative/10"
              }`}
            >
              {type === "income" ? "Доходи" : "Витрати"}
            </div>
            {byType(type).length === 0 ? (
              <p className="px-4 py-4 text-sm text-muted-foreground">Немає категорій</p>
            ) : (
              <ul className="divide-y divide-border">
                {byType(type).map((c) => (
                  <li key={c.id} className="flex items-center justify-between px-4 py-2.5">
                    <span className="text-sm">
                      {c.name}
                      {c.isSystem && (
                        <span className="ml-1.5 text-xs text-muted-foreground">(система)</span>
                      )}
                    </span>
                    {!c.isSystem && (
                      <button
                        onClick={() => handleDelete(c.id)}
                        className="p-1 text-muted-foreground hover:text-negative transition-colors"
                      >
                        <Trash2 size={14} />
                      </button>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
