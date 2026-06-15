import { AlertTriangle } from "lucide-react";
import { useAnomalies } from "@/lib/queries";
import { WidgetFrame } from "./WidgetFrame";
import { formatMoney } from "@/lib/utils";

export function AnomaliesWidget() {
  const { data: anomalies, isError } = useAnomalies();

  return (
    <WidgetFrame title="Аномалії витрат" to="/app/anomalies">
      {isError ? (
        <p className="py-8 text-center text-sm text-muted-foreground">ШІ-сервіс недоступний</p>
      ) : !anomalies || anomalies.length === 0 ? (
        <p className="py-8 text-center text-sm text-muted-foreground">Аномалій не виявлено</p>
      ) : (
        <div className="space-y-2">
          {anomalies.map((a) => (
            <div key={a.transactionId} className="rounded-xl border border-amber-500/30 bg-amber-500/5 p-3">
              <div className="flex items-center justify-between">
                <span className="flex items-center gap-1 text-xs font-medium text-amber-600 dark:text-amber-400">
                  <AlertTriangle size={12} /> {a.categoryName}
                </span>
                <span className="text-sm font-semibold text-negative">{formatMoney(a.amount)}</span>
              </div>
              <p className="mt-1 text-xs text-muted-foreground">{a.reason}</p>
            </div>
          ))}
        </div>
      )}
    </WidgetFrame>
  );
}
