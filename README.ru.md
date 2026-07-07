# blog-cms-api

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
[![CI](https://github.com/ezhigval/blog-cms-api/actions/workflows/ci.yml/badge.svg)](https://github.com/ezhigval/blog-cms-api/actions/workflows/ci.yml)
![License](https://img.shields.io/badge/license-MIT-blue)
![Tier](https://img.shields.io/badge/tier-middle-5319e7)

**[English](README.md)** · Русский

Блог CMS: **Go REST API** + фронтенд **Next.js 15**. Посты, категории, теги, поиск pg_trgm, JWT, загрузка картинок.

## Быстрый старт

```bash
make docker-up && sleep 5 && make seed
open http://localhost:3001        # публичный блог (ISR)
open http://localhost:3001/admin  # админка
```

API: `http://localhost:8090`

## Стек

| Слой | Технологии |
|------|------------|
| API | Go, chi, pgx, JWT, pg_trgm |
| Web | Next.js 15, React 19, ISR, react-markdown |
| DB | PostgreSQL 16 |

## API

| Метод | Путь | Auth |
|--------|------|------|
| POST | `/api/v1/auth/register` | нет (первый пользователь = admin) |
| POST | `/api/v1/auth/login` | нет |
| GET | `/api/v1/posts?q=` | публично, пагинация, trgm-поиск |
| GET | `/api/v1/posts/slug/{slug}` | публично |
| POST | `/api/v1/admin/posts` | JWT |
| POST | `/api/v1/admin/media` | JWT, multipart upload |

Поиск через `pg_trgm` (`title % query OR body % query`).

## Фронтенд

- `/` — опубликованные посты, `revalidate: 60`
- `/posts/[slug]` — markdown-тело, ISR
- `/admin` — логин + список постов (токен в localStorage)

```bash
make web-dev   # Next.js на :3001, API на :8090
```

## Контракт full-stack

JSON в стиле OpenAPI поверх REST. CORS для `localhost:3001`. Загрузки по `/uploads/*`.

Порт **8090** (API) · **3001** (web) · MIT
