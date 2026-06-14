import { LineChart, Line, XAxis, YAxis, CartesianGrid, ResponsiveContainer, Tooltip, Legend } from "recharts";
import { useTransactions, aggregateByDay } from "@/lib/queries";
import { WidgetFrame } from "./WidgetFrame";
import { tipMoney } from "@/lib/utils";

export function TrendWidget() {
  const { data } = useTransactions(200);
  const points = aggregateByDay(data?.items ?? [], 30).map((p) => ({
    ...p,
    label: p.date.slice(5),
  }));

  return (
    <WidgetFrame title="Динаміка (доходи/витрати)">
      {points.length === 0 ? (
        <p className="py-8 text-center text-sm text-muted-foreground">Недостатньо даних</p>
      ) : (
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={points} margin={{ top: 8, right: 8, left: -10, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
            <XAxis dataKey="label" tick={{ fontSize: 11, fill: "var(--muted-foreground)" }} />
            <YAxis tick={{ fontSize: 11, fill: "var(--muted-foreground)" }} width={48} />
            <Tooltip
              formatter={tipMoney}
              contentStyle={{ background: "var(--card)", border: "1px solid var(--border)", borderRadius: 12, color: "var(--foreground)" }}
            />
            <Legend wrapperStyle={{ fontSize: 12 }} />
            <Line type="monotone" dataKey="income" name="Доходи" stroke="var(--positive)" strokeWidth={2} dot={false} />
            <Line type="monotone" dataKey="expense" name="Витрати" stroke="var(--negative)" strokeWidth={2} dot={false} />
          </LineChart>
        </ResponsiveContainer>
      )}
    </WidgetFrame>
  );
}
