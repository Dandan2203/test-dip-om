import { useState, useEffect } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import {
  Wallet,
  LineChart,
  Target,
  AlertTriangle,
  Bot,
  CreditCard,
  Moon,
  Sun,
  MessageSquare,
} from "lucide-react";
import api from "@/lib/api";
import { useAuthStore } from "@/store/authStore";
import { useThemeStore } from "@/store/themeStore";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { cn } from "@/lib/utils";

const features = [
  {
    icon: Bot,
    title: "ШІ-агент",
    text: "Спілкуйтесь природною мовою — ставте запити, додавайте операції, отримуйте поради.",
  },
  {
    icon: LineChart,
    title: "Інтерактивна аналітика",
    text: "Налаштовуваний дашборд із графіками доходів, витрат та прогнозами.",
  },
  {
    icon: Target,
    title: "Фінансові цілі",
    text: "Ставте цілі, поповнюйте накопичення і відстежуйте прогрес у реальному часі.",
  },
  {
    icon: AlertTriangle,
    title: "Виявлення аномалій",
    text: "Автоматичні сповіщення при нетипових витратах на основі ваших паттернів.",
  },
  {
    icon: CreditCard,
    title: "Monobank",
    text: "Автоматичний імпорт транзакцій, актуальний курс валют і баланс рахунків.",
  },
  {
    icon: MessageSquare,
    title: "Що якщо?",
    text: "Моделюйте сценарії: «якщо скорочу витрати на X₴ — коли досягну цілі?»",
  },
];

const DEMO: { user: string; ai: string }[] = [
  {
    user: "Скільки я витратив цього місяця?",
    ai: "За червень — 14 200 ₴. Найбільша категорія: Продукти (4 600 ₴, 32%). Це на 6% менше, ніж у травні.",
  },
  {
    user: "Додай витрату 350 грн на кіно",
    ai: "✅ Додав «Кіно» — 350 ₴ у категорію Розваги. Якщо помилка — натисніть «Відмінити».",
  },
  {
    user: "Коли я накопичу 50 000 ₴ на відпустку?",
    ai: "За темпом ~3 800 ₴/міс ціль буде досягнута приблизно через 13 місяців — у липні 2027.",
  },
  {
    user: "На що йде найбільше грошей?",
    ai: "Топ-3 категорії: Продукти 32%, Транспорт 18%, Кафе 12%. Разом це 62% усіх витрат.",
  },
  {
    user: "Що буде, якщо скоротити кафе на 1 000 ₴/міс?",
    ai: "Зекономите 12 000 ₴ за рік. Ціль «Відпустка» наблизиться орієнтовно на 3 місяці.",
  },
  {
    user: "Чи були підозрілі витрати цього тижня?",
    ai: "Так, 1 аномалія: «Підписки» — 899 ₴, це втричі більше за ваш звичний рівень.",
  },
];

type DemoPhase = "typing" | "thinking" | "answer" | "out";

function ChatDemo() {
  const [idx, setIdx] = useState(0);
  const [typed, setTyped] = useState("");
  const [phase, setPhase] = useState<DemoPhase>("typing");

  const current = DEMO[idx];

  // Друк запиту користувача по буквах (швидко), потім — пауза.
  useEffect(() => {
    if (phase !== "typing") return;
    if (typed.length < current.user.length) {
      const t = setTimeout(() => setTyped(current.user.slice(0, typed.length + 1)), 34);
      return () => clearTimeout(t);
    }
    const t = setTimeout(() => setPhase("thinking"), 450);
    return () => clearTimeout(t);
  }, [phase, typed, current.user]);

  // «Думає» → відповідь.
  useEffect(() => {
    if (phase !== "thinking") return;
    const t = setTimeout(() => setPhase("answer"), 850);
    return () => clearTimeout(t);
  }, [phase]);

  // Тримаємо відповідь на екрані → плавне зникнення.
  useEffect(() => {
    if (phase !== "answer") return;
    const t = setTimeout(() => setPhase("out"), 2800);
    return () => clearTimeout(t);
  }, [phase]);

  // Зникло → наступний приклад (по колу).
  useEffect(() => {
    if (phase !== "out") return;
    const t = setTimeout(() => {
      setTyped("");
      setIdx((i) => (i + 1) % DEMO.length);
      setPhase("typing");
    }, 450);
    return () => clearTimeout(t);
  }, [phase]);

  const visible = phase !== "out";

  return (
    <div className="flex min-h-[300px] flex-col overflow-hidden rounded-2xl border border-border bg-card/80 p-4 backdrop-blur">
      <div className="mb-3 flex items-center gap-2 border-b border-border pb-3">
        <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-primary text-primary-foreground">
          <Bot size={14} />
        </div>
        <span className="text-sm font-medium">ШІ-асистент FinAgent</span>
        <span className="ml-auto flex h-2 w-2 rounded-full bg-positive animate-pulse" />
      </div>

      <div className="flex-1 space-y-2">
        <AnimatePresence mode="wait">
          {visible && (
            <motion.div
              key={idx}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.35 }}
              className="space-y-2"
            >
              {/* Запит користувача (друкується) */}
              <div className="flex justify-end">
                <div className="max-w-[85%] rounded-2xl bg-primary px-3.5 py-2 text-sm leading-relaxed text-primary-foreground">
                  {typed}
                  {phase === "typing" && (
                    <span className="ml-0.5 inline-block h-3.5 w-px -translate-y-px animate-pulse bg-primary-foreground/70 align-middle" />
                  )}
                </div>
              </div>

              {/* Індикатор набору відповіді */}
              {phase === "thinking" && (
                <div className="flex justify-start">
                  <div className="flex gap-1.5 rounded-2xl bg-muted px-4 py-3">
                    {[0, 1, 2].map((i) => (
                      <span
                        key={i}
                        className="h-1.5 w-1.5 animate-bounce rounded-full bg-muted-foreground"
                        style={{ animationDelay: `${i * 0.15}s` }}
                      />
                    ))}
                  </div>
                </div>
              )}

              {/* Відповідь ШІ */}
              {phase === "answer" && (
                <motion.div
                  initial={{ opacity: 0, y: 6 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.3 }}
                  className="flex justify-start"
                >
                  <div className="max-w-[85%] rounded-2xl bg-muted px-3.5 py-2 text-sm leading-relaxed text-foreground">
                    {current.ai}
                  </div>
                </motion.div>
              )}
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Індикатор поточного прикладу */}
      <div className="mt-3 flex justify-center gap-1.5">
        {DEMO.map((_, i) => (
          <span
            key={i}
            className={cn(
              "h-1.5 rounded-full transition-all duration-300",
              i === idx ? "w-4 bg-primary" : "w-1.5 bg-border",
            )}
          />
        ))}
      </div>
    </div>
  );
}

type AuthMode = "login" | "register" | null;

export function LandingPage() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const setToken = useAuthStore((s) => s.setToken);
  const navigate = useNavigate();
  const { mode, toggleMode } = useThemeStore();

  const [authMode, setAuthMode] = useState<AuthMode>(null);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [authError, setAuthError] = useState("");
  const [loading, setLoading] = useState(false);
  const [registered, setRegistered] = useState(false);

  if (isAuthenticated) return <Navigate to="/app" replace />;

  function openAuth(mode: AuthMode) {
    setEmail("");
    setPassword("");
    setName("");
    setAuthError("");
    setRegistered(false);
    setAuthMode(mode);
  }

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setAuthError("");
    try {
      const res = await api.post("/auth/login", { email, password });
      setToken(res.data.data.token);
      navigate("/app");
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message ?? "Помилка входу";
      setAuthError(msg);
    } finally {
      setLoading(false);
    }
  }

  async function handleRegister(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setAuthError("");
    try {
      await api.post("/auth/register", { email, password, name });
      setRegistered(true);
      setAuthMode("login");
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message ?? "Помилка реєстрації";
      setAuthError(msg);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      {/* Header */}
      <header className="sticky top-0 z-10 border-b border-border/60 bg-background/80 backdrop-blur">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-primary text-primary-foreground">
              <Wallet size={17} />
            </div>
            <span className="text-lg font-bold">FinAgent</span>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={toggleMode}
              className="flex h-9 w-9 items-center justify-center rounded-xl text-muted-foreground hover:bg-muted transition-colors"
            >
              {mode === "dark" ? <Sun size={17} /> : <Moon size={17} />}
            </button>
            <Button variant="ghost" size="sm" onClick={() => openAuth("login")}>
              Увійти
            </Button>
            <Button size="sm" onClick={() => openAuth("register")}>
              Реєстрація
            </Button>
          </div>
        </div>
      </header>

      {/* Hero */}
      <section className="mx-auto max-w-6xl px-6 pt-16 pb-12">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <div>
            <motion.h1
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
              className="text-4xl font-bold tracking-tight sm:text-5xl"
            >
              Ваш персональний{" "}
              <span className="text-primary">ШІ-фінансист</span>
            </motion.h1>
            <motion.p
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
              className="mt-5 text-lg text-muted-foreground"
            >
              Керуйте бюджетом у діалозі з ШІ. Підключіть Monobank, аналізуйте витрати,
              ставте цілі та отримуйте персональні поради.
            </motion.p>
            <motion.div
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
              className="mt-8 flex flex-wrap gap-3"
            >
              <Button size="lg" onClick={() => openAuth("register")}>
                Почати безкоштовно
              </Button>
              <Button size="lg" variant="outline" onClick={() => openAuth("login")}>
                Увійти
              </Button>
            </motion.div>
          </div>
          <motion.div
            initial={{ opacity: 0, scale: 0.96 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.6, delay: 0.2 }}
          >
            <ChatDemo />
          </motion.div>
        </div>
      </section>

      {/* Features */}
      <section className="mx-auto max-w-6xl px-6 pb-20">
        <h2 className="text-2xl font-bold text-center mb-8">Все що потрібно для фінансового контролю</h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {features.map(({ icon: Icon, title, text }, i) => (
            <motion.div
              key={title}
              initial={{ opacity: 0, y: 16 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: i * 0.06 }}
              className="rounded-2xl border border-border bg-card p-5"
            >
              <div className="mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-primary-soft text-primary">
                <Icon size={20} />
              </div>
              <h3 className="font-semibold">{title}</h3>
              <p className="mt-1 text-sm text-muted-foreground">{text}</p>
            </motion.div>
          ))}
        </div>
      </section>

      <footer className="border-t border-border py-8 text-center text-sm text-muted-foreground">
        FinAgent — дипломний проєкт ЖДТУ · {new Date().getFullYear()}
      </footer>

      {/* Auth: Login */}
      <Dialog open={authMode === "login"} onOpenChange={(o) => !o && setAuthMode(null)}>
        <DialogContent title="Вхід у FinAgent">
          {registered && (
            <p className="mb-3 rounded-xl bg-positive/10 px-3 py-2 text-sm text-positive">
              Реєстрацію завершено. Увійдіть у свій акаунт.
            </p>
          )}
          <form onSubmit={handleLogin} className="space-y-3 pt-1">
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1">Email</label>
              <Input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                autoFocus
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1">Пароль</label>
              <Input
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            {authError && <p className="text-sm text-negative">{authError}</p>}
            <Button type="submit" disabled={loading} className="w-full">
              {loading ? "Вхід..." : "Увійти"}
            </Button>
          </form>
          <p className="mt-4 text-center text-sm text-muted-foreground">
            Немає акаунта?{" "}
            <button
              onClick={() => openAuth("register")}
              className="font-medium text-primary hover:underline"
            >
              Зареєструватись
            </button>
          </p>
        </DialogContent>
      </Dialog>

      {/* Auth: Register */}
      <Dialog open={authMode === "register"} onOpenChange={(o) => !o && setAuthMode(null)}>
        <DialogContent title="Реєстрація в FinAgent">
          <form onSubmit={handleRegister} className="space-y-3 pt-1">
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1">Ім'я</label>
              <Input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                autoFocus
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1">Email</label>
              <Input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-muted-foreground mb-1">Пароль</label>
              <Input
                type="password"
                required
                minLength={8}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
              <p className="mt-1 text-xs text-muted-foreground">Мінімум 8 символів</p>
            </div>
            {authError && <p className="text-sm text-negative">{authError}</p>}
            <Button type="submit" disabled={loading} className="w-full">
              {loading ? "Реєстрація..." : "Зареєструватись"}
            </Button>
          </form>
          <p className="mt-4 text-center text-sm text-muted-foreground">
            Вже є акаунт?{" "}
            <button
              onClick={() => openAuth("login")}
              className="font-medium text-primary hover:underline"
            >
              Увійти
            </button>
          </p>
        </DialogContent>
      </Dialog>
    </div>
  );
}
