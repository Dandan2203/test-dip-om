import { TrendingUp, TrendingDown, Wallet } from "lucide-react";
import { useDashboard } from "@/lib/queries";
import { WidgetFrame } from "./WidgetFrame";
import { formatMoney, cn } from "@/lib/utils";

type Metric = "balance" | "income" | "expense";

const meta: Record<Metric, { label: string; icon: typeof Wallet; color: string }> = {
  balance: { label: "Баланс (місяць)", icon: Wallet, color: "text-primary" },
  income: { label: "Доходи (місяць)", icon: TrendingUp, color: "text-positive" },
  expense: { label: "Витрати (місяць)", icon: TrendingDown, color: "text-negative" },
};

export function StatWidget({ metric }: { metric: Metric }) {
  const { data, isLoading } = useDashboard();
  const { label, icon: Icon, color } = meta[metric];

  return (
    <WidgetFrame className="justify-center">
      <div className="flex h-full flex-col justify-center">
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">{label}</span>
          <Icon size={18} className={color} />
        </div>
        <p className={cn("mt-2 text-2xl font-bold", color)}>
          {isLoading ? "…" : formatMoney(data?.[metric] ?? 0)}
        </p>
      </div>
    </WidgetFrame>
  );
}
