from pydantic import BaseModel
from typing import Literal


class CategoryItem(BaseModel):
    id: int
    name: str
    type: Literal["income", "expense"]


class TransactionItem(BaseModel):
    id: int
    category_id: int | None
    category_name: str | None = None
    type: Literal["income", "expense"]
    amount: float
    description: str
    transaction_date: str


class GoalItem(BaseModel):
    id: int
    title: str
    target_amount: float
    current_amount: float
    deadline: str | None


# --- /categorize ---

class CategorizeRequest(BaseModel):
    description: str
    transaction_type: Literal["income", "expense"]
    available_categories: list[CategoryItem]


class CategorizeResponse(BaseModel):
    category_id: int | None
    confidence: float
    category_name: str | None


# --- /chat ---

class ChatRequest(BaseModel):
    message: str
    user_id: int
    transactions: list[TransactionItem]
    goals: list[GoalItem]
    categories: list[CategoryItem]


class ActionData(BaseModel):
    action_type: str | None = None
    amount: float | None = None
    transaction_type: Literal["income", "expense"] | None = None
    description: str | None = None
    category_name: str | None = None
    transaction_date: str | None = None
    transaction_id: int | None = None
    title: str | None = None
    target_amount: float | None = None
    deadline: str | None = None
    goal_id: int | None = None


class ChatResponse(BaseModel):
    response: str
    intent: str
    action: ActionData | None = None


# --- /audit ---

class AuditRequest(BaseModel):
    user_id: int
    transactions: list[TransactionItem]


class AnomalyItem(BaseModel):
    transaction_id: int
    description: str
    category_name: str
    amount: float
    mean: float
    reason: str


class AuditResponse(BaseModel):
    anomalies: list[AnomalyItem]
