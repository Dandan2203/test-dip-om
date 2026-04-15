export interface User {
  id: number;
  email: string;
  name: string;
}

export type TxType = "income" | "expense";

export interface Category {
  id: number;
  name: string;
  type: TxType;
  isSystem: boolean;
}

export interface Transaction {
  id: number;
  categoryId: number | null;
  type: TxType;
  amount: number;
  description: string;
  transactionDate: string;
  isAiCategorized: boolean;
  createdAt: string;
  source: string;
  currencyCode: number;
  originalAmount: number | null;
}

export interface TransactionList {
  items: Transaction[];
  total: number;
}

export interface Goal {
  id: number;
  title: string;
  targetAmount: number;
  currentAmount: number;
  deadline: string | null;
  createdAt: string;
}

export interface DashboardData {
  income: number;
  expense: number;
  balance: number;
  recentTransactions: Transaction[];
  goals: Goal[];
}

export interface ChatMessage {
  role: "user" | "assistant";
  content: string;
  intent?: string;
  action?: ChatAction | null;
  pendingAction?: PendingAction | null;
  pendingStatus?: "pending" | "executing" | "done" | "cancelled";
  pendingError?: string;
  timestamp: string;
}

export interface ChatAction {
  actionType: "create" | "delete" | "contribute";
  entityType: "transaction" | "goal";
  entityId: number;
}

// Підтвердження AI-дії
export interface PendingAction {
  actionType: string;
  entityType: string;
  amount?: number;
  transactionType?: string;
  description?: string;
  categoryId?: number | null;
  categoryName?: string;
  transactionDate?: string;
  transactionId?: number;
  goalId?: number;
  goalTitle?: string;
  targetAmount?: number;
  deadline?: string | null;
  confirmText: string;
  requiresConfirm: boolean;
}

export interface Anomaly {
  transactionId: number;
  description: string;
  categoryName: string;
  amount: number;
  mean: number;
  reason: string;
}

export interface MonoAccount {
  id: string;
  type: string;
  balance: number;
  currencyCode: number;
  maskedPan: string;
  iban: string;
}

export interface CurrencyRate {
  currencyCodeA: number;
  currencyCodeB: number;
  rateSell: number;
  rateBuy: number;
  rateCross: number;
}
