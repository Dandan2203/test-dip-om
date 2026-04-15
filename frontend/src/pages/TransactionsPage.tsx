import { Fragment, useMemo, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Pencil, Trash2, Download } from "lucide-react";
import api from "@/lib/api";
import type { Transaction, Category } from "@/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { formatDate, formatMoney, cn } from "@/lib/utils";

function monthLabel(key: string): string {
  const d = new Date(key + "-01T00:00:00");
  const s = d.toLocaleDateString("uk-UA", { month: "long", year: "numeric" });
  return s.charAt(0).toUpperCase() + s.slice(1);
}

const COLS = [
  { key: "date", label: "Дата", w: 120, min: 90, max: 200, align: "left" },
  { key: "desc", label: "Опис", w: 380, min: 180, max: 700, align: "left" },
  { key: "cat", label: "Категорія", w: 170, min: 100, max: 320, align: "left" },
  { key: "type", label: "Тип", w: 110, min: 90, max: 170, align: "left" },
  { key: "amount", label: "Сума", w: 130, min: 100, max: 240, align: "right" },
  { key: "actions", label: "", w: 80, min: 70, max: 130, align: "left" },
] as const;

interface Filters {
  from: string;
  to: string;
  type: string;
  category_id: string;
}

interface FormData {
  categoryId: string;
  type: "income" | "expense";
  amount: string;
  description: string;
  transactionDate: string;
}

const emptyForm: FormData = {
  categoryId: "",
  type: "expense",
  amount: "",
  description: "",
  transactionDate: new Date().toISOString().slice(0, 10),
};

const selCls =
  "border border-input rounded-lg px-3 py-1.5 text-sm bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary";
const PAGE_SIZE = 50;

export function TransactionsPage() {
  const qc = useQueryClient();
  const [filters, setFilters] = useState<Filters>({ from: "", to: "", type: "", category_id: "" });
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Transaction | null>(null);
  const [form, setForm] = useState<FormData>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [deleteId, setDeleteId] = useState<number | null>(null);
  const [exporting, setExporting] = useState(false);
  const [colW, setColW] = useState<number[]>(COLS.map((c) => c.w));
  const [page, setPage] = useState(1);

  function startColResize(i: number, e: React.MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    const startX = e.clientX;
    const startW = colW[i];
    const { min, max } = COLS[i];
    const onMove = (ev: MouseEvent) => {
      const next = Math.max(min, Math.min(max, startW + (ev.clientX - startX)));
      setColW((prev) => prev.map((w, idx) => (idx === i ? next : w)));
    };
    const onUp = () => {
      document.removeEventListener("mousemove", onMove);
      document.removeEventListener("mouseup", onUp);
      document.body.style.userSelect = "";
    };
    document.body.style.userSelect = "none";
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  }

  const qs = new URLSearchParams({
    limit: PAGE_SIZE.toString(),
    offset: ((page - 1) * PAGE_SIZE).toString(),
  });
  if (filters.from) qs.set("from", filters.from);
  if (filters.to) qs.set("to", filters.to);
  if (filters.type) qs.set("type", filters.type);
  if (filters.category_id) qs.set("category_id", filters.category_id);

  const { data, isLoading } = useQuery({
    queryKey: ["transactions", filters, page],
    queryFn: () =>
      api
        .get<{ data: { items: Transaction[]; total: number } }>(`/transactions?${qs}`)
        .then((r) => r.data.data),
  });

  const { data: categories = [] } = useQuery({
    queryKey: ["categories"],
    queryFn: () => api.get<{ data: Category[] }>("/categories").then((r) => r.data.data),
  });

  const totalPages = Math.max(1, Math.ceil((data?.total ?? 0) / PAGE_SIZE));

  function updateFilters(next: Filters) {
    setFilters(next);
    setPage(1);
  }

  function openCreate() {
    setEditing(null);
    setForm(emptyForm);
    setError("");
    setModalOpen(true);
  }

  function openEdit(t: Transaction) {
    setEditing(t);
    setForm({
      categoryId: t.categoryId?.toString() ?? "",
      type: t.type,
      amount: t.amount.toString(),
      description: t.description,
      transactionDate: t.transactionDate,
    });
    setError("");
    setModalOpen(true);
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    const body = {
      categoryId: form.categoryId ? Number(form.categoryId) : null,
      type: form.type,
      amount: parseFloat(form.amount),
      description: form.description,
      transactionDate: form.transactionDate,
    };
    try {
      if (editing) {
        await api.put(`/transactions/${editing.id}`, body);
      } else {
        await api.post("/transactions", body);
      }
      setModalOpen(false);
      qc.invalidateQueries({ queryKey: ["transactions"] });
      qc.invalidateQueries({ queryKey: ["dashboard"] });
      qc.invalidateQueries({ queryKey: ["by-category"] });
      qc.invalidateQueries({ queryKey: ["anomalies"] });
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message ?? "Помилка збереження";
      setError(msg);
    } finally {
      setSaving(false);
    }
  }

  async function confirmDelete() {
    if (deleteId === null) return;
    await api.delete(`/transactions/${deleteId}`);
    setDeleteId(null);
    if ((data?.items.length ?? 0) === 1 && page > 1) setPage(page - 1);
    qc.invalidateQueries({ queryKey: ["transactions"] });
    qc.invalidateQueries({ queryKey: ["dashboard"] });
    qc.invalidateQueries({ queryKey: ["by-category"] });
    qc.invalidateQueries({ queryKey: ["anomalies"] });
  }

  async function handleExport() {
    setExporting(true);
    try {
      const p = new URLSearchParams({ limit: "9999" });
      if (filters.from) p.set("from", filters.from);
      if (filters.to) p.set("to", filters.to);
      if (filters.type) p.set("type", filters.type);
      if (filters.category_id) p.set("category_id", filters.category_id);
      const res = await api.get(`/transactions/export?${p}`, { responseType: "blob" });
      const url = URL.createObjectURL(res.data as Blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `finagent_export_${new Date().toISOString().slice(0, 10)}.csv`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
    } finally {
      setExporting(false);
    }
  }

  const categoryName = (id: number | null) =>
    categories.find((c) => c.id === id)?.name ?? "—";

  const groups = useMemo(() => {
    const items = data?.items ?? [];
    const map = new Map<string, Transaction[]>();
    for (const t of items) {
      const key = t.transactionDate.slice(0, 7);
      const arr = map.get(key);
      if (arr) arr.push(t);
      else map.set(key, [t]);
    }
    return Array.from(map.entries())
      .sort((a, b) => b[0].localeCompare(a[0]))
      .map(([key, items]) => {
        const net = items.reduce(
          (acc, t) => acc + (t.type === "income" ? t.amount : -t.amount),
          0,
        );
        return { key, items, net };
      });
  }, [data]);

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h1 className="text-xl font-semibold">Транзакції</h1>
        <div className="grid grid-cols-2 gap-2 sm:flex">
          <Button variant="outline" size="sm" onClick={handleExport} disabled={exporting} className="w-full sm:w-auto">
            <Download size={15} /> {exporting ? "Експорт..." : "CSV"}
          </Button>
          <Button size="sm" onClick={openCreate} className="w-full sm:w-auto">
            <Plus size={15} /> Додати
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-2 rounded-xl border border-border bg-card px-4 py-3 sm:flex sm:flex-wrap sm:gap-3">
        <input
          type="date"
          value={filters.from}
          onChange={(e) => updateFilters({ ...filters, from: e.target.value })}
          className={`w-full sm:w-36 ${selCls}`}
        />
        <input
          type="date"
          value={filters.to}
          onChange={(e) => updateFilters({ ...filters, to: e.target.value })}
          className={`w-full sm:w-36 ${selCls}`}
        />
        <select
          value={filters.type}
          onChange={(e) => updateFilters({ ...filters, type: e.target.value })}
          className={`w-full sm:w-36 ${selCls}`}
        >
          <option value="">Всі типи</option>
          <option value="income">Доходи</option>
          <option value="expense">Витрати</option>
        </select>
        <select
          value={filters.category_id}
          onChange={(e) => updateFilters({ ...filters, category_id: e.target.value })}
          className={`w-full sm:w-48 ${selCls}`}
        >
          <option value="">Всі категорії</option>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </select>
        <button
          onClick={() => updateFilters({ from: "", to: "", type: "", category_id: "" })}
          className="col-span-2 text-sm text-muted-foreground transition-colors hover:text-foreground sm:col-span-1"
        >
          Скинути
        </button>
      </div>

      <div className="rounded-xl border border-border bg-card overflow-hidden">
        <div className="px-5 py-2.5 border-b border-border text-xs text-muted-foreground">
          Знайдено: {data?.total ?? 0}
        </div>
        {isLoading ? (
          <p className="px-5 py-10 text-sm text-muted-foreground text-center">Завантаження...</p>
        ) : (data?.items ?? []).length === 0 ? (
          <p className="px-5 py-10 text-sm text-muted-foreground text-center">Транзакцій не знайдено</p>
        ) : (
          <>
          <div className="divide-y divide-border md:hidden">
            {groups.map((g) => (
              <Fragment key={g.key}>
                <div className="flex items-center justify-between bg-primary-soft/60 px-4 py-3">
                  <span className="font-bold text-foreground">{monthLabel(g.key)}</span>
                  <span className={cn("font-bold", g.net >= 0 ? "text-positive" : "text-negative")}>
                    {g.net >= 0 ? "+" : "−"}
                    {formatMoney(Math.abs(g.net))}
                  </span>
                </div>
                {g.items.map((t) => (
                  <div key={t.id} className="space-y-3 px-4 py-3">
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <p className="truncate font-medium" title={t.description || undefined}>
                          {t.description || "Без опису"}
                        </p>
                        <p className="mt-1 text-xs text-muted-foreground">
                          {formatDate(t.transactionDate)} · {categoryName(t.categoryId)}
                        </p>
                      </div>
                      <div className="shrink-0 text-right">
                        <p
                          className={cn(
                            "font-semibold",
                            t.type === "income" ? "text-positive" : "text-negative",
                          )}
                        >
                          {t.type === "income" ? "+" : "−"}
                          {formatMoney(t.amount)}
                        </p>
                        <Badge variant={t.type === "income" ? "positive" : "negative"}>
                          {t.type === "income" ? "Дохід" : "Витрата"}
                        </Badge>
                      </div>
                    </div>
                    <div className="flex justify-end gap-2">
                      <button
                        onClick={() => openEdit(t)}
                        className="rounded-lg border border-border px-3 py-1.5 text-xs text-muted-foreground"
                      >
                        Редагувати
                      </button>
                      <button
                        onClick={() => setDeleteId(t.id)}
                        className="rounded-lg border border-border px-3 py-1.5 text-xs text-negative"
                      >
                        Видалити
                      </button>
                    </div>
                  </div>
                ))}
              </Fragment>
            ))}
          </div>

          <div className="hidden overflow-x-auto md:block">
            <table className="w-full table-fixed text-sm">
              <colgroup>
                {COLS.map((c, i) => (
                  <col key={c.key} style={{ width: colW[i] }} />
                ))}
              </colgroup>
              <thead className="bg-muted/40 text-xs text-muted-foreground uppercase">
                <tr>
                  {COLS.map((c, i) => (
                    <th
                      key={c.key}
                      className={cn(
                        "relative select-none px-4 py-3",
                        c.align === "right" ? "text-right" : "text-left",
                      )}
                    >
                      {c.label}
                      {i < COLS.length - 1 && (
                        <span
                          onMouseDown={(e) => startColResize(i, e)}
                          className="absolute -right-px top-0 z-10 h-full w-1.5 cursor-col-resize hover:bg-primary/50"
                          title="Перетягніть, щоб змінити ширину"
                        />
                      )}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {groups.map((g) => (
                  <Fragment key={g.key}>
                    <tr className="bg-primary-soft/60">
                      <td colSpan={4} className="border-l-4 border-primary px-4 py-3">
                        <span className="text-base font-bold text-foreground">
                          {monthLabel(g.key)}
                        </span>
                      </td>
                      <td
                        colSpan={2}
                        className={cn(
                          "px-4 py-3 text-right text-base font-bold whitespace-nowrap",
                          g.net >= 0 ? "text-positive" : "text-negative",
                        )}
                      >
                        {g.net >= 0 ? "+" : "−"}
                        {formatMoney(Math.abs(g.net))}
                      </td>
                    </tr>
                    {g.items.map((t) => (
                      <tr key={t.id} className="hover:bg-muted/30 transition-colors">
                        <td className="px-4 py-3 text-muted-foreground whitespace-nowrap">
                          {formatDate(t.transactionDate)}
                        </td>
                        <td className="px-4 py-3">
                          <span className="block min-w-0 truncate" title={t.description || undefined}>
                            {t.description || "—"}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-muted-foreground">
                          <span className="block max-w-32 truncate" title={categoryName(t.categoryId)}>
                            {categoryName(t.categoryId)}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <Badge variant={t.type === "income" ? "positive" : "negative"}>
                            {t.type === "income" ? "Дохід" : "Витрата"}
                          </Badge>
                        </td>
                        <td
                          className={`px-4 py-3 text-right font-medium whitespace-nowrap ${
                            t.type === "income" ? "text-positive" : "text-negative"
                          }`}
                        >
                          {t.type === "income" ? "+" : "−"}
                          {formatMoney(t.amount)}
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex justify-end gap-1">
                            <button
                              onClick={() => openEdit(t)}
                              className="p-1 text-muted-foreground hover:text-primary transition-colors"
                              title="Редагувати"
                            >
                              <Pencil size={14} />
                            </button>
                            <button
                              onClick={() => setDeleteId(t.id)}
                              className="p-1 text-muted-foreground hover:text-negative transition-colors"
                              title="Видалити"
                            >
                              <Trash2 size={14} />
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </Fragment>
                ))}
              </tbody>
            </table>
          </div>
          </>
        )}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-3">
          <Button
            variant="outline"
            size="sm"
            disabled={page === 1 || isLoading}
            onClick={() => setPage((p) => p - 1)}
          >
            Попередня
          </Button>
          <span className="text-sm text-muted-foreground">
            Сторінка {page} з {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page === totalPages || isLoading}
            onClick={() => setPage((p) => p + 1)}
          >
            Наступна
          </Button>
        </div>
      )}

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(o) => !o && setDeleteId(null)}
        title="Видалити транзакцію?"
        description="Транзакцію буде видалено. Цю дію можна скасувати через «Відмінити»."
        confirmLabel="Видалити"
        destructive
        onConfirm={confirmDelete}
      />

      <Dialog open={modalOpen} onOpenChange={setModalOpen}>
        <DialogContent title={editing ? "Редагувати транзакцію" : "Нова транзакція"}>
          <form onSubmit={handleSave} className="space-y-3 pt-1">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">Тип</label>
                <select
                  value={form.type}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, type: e.target.value as "income" | "expense" }))
                  }
                  className={`w-full ${selCls}`}
                >
                  <option value="expense">Витрата</option>
                  <option value="income">Дохід</option>
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">Сума (₴)</label>
                <Input
                  type="number"
                  required
                  min="0.01"
                  step="0.01"
                  value={form.amount}
                  onChange={(e) => setForm((f) => ({ ...f, amount: e.target.value }))}
                />
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1">Категорія</label>
              <select
                value={form.categoryId}
                onChange={(e) => setForm((f) => ({ ...f, categoryId: e.target.value }))}
                className={`w-full ${selCls}`}
              >
                <option value="">Без категорії</option>
                {categories
                  .filter((c) => c.type === form.type)
                  .map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                      {c.isSystem ? "" : " (власна)"}
                    </option>
                  ))}
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1">Опис</label>
              <Input
                type="text"
                value={form.description}
                onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1">Дата</label>
              <Input
                type="date"
                required
                value={form.transactionDate}
                onChange={(e) => setForm((f) => ({ ...f, transactionDate: e.target.value }))}
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
    </div>
  );
}
