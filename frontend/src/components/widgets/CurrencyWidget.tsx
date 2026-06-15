import { useCurrencyRates } from "@/lib/queries";
import { WidgetFrame } from "./WidgetFrame";

// Валюти, які показуємо проти гривні (currencyCodeB === 980).
const CCY: Record<number, string> = {
  840: "USD",
  978: "EUR",
  826: "GBP",
  985: "PLN",
};

export function CurrencyWidget() {
  const { data: rates, isLoading } = useCurrencyRates();

  const list = (rates ?? [])
    .filter((r) => r.currencyCodeB === 980 && CCY[r.currencyCodeA])
    .sort((a, b) => a.currencyCodeA - b.currencyCodeA);

  return (
    <WidgetFrame title="Курс валют (Monobank)">
      {isLoading ? (
        <p className="py-8 text-center text-sm text-muted-foreground">Завантаження…</p>
      ) : list.length === 0 ? (
        <p className="py-8 text-center text-sm text-muted-foreground">Курси недоступні</p>
      ) : (
        <div className="space-y-1.5">
          {/* Колонки: валюта — гнучка (truncate), числа — за вмістом і без переносу.
              Так дані не накладаються навіть на вузькому віджеті. */}
          <div className="grid grid-cols-[minmax(0,1fr)_auto_auto] gap-3 px-2 text-xs font-medium text-muted-foreground">
            <span className="truncate">Валюта</span>
            <span className="text-right">Купівля</span>
            <span className="text-right">Продаж</span>
          </div>
          {list.map((r) => {
            const buy = r.rateBuy || r.rateCross;
            const sell = r.rateSell || r.rateCross;
            return (
              <div
                key={r.currencyCodeA}
                className="grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-3 rounded-xl bg-muted/40 px-2 py-2 text-sm"
              >
                <span className="truncate font-medium">
                  {CCY[r.currencyCodeA]}
                  <span className="text-xs text-muted-foreground">/UAH</span>
                </span>
                <span className="whitespace-nowrap text-right tabular-nums text-positive">
                  {buy ? buy.toFixed(2) : "—"}
                </span>
                <span className="whitespace-nowrap text-right tabular-nums text-negative">
                  {sell ? sell.toFixed(2) : "—"}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </WidgetFrame>
  );
}
