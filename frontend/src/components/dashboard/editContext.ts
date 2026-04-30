import { createContext, useContext } from "react";

export const DashboardEditContext = createContext(false);

export const useDashboardEdit = () => useContext(DashboardEditContext);
