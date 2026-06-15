import { createContext, useContext } from "react";

// Чи дашборд у режимі редагування. Споживає WidgetFrame, щоб у режимі
// перетягування не робити заголовки-посилання активними (уникаємо випадкових переходів).
export const DashboardEditContext = createContext(false);

export const useDashboardEdit = () => useContext(DashboardEditContext);
