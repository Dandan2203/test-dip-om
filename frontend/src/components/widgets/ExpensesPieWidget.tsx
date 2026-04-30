import { useState } from "react";
import { PieChart, Pie, Cell, ResponsiveContainer } from "recharts";
import { useExpenseByCategory, useCategories } from "@/lib/queries";
import { WidgetFrame } from "./WidgetFrame";
import { CHART_COLORS, formatMoney, cn } from "@/lib/utils";

export function ExpensesPieWidget() {
  const { data: stats } = useExpenseByCategory();
  const { data: categories } = useCategories();
  const [hovered, setHovered] = useState<number | null>(null);
  const [selected, setSelected] = useState<number | null>(null);

  const nameById = new Map((categories ?? []).map((c) => [c.id, c.name]));
  const raw = (stats ?? []).map((s) => ({
    name: s.categoryId ? nameById.get(s.categoryId) ?? "Інше" : "Без категорії",
    value: s.total,
  }));
  const total = raw.reduce((acc, d) => acc + d.value, 0);
  const data = raw
    .map((d) => ({ ...d, pct: total > 0 ? (d.value / total) * 100 : 0 }))
    .sort((a, b) => b.value - a.value);

  const active = hovered !== null ? hovered : selected;

  if (data.length === 0) {
    return (
      <WidgetFrame title="Витрати за категоріями">
        <p className="py-8 text-center text-sm text-muted-foreground">Немає витрат за період</p>
      </WidgetFrame>
    );
  }

  const center = active !== null ? data[active] : null;

  function toggle(i: number) {
    setSelected((prev) => (prev === i ? null : i));
  }

  return (
    <WidgetFrame title="Витрати за категоріями">
      <div className="flex h-full gap-4">
        <div className="min-h-0 flex-1 space-y-0.5 overflow-auto">
          {data.map((d, i) => {
            const isOn = active === i;
            return (
              <button
                key={d.name}
                onMouseEnter={() => setHovered(i)}
                onMouseLeave={() => setHovered(null)}
                onClick={() => toggle(i)}
                className={cn(
                  "flex w-full items-center gap-2.5 rounded-lg px-2 py-2 text-left text-[15px] outline-none transition-colors focus:outline-none focus-visible:outline-none active:bg-muted",
                  selected === i ? "bg-muted" : isOn ? "bg-muted/60" : "hover:bg-muted/60",
                )}
              >
                <span
                  className="h-3 w-3 shrink-0 rounded-full transition-all"
                  style={{
                    background: CHART_COLORS[i % CHART_COLORS.length],
                    boxShadow: isOn ? `0 0 0 4px ${CHART_COLORS[i % CHART_COLORS.length]}33` : undefined,
                  }}
                />
                <span className="flex-1 truncate text-foreground" title={d.name}>
                  {d.name}
                </span>
                <span className="shrink-0 tabular-nums font-semibold text-muted-foreground">
                  {d.pct.toFixed(1)}%
                </span>
              </button>
            );
          })}
        </div>

        <div className="relative aspect-square w-[42%] min-w-[120px] max-w-[300px] shrink-0 self-center outline-none [&_*]:outline-none [&_*:focus]:outline-none">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={data}
                dataKey="value"
                nameKey="name"
                innerRadius="64%"
                outerRadius="100%"
                paddingAngle={1.5}
                stroke="none"
                isAnimationActive={false}
                tabIndex={-1}
                onMouseEnter={(_, i) => setHovered(i)}
                onMouseLeave={() => setHovered(null)}
                onClick={(_, i) => toggle(i)}
              >
                {data.map((_, i) => (
                  <Cell
                    key={i}
                    tabIndex={-1}
                    fill={CHART_COLORS[i % CHART_COLORS.length]}
                    style={{
                      cursor: "pointer",
                      outline: "none",
                      opacity: active === null || active === i ? 1 : 0.3,
                      transition: "opacity 0.2s",
                    }}
                  />
                ))}
              </Pie>
            </PieChart>
          </ResponsiveContainer>
          <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center px-3 text-center">
            {center ? (
              <>
                <span className="text-3xl font-bold leading-none">{center.pct.toFixed(1)}%</span>
                <span className="mt-1 line-clamp-2 text-xs leading-tight text-muted-foreground">
                  {center.name}
                </span>
              </>
            ) : (
              <>
                <span className="text-xl font-bold leading-none">{formatMoney(total)}</span>
                <span className="mt-1 text-xs text-muted-foreground">усього</span>
              </>
            )}
          </div>
        </div>
      </div>
    </WidgetFrame>
  );
}
