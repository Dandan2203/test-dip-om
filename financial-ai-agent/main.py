import logging
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware

from config import AI_SERVICE_PORT
from models import (
    CategorizeRequest, CategorizeResponse,
    ChatRequest, ChatResponse,
    AuditRequest, AuditResponse,
)
from agents.categorizer import categorize
from agents.chat_graph import chat_graph
from agents.auditor import audit

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="FinAgent AI Service", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/categorize", response_model=CategorizeResponse)
def categorize_endpoint(req: CategorizeRequest):
    try:
        return categorize(req)
    except Exception as e:
        logger.error("categorize error: %s", e)
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/chat", response_model=ChatResponse)
def chat_endpoint(req: ChatRequest):
    try:
        initial_state = {
            "message": req.message,
            "transactions": [t.model_dump() for t in req.transactions],
            "goals": [g.model_dump() for g in req.goals],
            "categories": [c.model_dump() for c in req.categories],
            "intent": "",
            "response": "",
            "action": None,
        }
        result = chat_graph.invoke(initial_state)
        from models import ActionData
        action = None
        if result.get("action"):
            action = ActionData(**result["action"])
        return ChatResponse(response=result["response"], intent=result["intent"], action=action)
    except Exception as e:
        logger.error("chat error: %s", e)
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/audit", response_model=AuditResponse)
def audit_endpoint(req: AuditRequest):
    try:
        return audit(req)
    except Exception as e:
        logger.error("audit error: %s", e)
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=AI_SERVICE_PORT, reload=True)
