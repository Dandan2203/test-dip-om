import { create } from "zustand";

export const CHAT_MIN_WIDTH = 320;

const maxWidth = () =>
  typeof window === "undefined"
    ? 640
    : Math.max(CHAT_MIN_WIDTH, Math.round(window.innerWidth * 0.4));

const clampWidth = (w: number) => Math.max(CHAT_MIN_WIDTH, Math.min(maxWidth(), w));

interface ChatState {
  open: boolean;
  width: number;
  toggle: () => void;
  setOpen: (open: boolean) => void;
  setWidth: (width: number) => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  open: false,
  width: clampWidth(384),
  toggle: () => set({ open: !get().open }),
  setOpen: (open) => set({ open }),
  setWidth: (width) => set({ width: clampWidth(width) }),
}));
