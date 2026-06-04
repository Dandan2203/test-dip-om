import os
import sys
from pathlib import Path

os.environ.setdefault("ANTHROPIC_API_KEY", "test-key")
os.environ.setdefault("INTERNAL_API_TOKEN", "test-token")

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))
