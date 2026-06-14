import { AlertTriangle, RefreshCw, TrendingUp } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import api from "@/lib/api";
import type { Anomaly } from "@/types";
import { Button } from "@/components/ui/button";
import { formatMoney } from "@/lib/utils";

export function AnomaliesPage() {
  const { data, isLoading, error, refetch, isFetching } = useQuery({
    queryKey: ["anomalies"],
    queryFn: () => api.get<{ data: Anomaly[] }>("/anomalies").then((r) => r.data.data ?? []),
    retry: 1,
  });

  const anomalies = data ?? [];

  return (
    <div className="mx-auto max-w-3xl space-y-5">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-amber-100 dark:bg-amber-900/30">
            <AlertTriangle size={18} className="text-amber-600 dark:text-amber-400" />
          </div>
          <h1 className="text-xl font-semibold">Аномалії витрат</h1>
        </div>
        <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isFetching}>
          <RefreshCw size={14} className={isFetching ? "animate-spin" : ""} />
          Оновити
        </Button>
      </div>

      {isLoading && (
        <p className="py-16 text-center text-muted-foreground">Аналіз транзакцій...</p>
      )}

      {error && (
        <div className="rounded-xl border border-negative/30 bg-negative/10 px-4 py-3 text-sm text-negative">
          Не вдалося завантажити аномалії. ШІ-сервіс може бути недоступний.
        </div>
      )}

      {!isLoading && !error && anomalies.length === 0 && (
        <div className="rounded-2xl border border-border bg-card px-6 py-16 text-center">
          <AlertTriangle size={40} className="mx-auto mb-3 text-muted-foreground/40" />
          <p className="font-medium">Аномалій не виявлено</p>
          <p className="mt-1 text-sm text-muted-foreground">Ваші витрати в межах норми.</p>
        </div>
      )}

      {!isLoading &&
        anomalies.map((a) => {
          const overPct = Math.round(((a.amount - a.mean) / a.mean) * 100);
          return (
            <div
              key={a.transactionId}
              className="rounded-2xl border border-amber-200/60 dark:border-amber-800/40 bg-card p-5 shadow-sm"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1.5">
                    <span className="inline-flex items-center gap-1 rounded-full bg-amber-100 dark:bg-amber-900/30 px-2.5 py-0.5 text-xs font-medium text-amber-700 dark:text-amber-400">
                      <AlertTriangle size={11} />
                      {a.categoryName}
                    </span>
                    {overPct > 0 && (
                      <span className="inline-flex items-center gap-1 rounded-full bg-negative/10 px-2.5 py-0.5 text-xs font-medium text-negative">
                        <TrendingUp size={11} />
                        +{overPct}%
                      </span>
                    )}
                  </div>
                  <p className="text-sm font-medium">{a.description || "Без опису"}</p>
                  <p className="mt-1.5 text-sm text-muted-foreground">{a.reason}</p>
                </div>
                <div className="text-right flex-shrink-0">
                  <p className="text-lg font-semibold text-negative">{formatMoney(a.amount)}</p>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    Середня: {formatMoney(a.mean)}
                  </p>
                </div>
              </div>
            </div>
          );
        })}
    </div>
  );
}
