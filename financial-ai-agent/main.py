import logging
from fastapi import FastAPI, HTTPException, Header, Depends

from config import AI_SERVICE_PORT, INTERNAL_API_TOKEN
from models import (
    CategorizeRequest, CategorizeResponse,
    ChatRequest, ChatResponse,
    AuditRequest, AuditResponse,
)
from agents.categorizer import categorize
from agents.chat import run_chat
from agents.auditor import audit

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="FinAgent AI Service", version="1.0.0")


def require_internal(x_internal_token: str = Header(default="")):
    # Приватний endpoint.
    if x_internal_token != INTERNAL_API_TOKEN:
        raise HTTPException(status_code=401, detail="unauthorized")


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/categorize", response_model=CategorizeResponse, dependencies=[Depends(require_internal)])
def categorize_endpoint(req: CategorizeRequest):
    try:
        return categorize(req)
    except Exception as e:
        logger.error("categorize error: %s", e)
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/chat", response_model=ChatResponse, dependencies=[Depends(require_internal)])
def chat_endpoint(req: ChatRequest):
    try:
        return run_chat(req)
    except Exception as e:
        logger.error("chat error: %s", e)
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/audit", response_model=AuditResponse, dependencies=[Depends(require_internal)])
def audit_endpoint(req: AuditRequest):
    try:
        return audit(req)
    except Exception as e:
        logger.error("audit error: %s", e)
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=AI_SERVICE_PORT, reload=True)
