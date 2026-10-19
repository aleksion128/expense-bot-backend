# Деплой Budgetly рядом с formula-ai.online

Можно держать оба сайта на одном сервере: nginx слушает 80/443 и проксирует на разные docker-контейнеры.

## 1. DNS

В панели домена добавь:

| Тип | Имя | Значение |
|-----|-----|----------|
| A | `budgetly` | IP твоего сервера |

Подожди пару минут, проверь: `ping budgetly.formula-ai.online`.

## 2. Код на сервере

```bash
cd /opt   # или куда удобно
git clone https://github.com/aleksion128/expense-bot-backend.git budgetly
cd budgetly

# пароль БД (обязательно смени)
export POSTGRES_PASSWORD='сильный-пароль'

docker compose -f docker-compose.prod.yml up -d --build
```

API+UI будут на `127.0.0.1:8443` (снаружи не торчат).

Проверка:

```bash
curl -s http://127.0.0.1:8443/health
```

## 3. Nginx

```bash
sudo cp deploy/nginx-budgetly.conf.example /etc/nginx/sites-available/budgetly.formula-ai.online
sudo ln -sf /etc/nginx/sites-available/budgetly.formula-ai.online /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

HTTPS:

```bash
sudo certbot --nginx -d budgetly.formula-ai.online
```

## 4. Обновление

```bash
cd /opt/budgetly
git pull
docker compose -f docker-compose.prod.yml up -d --build
```

## Важно

- Порт `8443` не должен быть занят другим контейнером. Если занят — поменяй в `docker-compose.prod.yml` и в nginx.
- Postgres Budgetly **отдельный** (свой volume) — не смешивается с БД formula-ai.
- `formula-ai.online` продолжает работать как раньше: у него свой `server {}` и свой контейнер.
