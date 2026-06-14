import { create } from "zustand";

export const CHAT_MIN_WIDTH = 320;
export const CHAT_MAX_WIDTH = 640;

const clampWidth = (w: number) => Math.max(CHAT_MIN_WIDTH, Math.min(CHAT_MAX_WIDTH, w));

interface ChatState {
  open: boolean;
  width: number;
  toggle: () => void;
  setOpen: (open: boolean) => void;
  setWidth: (width: number) => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  open: false,
  width: 384,
  toggle: () => set({ open: !get().open }),
  setOpen: (open) => set({ open }),
  setWidth: (width) => set({ width: clampWidth(width) }),
}));
