import type { ReactNode } from "react";
import { StatWidget } from "@/components/widgets/StatWidget";
import { ExpensesPieWidget } from "@/components/widgets/ExpensesPieWidget";
import { TrendWidget } from "@/components/widgets/TrendWidget";
import { RecentWidget } from "@/components/widgets/RecentWidget";
import { GoalsWidget } from "@/components/widgets/GoalsWidget";
import { AnomaliesWidget } from "@/components/widgets/AnomaliesWidget";
import { TopCategoriesWidget } from "@/components/widgets/TopCategoriesWidget";
import { CurrencyWidget } from "@/components/widgets/CurrencyWidget";

export interface LayoutItem {
  i: string;
  x: number;
  y: number;
  w: number;
  h: number;
}

interface WidgetDef {
  title: string;
  render: () => ReactNode;
  w: number;
  h: number;
  minW: number;
  minH: number;
  /** Малі стат-картки фіксовані; списки/графіки можна масштабувати. За замовч. true. */
  resizable?: boolean;
}

// Сітка з 6 широких колонок: малі стат-картки = 2 колонки (фіксовані),
// списки/графіки масштабуються кроком в 1 клітинку.
export const WIDGETS: Record<string, WidgetDef> = {
  balance: { title: "Баланс", render: () => <StatWidget metric="balance" />, w: 2, h: 2, minW: 2, minH: 2, resizable: false },
  income: { title: "Доходи", render: () => <StatWidget metric="income" />, w: 2, h: 2, minW: 2, minH: 2, resizable: false },
  expense: { title: "Витрати", render: () => <StatWidget metric="expense" />, w: 2, h: 2, minW: 2, minH: 2, resizable: false },
  trend: { title: "Динаміка доходів/витрат", render: () => <TrendWidget />, w: 4, h: 5, minW: 2, minH: 3 },
  "expenses-pie": { title: "Витрати за категоріями", render: () => <ExpensesPieWidget />, w: 4, h: 4, minW: 2, minH: 3 },
  "top-categories": { title: "Топ витрат", render: () => <TopCategoriesWidget />, w: 2, h: 5, minW: 2, minH: 3 },
  recent: { title: "Останні операції", render: () => <RecentWidget />, w: 3, h: 6, minW: 2, minH: 3 },
  goals: { title: "Прогрес цілей", render: () => <GoalsWidget />, w: 3, h: 6, minW: 2, minH: 3 },
  anomalies: { title: "Аномалії витрат", render: () => <AnomaliesWidget />, w: 3, h: 6, minW: 2, minH: 3 },
  currency: { title: "Курс валют", render: () => <CurrencyWidget />, w: 2, h: 3, minW: 2, minH: 2 },
};

export const DEFAULT_LAYOUTS: Record<string, LayoutItem[]> = {
  overview: [
    { i: "balance", x: 0, y: 0, w: 2, h: 2 },
    { i: "income", x: 2, y: 0, w: 2, h: 2 },
    { i: "expense", x: 4, y: 0, w: 2, h: 2 },
    { i: "expenses-pie", x: 0, y: 2, w: 4, h: 5 },
    { i: "currency", x: 4, y: 2, w: 2, h: 3 },
    { i: "trend", x: 0, y: 7, w: 6, h: 5 },
    { i: "recent", x: 0, y: 12, w: 3, h: 6 },
    { i: "anomalies", x: 3, y: 12, w: 3, h: 6 },
  ],
  goals: [
    { i: "goals", x: 0, y: 0, w: 4, h: 8 },
    { i: "balance", x: 4, y: 0, w: 2, h: 2 },
    { i: "top-categories", x: 4, y: 2, w: 2, h: 6 },
  ],
};
