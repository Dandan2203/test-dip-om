import { create } from "zustand";

export type ThemeMode = "light" | "dark";
export type Accent = "indigo" | "violet" | "emerald" | "rose" | "amber" | "sky";

export const ACCENTS: { value: Accent; color: string }[] = [
  { value: "indigo", color: "#6366f1" },
  { value: "violet", color: "#8b5cf6" },
  { value: "emerald", color: "#10b981" },
  { value: "rose", color: "#f43f5e" },
  { value: "amber", color: "#f59e0b" },
  { value: "sky", color: "#0ea5e9" },
];

interface ThemeState {
  mode: ThemeMode;
  accent: Accent;
  setMode: (mode: ThemeMode) => void;
  toggleMode: () => void;
  setAccent: (accent: Accent) => void;
}

function apply(mode: ThemeMode, accent: Accent) {
  const root = document.documentElement;
  root.classList.toggle("dark", mode === "dark");
  if (accent === "indigo") root.removeAttribute("data-accent");
  else root.setAttribute("data-accent", accent);
}

const storedMode = (localStorage.getItem("theme-mode") as ThemeMode) || "light";
const storedAccent = (localStorage.getItem("theme-accent") as Accent) || "indigo";
apply(storedMode, storedAccent);

export const useThemeStore = create<ThemeState>((set, get) => ({
  mode: storedMode,
  accent: storedAccent,
  setMode: (mode) => {
    localStorage.setItem("theme-mode", mode);
    apply(mode, get().accent);
    set({ mode });
  },
  toggleMode: () => get().setMode(get().mode === "dark" ? "light" : "dark"),
  setAccent: (accent) => {
    localStorage.setItem("theme-accent", accent);
    apply(get().mode, accent);
    set({ accent });
  },
}));
