import { useState, useEffect } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  User,
  Link2,
  Link2Off,
  RefreshCw,
  LogOut,
  Palette,
  Moon,
  Sun,
  DollarSign,
  ExternalLink,
} from "lucide-react";
import { useNavigate } from "react-router-dom";
import api from "@/lib/api";
import { useAuthStore } from "@/store/authStore";
import { useThemeStore, ACCENTS } from "@/store/themeStore";
import type { MonoAccount, CurrencyRate } from "@/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { formatMoney } from "@/lib/utils";

interface UserData {
  id: number;
  email: string;
  name: string;
}

interface MonoStatus {
  connected: boolean;
  lastImportAt: string | null;
}

const CCY: Record<number, string> = { 980: "UAH", 840: "USD", 978: "EUR", 826: "GBP" };

export function ProfilePage() {
  const qc = useQueryClient();
  const navigate = useNavigate();
  const logout = useAuthStore((s) => s.logout);
  const setUser = useAuthStore((s) => s.setUser);
  const { mode, toggleMode, accent, setAccent } = useThemeStore();

  const [saving, setSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [monoDialogOpen, setMonoDialogOpen] = useState(false);
  const [monoToken, setMonoToken] = useState("");
  const [connecting, setConnecting] = useState(false);
  const [connectError, setConnectError] = useState("");
  const [importing, setImporting] = useState(false);
  const [importMsg, setImportMsg] = useState("");

  const { data: userData, isLoading: userLoading } = useQuery({
    queryKey: ["me"],
    queryFn: () => api.get<{ data: UserData }>("/users/me").then((r) => r.data.data),
  });

  useEffect(() => {
    if (userData) setUser(userData);
  }, [userData, setUser]);

  const [name, setName] = useState<string>("");
  const effectiveName = name || userData?.name || "";

  const { data: monoStatus, refetch: refetchMono } = useQuery({
    queryKey: ["mono-status"],
    queryFn: () =>
      api.get<{ data: MonoStatus }>("/mono/status").then((r) => r.data.data),
    retry: false,
  });

  const { data: accounts } = useQuery({
    queryKey: ["mono-accounts"],
    queryFn: () =>
      api
        .get<{ data: { name: string; accounts: MonoAccount[] } }>("/mono/accounts")
        .then((r) => r.data.data.accounts),
    enabled: monoStatus?.connected === true,
    retry: false,
  });

  const { data: rates } = useQuery({
    queryKey: ["mono-currency"],
    queryFn: () =>
      api.get<{ data: CurrencyRate[] }>("/mono/currency").then((r) => r.data.data),
    enabled: monoStatus?.connected === true,
    retry: false,
  });

  async function handleSaveProfile(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setSaveSuccess(false);
    setSaveError("");
    try {
      const n = effectiveName;
      await api.put("/users/me", { name: n });
      setSaveSuccess(true);
      qc.invalidateQueries({ queryKey: ["me"] });
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message ?? "Помилка збереження";
      setSaveError(msg);
    } finally {
      setSaving(false);
    }
  }

  async function handleConnect(e: React.FormEvent) {
    e.preventDefault();
    setConnecting(true);
    setConnectError("");
    try {
      const res = await api.post<{ data: { accounts: MonoAccount[] } }>("/mono/connect", {
        token: monoToken,
      });
      setMonoDialogOpen(false);
      setMonoToken("");
      await refetchMono();
      qc.invalidateQueries({ queryKey: ["mono-accounts"] });
      // Одразу підтягуємо виписку, щоб дані з'явилися без зайвого кліку.
      const acc = pickAccount(res.data.data.accounts ?? []);
      if (acc) void runImport(acc.id, acc.currencyCode);
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message ?? "Невірний токен або помилка";
      setConnectError(msg);
    } finally {
      setConnecting(false);
    }
  }

  async function handleDisconnect() {
    if (!confirm("Від'єднати Monobank?")) return;
    await api.delete("/mono/connect");
    await refetchMono();
    qc.invalidateQueries({ queryKey: ["mono-accounts"] });
  }

  // Для імпорту обираємо гривневий рахунок (980), інакше — перший доступний.
  function pickAccount(list: MonoAccount[]): MonoAccount | undefined {
    return list.find((a) => a.currencyCode === 980) ?? list[0];
  }

  async function runImport(account: string, currencyCode: number) {
    setImporting(true);
    setImportMsg("");
    try {
      const res = await api.post<{ data: { imported: number } }>("/mono/import", {
        account,
        currencyCode,
      });
      const n = res.data.data.imported;
      setImportMsg(
        n > 0 ? `Імпортовано ${n} транзакцій за останні 31 день` : "Нових транзакцій не знайдено",
      );
      qc.invalidateQueries({ queryKey: ["transactions"] });
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message ?? "Помилка імпорту";
      setImportMsg(msg);
    } finally {
      setImporting(false);
    }
  }

  function handleImport() {
    const acc = pickAccount(accounts ?? []);
    if (!acc) {
      setImportMsg("Немає рахунків для імпорту");
      return;
    }
    void runImport(acc.id, acc.currencyCode);
  }

  function handleLogout() {
    logout();
    navigate("/");
  }

  const uahRates = (rates ?? []).filter(
    (r) => r.currencyCodeB === 980 && (r.currencyCodeA === 840 || r.currencyCodeA === 978)
  );

  if (userLoading) return <p className="text-muted-foreground">Завантаження...</p>;

  return (
    <div className="mx-auto max-w-xl space-y-5">
      <h1 className="text-xl font-semibold">Профіль</h1>

      {/* Особисті дані */}
      <section className="rounded-2xl border border-border bg-card p-5">
        <div className="mb-4 flex items-center gap-2">
          <User size={16} className="text-primary" />
          <span className="font-medium text-sm">Особисті дані</span>
        </div>
        <form onSubmit={handleSaveProfile} className="space-y-3">
          <div>
            <label className="block text-xs font-medium text-muted-foreground mb-1">Email</label>
            <Input type="email" value={userData?.email ?? ""} disabled />
          </div>
          <div>
            <label className="block text-xs font-medium text-muted-foreground mb-1">Ім'я</label>
            <Input
              type="text"
              value={effectiveName}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          {saveSuccess && <p className="text-sm text-positive">Збережено</p>}
          {saveError && <p className="text-sm text-negative">{saveError}</p>}
          <Button type="submit" disabled={saving} size="sm">
            {saving ? "Збереження..." : "Зберегти"}
          </Button>
        </form>
      </section>

      {/* Тема */}
      <section className="rounded-2xl border border-border bg-card p-5">
        <div className="mb-4 flex items-center gap-2">
          <Palette size={16} className="text-primary" />
          <span className="font-medium text-sm">Вигляд</span>
        </div>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Темна тема</span>
            <button
              onClick={toggleMode}
              className="flex h-9 w-9 items-center justify-center rounded-xl border border-border hover:bg-muted transition-colors"
            >
              {mode === "dark" ? <Sun size={16} /> : <Moon size={16} />}
            </button>
          </div>
          <div>
            <p className="text-sm text-muted-foreground mb-2">Акцентний колір</p>
            <div className="flex gap-2">
              {ACCENTS.map((a) => (
                <button
                  key={a.value}
                  onClick={() => setAccent(a.value)}
                  className="h-7 w-7 rounded-full border-2 transition-all"
                  style={{
                    backgroundColor: a.color,
                    borderColor: accent === a.value ? a.color : "transparent",
                    outline: accent === a.value ? `2px solid ${a.color}` : "none",
                    outlineOffset: 2,
                  }}
                  title={a.value}
                />
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* Monobank */}
      <section className="rounded-2xl border border-border bg-card p-5">
        <div className="mb-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="flex h-6 w-6 items-center justify-center rounded-md bg-black text-white text-xs font-bold">m</div>
            <span className="font-medium text-sm">Monobank</span>
          </div>
          {monoStatus?.connected ? (
            <span className="inline-flex items-center gap-1 rounded-full bg-positive/10 px-2.5 py-0.5 text-xs font-medium text-positive">
              Підключено
            </span>
          ) : (
            <span className="inline-flex items-center gap-1 rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium text-muted-foreground">
              Не підключено
            </span>
          )}
        </div>

        {!monoStatus?.connected ? (
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">
              Підключіть Monobank для автоматичного імпорту транзакцій та актуального курсу валют.
            </p>
            <Button size="sm" onClick={() => setMonoDialogOpen(true)}>
              <Link2 size={14} /> Підключити
            </Button>
          </div>
        ) : (
          <div className="space-y-4">
            {monoStatus.lastImportAt && (
              <p className="text-xs text-muted-foreground">
                Останній імпорт: {new Date(monoStatus.lastImportAt).toLocaleString("uk-UA")}
              </p>
            )}

            {accounts && accounts.length > 0 && (
              <div className="space-y-2">
                <p className="text-xs font-medium text-muted-foreground uppercase">Рахунки</p>
                {accounts.map((a) => (
                  <div
                    key={a.id}
                    className="flex items-center justify-between rounded-xl bg-muted/40 px-3 py-2.5 text-sm"
                  >
                    <span className="text-muted-foreground">
                      {a.maskedPan || a.iban || a.type}
                    </span>
                    <span className="font-medium">
                      {formatMoney(a.balance / 100)} {CCY[a.currencyCode] ?? a.currencyCode}
                    </span>
                  </div>
                ))}
              </div>
            )}

            {uahRates.length > 0 && (
              <div className="space-y-2">
                <p className="text-xs font-medium text-muted-foreground uppercase">Курс валют</p>
                <div className="flex gap-3">
                  {uahRates.map((r) => {
                    const ccy = CCY[r.currencyCodeA] ?? r.currencyCodeA;
                    const rate = r.rateBuy || r.rateCross;
                    return (
                      <div
                        key={r.currencyCodeA}
                        className="flex items-center gap-2 rounded-xl bg-muted/40 px-3 py-2"
                      >
                        <DollarSign size={14} className="text-primary" />
                        <span className="text-sm font-medium">
                          {ccy}/{CCY[r.currencyCodeB]}
                        </span>
                        <span className="text-sm text-muted-foreground">{rate?.toFixed(2)}</span>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}

            {importMsg && (
              <p className="text-sm text-muted-foreground">{importMsg}</p>
            )}

            <div className="flex gap-2 pt-1">
              <Button size="sm" onClick={handleImport} disabled={importing}>
                <RefreshCw size={14} className={importing ? "animate-spin" : ""} />
                {importing ? "Імпорт..." : "Імпортувати"}
              </Button>
              <Button size="sm" variant="outline" onClick={handleDisconnect}>
                <Link2Off size={14} /> Від'єднати
              </Button>
            </div>
          </div>
        )}
      </section>

      {/* Вийти */}
      <section className="rounded-2xl border border-border bg-card p-5">
        <div className="flex items-center justify-between">
          <div>
            <p className="font-medium text-sm">Вийти з акаунта</p>
            <p className="text-xs text-muted-foreground mt-0.5">Токен буде видалено з браузера</p>
          </div>
          <Button variant="destructive" size="sm" onClick={handleLogout}>
            <LogOut size={14} /> Вийти
          </Button>
        </div>
      </section>

      <Dialog open={monoDialogOpen} onOpenChange={setMonoDialogOpen}>
        <DialogContent title="Підключити Monobank">
          <form onSubmit={handleConnect} className="space-y-4 pt-1">
            <div className="rounded-xl border border-border bg-muted/40 p-3.5 space-y-3">
              <p className="text-sm text-muted-foreground">
                Токен видається на сайті{" "}
                <span className="font-medium text-foreground">api.monobank.ua</span> (не в самому
                застосунку), скануванням QR-коду:
              </p>
              <ol className="list-decimal space-y-1.5 pl-4 text-sm text-muted-foreground marker:font-medium marker:text-primary">
                <li>
                  Відкрийте{" "}
                  <a
                    href="https://api.monobank.ua/"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="font-medium text-primary hover:underline"
                  >
                    api.monobank.ua
                  </a>{" "}
                  на комп'ютері
                </li>
                <li>У застосунку monobank натисніть «Сканувати QR»</li>
                <li>Наведіть на QR-код із сайту та підтвердіть вхід</li>
                <li>Скопіюйте створений токен і вставте нижче</li>
              </ol>
              <a
                href="https://api.monobank.ua/"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 rounded-lg bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
              >
                <ExternalLink size={14} /> Відкрити api.monobank.ua
              </a>
            </div>
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1">
                Персональний токен
              </label>
              <Input
                type="password"
                required
                value={monoToken}
                onChange={(e) => setMonoToken(e.target.value)}
                placeholder="u•••••••••••••••"
                autoFocus
              />
              <p className="mt-1.5 text-xs text-muted-foreground">
                🔒 Зберігається зашифрованим (AES-GCM). Імпорт — не частіше ніж раз на 60 с (ліміт
                Monobank).
              </p>
            </div>
            {connectError && <p className="text-sm text-negative">{connectError}</p>}
            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={() => setMonoDialogOpen(false)}>
                Скасувати
              </Button>
              <Button type="submit" disabled={connecting}>
                {connecting ? "Перевірка..." : "Підключити"}
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
