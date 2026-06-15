# FinAgent

Сервіс контролю особистих фінансів із вбудованим ШІ-асистентом: користувач веде транзакції, цілі та бюджет вручну або керує ними природною мовою через чат. ШІ також авто-категоризує транзакції та шукає аномалії у витратах.

## Стек
- **Backend** — Go (Clean Architecture: domain → usecase → repository → delivery), Gin, PostgreSQL, JWT, AES-256-GCM.
- **AI-сервіс** — Python (FastAPI), Anthropic Claude Haiku 4.5, нативний tool-use.
- **Frontend** — React + TypeScript + Vite.
- **Інфраструктура** — Docker Compose, інтеграція Monobank.

Деталі — [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md). Метрики — [docs/METRICS.md](docs/METRICS.md).

## Запуск (Docker)
1. `cp .env.example .env`
2. Заповніть у `.env` усі обов'язкові змінні:
   - `JWT_SECRET` — довгий випадковий рядок
   - `ANTHROPIC_API_KEY` — ключ Anthropic
   - `ENCRYPTION_KEY` — довгий випадковий рядок (зміна робить наявні Mono-токени нечитними)
   - `INTERNAL_API_TOKEN` — спільний секрет бекенда й AI-сервісу (будь-який довгий рядок)
3. `docker compose up -d --build`
4. Фронтенд — http://localhost, API — http://localhost:8080

AI-сервіс назовні не публікується — доступний лише бекенду по внутрішній мережі зі спільним секретом `INTERNAL_API_TOKEN`.

## Локальний запуск (без Docker)
- PostgreSQL на `:5433`.
- **Backend**: у `backend/.env` — `DATABASE_URL`, `JWT_SECRET`, `ENCRYPTION_KEY`, `AI_INTERNAL_TOKEN`; `go run ./cmd`.
- **AI**: у `financial-ai-agent/.env` — `ANTHROPIC_API_KEY`, `INTERNAL_API_TOKEN`; `uvicorn main:app`.
- **Frontend**: `cd frontend && npm install && npm run dev`.

`AI_INTERNAL_TOKEN` (backend) і `INTERNAL_API_TOKEN` (AI-сервіс) **мають збігатися**.

## Тести та CI
- **Backend (Go)**: `cd backend && go test ./...` — юзкейси, AES-крипто, JWT, retry-клієнт AI, rate-limit, зіставлення цілей.
- **AI-сервіс (Python)**: `cd financial-ai-agent && pytest tests/` — read-інструменти, конвертер дій, детектор аномалій (без мережі).
- **Frontend**: `cd frontend && npm run lint && npm run build`.
- **CI**: [.github/workflows/ci.yml](.github/workflows/ci.yml) — три паралельні джоби (Go build+vet+test -race, Python pytest, React lint+build) на кожен push/PR.

## Оцінка якості (eval)
Піднятий AI-сервіс + `python eval/run_eval.py` — міряє точність інтентів і захоплення дій. Див. [eval/README.md](eval/README.md).

## Безпека й комплаєнс (стисло)
- ШІ-чат не дає персональних інвестиційних порад (guardrail у системному промті) і додає дисклеймер до фінансових порад.
- Будь-яка зміна даних через чат — лише після підтвердження користувача (human-in-the-loop), із логуванням для undo.
- AI-сервіс приватний (shared-secret, порт закрито). Секрети обов'язкові (fail-fast). Mono-токени — AES-256-GCM.
- Деталі — [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) §Безпека.
