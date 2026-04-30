import { useEffect, useRef, useState, type RefObject } from "react";
import { GridLayout, useContainerWidth, verticalCompactor, type Layout } from "react-grid-layout";
import { useQuery } from "@tanstack/react-query";
import { Plus, X, GripVertical } from "lucide-react";
import api from "@/lib/api";
import { WIDGETS, DEFAULT_LAYOUTS, type LayoutItem } from "./widgetRegistry";
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { DashboardEditContext } from "./editContext";

function toItems(layout: Layout): LayoutItem[] {
  return layout.map((l) => ({ i: l.i, x: l.x, y: l.y, w: l.w, h: l.h }));
}

export function DashboardGrid({ name, editMode }: { name: string; editMode: boolean }) {
  const [layout, setLayout] = useState<LayoutItem[]>([]);
  const [loadedName, setLoadedName] = useState<string | null>(null);
  const [saveError, setSaveError] = useState("");
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const { width, containerRef, mounted } = useContainerWidth({ initialWidth: 1200 });

  const { data, isSuccess } = useQuery({
    queryKey: ["dashboard-layout", name],
    queryFn: async () => (await api.get<{ data: LayoutItem[] }>(`/dashboards/${name}`)).data.data,
  });

  // Пауза анімації під час resize.
  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    let t: ReturnType<typeof setTimeout>;
    const ro = new ResizeObserver(() => {
      el.classList.add("grid-resizing");
      clearTimeout(t);
      t = setTimeout(() => el.classList.remove("grid-resizing"), 180);
    });
    ro.observe(el);
    return () => {
      ro.disconnect();
      clearTimeout(t);
    };
  }, [containerRef]);

  if (isSuccess && loadedName !== name) {
    setLoadedName(name);
    setLayout(data && data.length > 0 ? data : DEFAULT_LAYOUTS[name] ?? []);
  }

  function save(next: LayoutItem[]) {
    if (saveTimer.current) clearTimeout(saveTimer.current);
    saveTimer.current = setTimeout(() => {
      api
        .put(`/dashboards/${name}`, { layout: next })
        .then(() => setSaveError(""))
        .catch((err: unknown) => {
          console.warn("Не вдалося зберегти дашборд", err);
          setSaveError("Не вдалося зберегти розкладку");
        });
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
      isResizable: WIDGETS[item.i].resizable !== false,
    }));

  if (mounted && width < 640) {
    const mobileLayout = (layout.length > 0 ? layout : DEFAULT_LAYOUTS[name] ?? []).filter(
      (item) => WIDGETS[item.i],
    );

    return (
      <DashboardEditContext.Provider value={false}>
        <div className="space-y-4">
          {mobileLayout.map((item) => {
            const def = WIDGETS[item.i];
            return (
              <div
                key={item.i}
                style={{ height: Math.max(140, def.h * 56) }}
              >
                {def.render()}
              </div>
            );
          })}
        </div>
      </DashboardEditContext.Provider>
    );
  }

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
      {editMode && saveError && (
        <p className="mb-3 text-sm text-negative">{saveError}</p>
      )}

      <DashboardEditContext.Provider value={editMode}>
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
      </DashboardEditContext.Provider>
    </div>
  );
}
