import { useEffect, useState } from "react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
import {
  LayoutDashboard,
  ArrowLeftRight,
  Target,
  AlertTriangle,
  Tag,
  User,
  LogOut,
  Moon,
  Sun,
  MessageSquare,
  Wallet,
} from "lucide-react";
import { useAuthStore } from "@/store/authStore";
import { useThemeStore } from "@/store/themeStore";
import { useChatStore } from "@/store/chatStore";
import { ChatPanel } from "@/components/ChatPanel";
import { cn } from "@/lib/utils";

const navItems = [
  { to: "/app", icon: LayoutDashboard, label: "Огляд", end: true },
  { to: "/app/transactions", icon: ArrowLeftRight, label: "Транзакції", end: false },
  { to: "/app/goals", icon: Target, label: "Цілі", end: false },
  { to: "/app/anomalies", icon: AlertTriangle, label: "Аномалії", end: false },
  { to: "/app/categories", icon: Tag, label: "Категорії", end: false },
];

function navClass({ isActive }: { isActive: boolean }) {
  return cn(
    "flex items-center gap-3 rounded-xl px-3 py-2 text-sm font-medium transition-colors",
    isActive ? "bg-primary-soft text-primary" : "text-muted-foreground hover:bg-muted hover:text-foreground",
  );
}

export function AppShell() {
  const logout = useAuthStore((s) => s.logout);
  const navigate = useNavigate();
  const { mode, toggleMode } = useThemeStore();
  const toggleChat = useChatStore((s) => s.toggle);
  const chatOpen = useChatStore((s) => s.open);
  const chatWidth = useChatStore((s) => s.width);

  const [isDesktop, setIsDesktop] = useState(
    () => typeof window !== "undefined" && window.matchMedia("(min-width: 768px)").matches,
  );
  useEffect(() => {
    const mq = window.matchMedia("(min-width: 768px)");
    const handler = () => setIsDesktop(mq.matches);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  function handleLogout() {
    logout();
    navigate("/");
  }

  return (
    <div className="flex h-screen overflow-hidden bg-background text-foreground">
      <aside className="hidden w-60 flex-col border-r border-border bg-sidebar md:flex">
        <div className="flex items-center gap-2 px-5 py-5">
          <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-primary text-primary-foreground">
            <Wallet size={18} />
          </div>
          <span className="text-lg font-bold">FinAgent</span>
        </div>
        <nav className="flex-1 space-y-1 px-3 py-2">
          {navItems.map(({ to, icon: Icon, label, end }) => (
            <NavLink key={to} to={to} end={end} className={navClass}>
              <Icon size={18} />
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="space-y-1 border-t border-border px-3 py-3">
          <NavLink to="/app/profile" className={navClass}>
            <User size={18} />
            Профіль
          </NavLink>
          <button onClick={handleLogout} className={cn(navClass({ isActive: false }), "w-full")}>
            <LogOut size={18} />
            Вийти
          </button>
        </div>
      </aside>

      <div
        className="flex flex-1 flex-col overflow-hidden transition-[padding] duration-200 ease-out"
        style={{ paddingRight: isDesktop && chatOpen ? chatWidth : 0 }}
      >
        <header className="flex h-14 items-center justify-between border-b border-border bg-card/60 px-4 backdrop-blur md:px-6">
          <div className="flex items-center gap-2 md:hidden">
            <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-primary text-primary-foreground">
              <Wallet size={15} />
            </div>
            <span className="font-bold">FinAgent</span>
          </div>
          <div className="ml-auto flex items-center gap-1">
            <NavLink
              to="/app/profile"
              className="flex h-9 w-9 items-center justify-center rounded-xl text-muted-foreground hover:bg-muted md:hidden"
              aria-label="Профіль"
            >
              <User size={18} />
            </NavLink>
            <button
              onClick={toggleChat}
              className="flex items-center gap-2 rounded-xl bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
            >
              <MessageSquare size={16} />
              <span className="hidden sm:inline">ШІ-чат</span>
            </button>
            <button
              onClick={toggleMode}
              className="flex h-9 w-9 items-center justify-center rounded-xl text-muted-foreground hover:bg-muted"
              aria-label="Перемкнути тему"
            >
              {mode === "dark" ? <Sun size={18} /> : <Moon size={18} />}
            </button>
            <button
              onClick={handleLogout}
              className="flex h-9 w-9 items-center justify-center rounded-xl text-muted-foreground hover:bg-muted md:hidden"
              aria-label="Вийти"
            >
              <LogOut size={18} />
            </button>
          </div>
        </header>

        <main className="flex-1 overflow-y-auto p-4 pb-24 md:p-6 md:pb-6">
          <Outlet />
        </main>
      </div>

      <nav className="fixed bottom-0 left-0 right-0 z-30 flex border-t border-border bg-card md:hidden">
        {navItems.map(({ to, icon: Icon, label, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) =>
              cn(
                "flex flex-1 flex-col items-center gap-0.5 py-2 text-[10px] font-medium",
                isActive ? "text-primary" : "text-muted-foreground",
              )
            }
          >
            <Icon size={20} />
            {label}
          </NavLink>
        ))}
      </nav>

      <ChatPanel />
    </div>
  );
}
