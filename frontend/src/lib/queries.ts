import { useQuery } from "@tanstack/react-query";
import api from "@/lib/api";
import type { Anomaly, Category, CurrencyRate, DashboardData, Goal, Transaction, TransactionList } from "@/types";

export interface CategoryStat {
  categoryId: number | null;
  total: number;
  count: number;
}

export function useDashboard() {
  return useQuery({
    queryKey: ["dashboard"],
    queryFn: async () => (await api.get<{ data: DashboardData }>("/dashboard")).data.data,
  });
}

export function useCategories() {
  return useQuery({
    queryKey: ["categories"],
    queryFn: async () => (await api.get<{ data: Category[] }>("/categories")).data.data,
  });
}

export function useGoals() {
  return useQuery({
    queryKey: ["goals"],
    queryFn: async () => (await api.get<{ data: Goal[] }>("/goals")).data.data,
  });
}

export function useExpenseByCategory() {
  return useQuery({
    queryKey: ["by-category", "expense"],
    queryFn: async () =>
      (await api.get<{ data: CategoryStat[] }>("/stats/by-category?type=expense")).data.data,
  });
}

export function useCurrencyRates() {
  return useQuery({
    queryKey: ["mono-currency"],
    queryFn: async () =>
      (await api.get<{ data: CurrencyRate[] }>("/mono/currency")).data.data ?? [],
    retry: false,
    staleTime: 30 * 60 * 1000,
  });
}

export function useAnomalies() {
  return useQuery({
    queryKey: ["anomalies"],
    queryFn: async () => (await api.get<{ data: Anomaly[] }>("/anomalies")).data.data ?? [],
    retry: false,
  });
}

export function useTransactions(limit = 200) {
  return useQuery({
    queryKey: ["transactions", limit],
    queryFn: async () =>
      (await api.get<{ data: TransactionList }>(`/transactions?limit=${limit}`)).data.data,
  });
}

export interface DayPoint {
  date: string;
  income: number;
  expense: number;
}

// Агрегує транзакції за днями для лінійного графіка динаміки.
export function aggregateByDay(txs: Transaction[], days = 30): DayPoint[] {
  const map = new Map<string, DayPoint>();
  for (const t of txs) {
    const day = t.transactionDate.slice(0, 10);
    const point = map.get(day) ?? { date: day, income: 0, expense: 0 };
    if (t.type === "income") point.income += t.amount;
    else point.expense += t.amount;
    map.set(day, point);
  }
  return Array.from(map.values())
    .sort((a, b) => a.date.localeCompare(b.date))
    .slice(-days);
}
