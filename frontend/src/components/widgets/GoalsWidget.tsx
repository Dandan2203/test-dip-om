import { useGoals } from "@/lib/queries";
import { WidgetFrame } from "./WidgetFrame";
import { formatMoney } from "@/lib/utils";

export function GoalsWidget() {
  const { data: goals } = useGoals();

  return (
    <WidgetFrame title="Прогрес цілей">
      {!goals || goals.length === 0 ? (
        <p className="py-8 text-center text-sm text-muted-foreground">Цілей ще немає</p>
      ) : (
        <div className="space-y-4">
          {goals.map((g) => {
            const pct = g.targetAmount > 0 ? Math.min(100, (g.currentAmount / g.targetAmount) * 100) : 0;
            return (
              <div key={g.id}>
                <div className="mb-1 flex justify-between text-sm">
                  <span className="truncate" title={g.title}>{g.title}</span>
                  <span className="shrink-0 text-muted-foreground">
                    {formatMoney(g.currentAmount)} / {formatMoney(g.targetAmount)}
                  </span>
                </div>
                <div className="h-2 overflow-hidden rounded-full bg-muted">
                  <div className="h-full rounded-full bg-primary transition-all" style={{ width: `${pct}%` }} />
                </div>
              </div>
            );
          })}
        </div>
      )}
    </WidgetFrame>
  );
}
