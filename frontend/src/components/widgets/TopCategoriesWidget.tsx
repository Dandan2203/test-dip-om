import { BarChart, Bar, XAxis, YAxis, ResponsiveContainer, Tooltip, Cell } from "recharts";
import { useExpenseByCategory, useCategories } from "@/lib/queries";
import { WidgetFrame } from "./WidgetFrame";
import { tipMoney, CHART_COLORS } from "@/lib/utils";

export function TopCategoriesWidget() {
  const { data: stats } = useExpenseByCategory();
  const { data: categories } = useCategories();

  const nameById = new Map((categories ?? []).map((c) => [c.id, c.name]));
  const data = (stats ?? [])
    .map((s) => ({ name: s.categoryId ? nameById.get(s.categoryId) ?? "Інше" : "Без категорії", total: s.total }))
    .slice(0, 6);

  return (
    <WidgetFrame title="Топ витрат">
      {data.length === 0 ? (
        <p className="py-8 text-center text-sm text-muted-foreground">Немає даних</p>
      ) : (
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={data} layout="vertical" margin={{ top: 4, right: 8, left: 8, bottom: 0 }}>
            <XAxis type="number" hide />
            <YAxis type="category" dataKey="name" width={90} tick={{ fontSize: 11, fill: "var(--muted-foreground)" }} />
            <Tooltip
              cursor={{ fill: "var(--muted)" }}
              formatter={tipMoney}
              contentStyle={{ background: "var(--card)", border: "1px solid var(--border)", borderRadius: 12, color: "var(--foreground)" }}
            />
            <Bar dataKey="total" radius={[0, 6, 6, 0]}>
              {data.map((_, i) => (
                <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      )}
    </WidgetFrame>
  );
}
