# Runbook FinAgent

Операційні процедури для локального запуску, діагностики та демонстрації.

---

## Запуск і зупинка

```bash
# Старт усіх сервісів
docker compose up -d --build

# Переглянути логи
docker compose logs -f backend
docker compose logs -f ai
docker compose logs -f frontend

# Зупинити
docker compose down

# Зупинити + видалити volumes (скидає БД)
docker compose down -v
```

## Перевірка стану

```bash
# Health-check бекенду
curl http://localhost:8080/health
# → {"status":"ok","db":"up"}

# AI-сервіс (внутрішній, потребує токена)
curl -H "X-Internal-Token: $INTERNAL_API_TOKEN" http://localhost:8000/health
# → {"status":"ok"}  (якщо ендпоінт реалізований)
# Або перевірити через docker compose exec:
docker compose exec ai python -c "import anthropic; print('ok')"
```

## Обертання секретів

> Після зміни секретів необхідно перестартувати відповідний сервіс.

| Секрет | Що змінити | Наслідок |
|---|---|---|
| `JWT_SECRET` | Новий рядок у `.env`, рестарт бекенду | Усі активні JWT стають недійсними — користувачі перелогінюються |
| `ANTHROPIC_API_KEY` | Новий ключ у `.env`, рестарт `ai` | Без downtime для бекенду |
| `ENCRYPTION_KEY` | **Небезпечно**: наявні Mono-токени в БД стають нечитними | Перед зміною — від'єднати всі Mono-акаунти |
| `INTERNAL_API_TOKEN` | Оновити **обидва**: `backend` і `ai` в `.env`, рестарт обох | |

```bash
# Рестарт одного сервісу без перебудови
docker compose restart backend
docker compose restart ai
```

## Якщо AI-сервіс недоступний

Бекенд повертає `{"code":"AI_UNAVAILABLE"}` — фронт показує «ШІ-сервіс зараз недоступний».

Перевірити:
```bash
docker compose ps ai          # статус: running?
docker compose logs --tail=50 ai
```

Типові причини:
- `ANTHROPIC_API_KEY` невалідний → Claude повертає 401; у логах `authentication_error`
- `INTERNAL_API_TOKEN` не збігається → бекенд отримує 401 від AI-сервісу; у логах бекенду `ai: status 401`
- AI-сервіс не встиг стартувати → `docker compose restart ai`

## Якщо бекенд не стартує

```bash
docker compose logs backend | grep "FATAL\|required"
```

Типові причини:
- Не задані обов'язкові env-змінні (`JWT_SECRET`, `ENCRYPTION_KEY`, `AI_INTERNAL_TOKEN`, `DATABASE_URL`) → повідомлення `required env var … is not set`
- БД не готова → `connection refused`; бекенд має вбудований retry; якщо не стартував через хвилину — `docker compose restart backend`

## Перегляд чат-логів (БД)

```bash
docker compose exec db psql -U finagent_user -d finagent -c \
  "SELECT id, user_id, intent, left(response,60), created_at FROM ai_chat_logs ORDER BY created_at DESC LIMIT 20;"
```

Очистити логи (ретенція 90 днів):
```sql
DELETE FROM ai_chat_logs WHERE created_at < now() - interval '90 days';
```

## Запуск тестів (офлайн, без мережі)

```bash
# Backend (Go): юзкейси, крипто, JWT, retry-клієнт, rate-limit
cd backend && go test ./... -race

# AI-сервіс (Python): інструменти, конвертер дій, детектор аномалій
cd financial-ai-agent && pip install -r requirements.txt pytest && pytest tests/

# Frontend: статичний аналіз і збірка
cd frontend && npm ci && npm run lint && npm run build
```

Ті самі кроки виконує CI ([.github/workflows/ci.yml](../.github/workflows/ci.yml)) на кожен push/PR.

## Запуск eval-харнесу (потребує живого AI-сервісу)

```bash
# AI-сервіс має бути запущений
export INTERNAL_API_TOKEN=$(grep INTERNAL_API_TOKEN .env | cut -d= -f2)
python eval/run_eval.py
# Загальна точність ≥ 80 % — exit code 0
```

## Ручне тестування чату (curl)

```bash
TOKEN="<JWT-токен після /api/auth/login>"

curl -X POST http://localhost:8080/api/chat \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"message":"Скільки я витратив цього місяця?"}'
```

## Скидання даних для демо

```bash
# Видалити всі транзакції/цілі/логи одного користувача (user_id=1)
docker compose exec db psql -U finagent_user -d finagent -c "
  DELETE FROM transactions WHERE user_id=1;
  DELETE FROM goals WHERE user_id=1;
  DELETE FROM ai_chat_logs WHERE user_id=1;
  DELETE FROM action_log WHERE user_id=1;
"
```
