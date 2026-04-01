from dotenv import load_dotenv
import os

load_dotenv()

ANTHROPIC_API_KEY: str = os.environ["ANTHROPIC_API_KEY"]
ANTHROPIC_MODEL: str = os.getenv("ANTHROPIC_MODEL", "claude-haiku-4-5-20251001")
AI_SERVICE_PORT: int = int(os.getenv("AI_SERVICE_PORT", "8000"))
INTERNAL_API_TOKEN: str = os.environ["INTERNAL_API_TOKEN"]
