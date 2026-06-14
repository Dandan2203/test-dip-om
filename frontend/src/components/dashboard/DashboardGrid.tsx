import { useEffect, useRef, useState, type RefObject } from "react";
import { GridLayout, useContainerWidth, verticalCompactor, type Layout } from "react-grid-layout";
import { useQuery } from "@tanstack/react-query";
import { Plus, X, GripVertical } from "lucide-react";
import api from "@/lib/api";
import { WIDGETS, DEFAULT_LAYOUTS, type LayoutItem } from "./widgetRegistry";
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";

function toItems(layout: Layout): LayoutItem[] {
  return layout.map((l) => ({ i: l.i, x: l.x, y: l.y, w: l.w, h: l.h }));
}

export function DashboardGrid({ name, editMode }: { name: string; editMode: boolean }) {
  const [layout, setLayout] = useState<LayoutItem[]>([]);
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const { width, containerRef, mounted } = useContainerWidth({ initialWidth: 1200 });

  const { data, isSuccess } = useQuery({
    queryKey: ["dashboard-layout", name],
    queryFn: async () => (await api.get<{ data: LayoutItem[] }>(`/dashboards/${name}`)).data.data,
  });

  useEffect(() => {
    if (!isSuccess) return;
    setLayout(data && data.length > 0 ? data : DEFAULT_LAYOUTS[name] ?? []);
  }, [isSuccess, data, name]);

  function save(next: LayoutItem[]) {
    if (saveTimer.current) clearTimeout(saveTimer.current);
    saveTimer.current = setTimeout(() => {
      api.put(`/dashboards/${name}`, { layout: next }).catch(() => {});
    }, 600);
  }

  function handleLayoutChange(next: Layout) {
    const items = toItems(next);
    setLayout(items);
    if (editMode) save(items);
  }

  function addWidget(type: string) {
    if (layout.some((l) => l.i === type)) return;
    const def = WIDGETS[type];
    const next = [...layout, { i: type, x: 0, y: 999, w: def.w, h: def.h }];
    setLayout(next);
    save(next);
  }

  function removeWidget(type: string) {
    const next = layout.filter((l) => l.i !== type);
    setLayout(next);
    save(next);
  }

  const available = Object.keys(WIDGETS).filter((t) => !layout.some((l) => l.i === t));

  const gridLayout: Layout = layout
    .filter((item) => WIDGETS[item.i])
    .map((item) => ({
      ...item,
      minW: WIDGETS[item.i].minW,
      minH: WIDGETS[item.i].minH,
      // Малі стат-картки не масштабуються; списки/графіки — так.
      isResizable: WIDGETS[item.i].resizable !== false,
    }));

  return (
    <div>
      {editMode && (
        <div className="mb-3 flex items-center justify-between rounded-xl border border-dashed border-primary/40 bg-primary-soft px-4 py-2">
          <span className="text-sm text-primary">Режим редагування: перетягуйте та змінюйте розмір віджетів</span>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button size="sm" variant="soft" disabled={available.length === 0}>
                <Plus size={15} /> Додати віджет
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {available.map((t) => (
                <DropdownMenuItem key={t} onSelect={() => addWidget(t)}>
                  {WIDGETS[t].title}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )}

      <div ref={containerRef as RefObject<HTMLDivElement>}>
        {mounted && (
          <GridLayout
            width={width}
            layout={gridLayout}
            gridConfig={{ cols: 6, rowHeight: 56, margin: [16, 16] as const }}
            dragConfig={{ enabled: editMode, handle: ".widget-drag" }}
            resizeConfig={{ enabled: editMode }}
            compactor={verticalCompactor}
            onLayoutChange={handleLayoutChange}
          >
        {gridLayout.map((item) => {
          const def = WIDGETS[item.i];
          if (!def) return <div key={item.i} />;
          return (
            <div key={item.i} className="relative">
              {editMode && (
                <>
                  <div className="widget-drag absolute left-0 right-0 top-0 z-10 flex h-7 cursor-move items-center justify-center rounded-t-2xl bg-primary-soft text-primary">
                    <GripVertical size={14} />
                  </div>
                  <button
                    onClick={() => removeWidget(item.i)}
                    className="absolute right-1.5 top-1.5 z-20 flex h-5 w-5 items-center justify-center rounded-md bg-negative text-white"
                    aria-label="Прибрати віджет"
                  >
                    <X size={12} />
                  </button>
                </>
              )}
              <div className={editMode ? "h-full pt-7" : "h-full"}>{def.render()}</div>
            </div>
          );
        })}
          </GridLayout>
        )}
      </div>
    </div>
  );
}
