"""Спільна підготовка для тестів.

Модулі сервісу читають секрети з оточення на імпорті (config.py) і створюють
anthropic.Anthropic(...) на рівні модуля. Конструктор клієнта не ходить у мережу,
тож достатньо підставити фіктивні значення ще до перших імпортів.
"""
import os
import sys
from pathlib import Path

os.environ.setdefault("ANTHROPIC_API_KEY", "test-key")
os.environ.setdefault("INTERNAL_API_TOKEN", "test-token")

# Корінь сервісу у sys.path, щоб імпортувати config/models/agents як у проді.
sys.path.insert(0, str(Path(__file__).resolve().parent.parent))
