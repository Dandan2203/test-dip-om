import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

const UAH = new Intl.NumberFormat("uk-UA", {
  style: "currency",
  currency: "UAH",
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});

export function formatMoney(value: number): string {
  return UAH.format(value);
}

export function formatDate(value: string | Date): string {
  const d = typeof value === "string" ? new Date(value) : value;
  return d.toLocaleDateString("uk-UA", { day: "2-digit", month: "2-digit", year: "numeric" });
}

export function tipMoney(value: unknown): string {
  return formatMoney(Number(value));
}

export const CHART_COLORS = [
  "#6366f1", "#10b981", "#f59e0b", "#f43f5e",
  "#0ea5e9", "#8b5cf6", "#14b8a6", "#fb923c",
  "#ec4899", "#84cc16",
];
