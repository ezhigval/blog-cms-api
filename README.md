# blog-cms-api

Blog CMS: **Go REST API** + **Next.js 15** frontend. Posts, categories, tags, pg_trgm search, JWT auth, image uploads.

## Quick start

```bash
make docker-up && sleep 5 && make seed
open http://localhost:3001        # public blog (ISR)
open http://localhost:3001/admin  # admin panel
```

API: `http://localhost:8090`

## Stack

| Layer | Tech |
|-------|------|
| API | Go, chi, pgx, JWT, pg_trgm |
| Web | Next.js 15, React 19, ISR, react-markdown |
| DB | PostgreSQL 16 |

## API highlights

| Method | Path | Auth |
|--------|------|------|
| POST | `/api/v1/auth/register` | no (first user = admin) |
| POST | `/api/v1/auth/login` | no |
| GET | `/api/v1/posts?q=` | public, paginated, trgm search |
| GET | `/api/v1/posts/slug/{slug}` | public |
| POST | `/api/v1/admin/posts` | JWT |
| POST | `/api/v1/admin/media` | JWT, multipart upload |

Search uses `pg_trgm` (`title % query OR body % query`).

## Frontend

- `/` — published posts, `revalidate: 60`
- `/posts/[slug]` — markdown body, ISR
- `/admin` — login + post list (token in localStorage)

```bash
make web-dev   # Next.js on :3001, API on :8090
```

## Full-stack contract

OpenAPI-style JSON over REST. CORS enabled for `localhost:3001`. Uploads served at `/uploads/*`.

Port **8090** (API) · **3001** (web) · MIT
