# Локальна робота з FinAgent

## Запуск

```bash
cp .env.example .env
docker compose up -d --build
docker compose ps
```

Зупинка:

```bash
docker compose down
```

Команда `docker compose down -v` також видаляє локальні дані PostgreSQL.

## Логи та перевірка стану

```bash
docker compose logs -f backend
docker compose logs -f ai
docker compose logs -f frontend
curl http://localhost:8080/health
```

AI-сервіс не публікує порт назовні. Його стан можна перевірити з контейнера:

```bash
docker compose exec ai python -c "import urllib.request; print(urllib.request.urlopen('http://localhost:8000/health').read())"
```

## Міграції

Backend запускає міграції під час старту контейнера. Для ручного локального запуску:

```bash
cd backend
go run ./cmd/migrate up
```

## Тести

```bash
cd backend
go test ./...

cd ../financial-ai-agent
python -m pytest tests -v

cd ../frontend
npm run lint
npm run build

cd ..
docker compose config
```

Eval-скрипт потребує запущеного AI-сервісу та ключа Anthropic:

```bash
python eval/run_eval.py
```

## Типові проблеми

- Backend не запускається: перевірити змінні у `.env` і `docker compose logs backend`.
- AI-функції недоступні: перевірити `ANTHROPIC_API_KEY`, `INTERNAL_API_TOKEN` і логи контейнера `ai`.
- Monobank не імпортує дані: перевірити токен, доступні рахунки та часовий ліміт API.

Після зміни `ENCRYPTION_KEY` раніше збережені токени Monobank неможливо розшифрувати.
