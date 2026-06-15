import { Landmark } from "lucide-react";
import { useDashboard } from "@/lib/queries";
import { WidgetFrame } from "./WidgetFrame";
import { formatMoney, formatDate, cn } from "@/lib/utils";

export function RecentWidget() {
  const { data } = useDashboard();
  const txs = data?.recentTransactions ?? [];

  return (
    <WidgetFrame title="Останні операції" to="/app/transactions">
      {txs.length === 0 ? (
        <p className="py-8 text-center text-sm text-muted-foreground">Поки немає операцій</p>
      ) : (
        <div className="space-y-1">
          {txs.map((t) => (
            <div key={t.id} className="flex items-center justify-between border-b border-border py-2 last:border-0">
              <div className="flex min-w-0 items-center gap-2">
                {t.source === "monobank" && <Landmark size={13} className="shrink-0 text-primary" />}
                <div className="min-w-0">
                  <p className="truncate text-sm" title={t.description || "Без опису"}>
                    {t.description || "Без опису"}
                  </p>
                  <p className="text-xs text-muted-foreground">{formatDate(t.transactionDate)}</p>
                </div>
              </div>
              <span className={cn("shrink-0 text-sm font-medium", t.type === "income" ? "text-positive" : "text-negative")}>
                {t.type === "income" ? "+" : "−"}{formatMoney(t.amount)}
              </span>
            </div>
          ))}
        </div>
      )}
    </WidgetFrame>
  );
}
