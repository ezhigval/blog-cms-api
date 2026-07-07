package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ezhigval/blog-cms-api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) Create(ctx context.Context, name, slug string) (*model.Category, error) {
	var c model.Category
	err := r.pool.QueryRow(ctx, `
		INSERT INTO categories (name, slug) VALUES ($1, $2) RETURNING id, name, slug
	`, name, slug).Scan(&c.ID, &c.Name, &c.Slug)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) List(ctx context.Context) ([]model.Category, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, slug FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

type TagRepository struct {
	pool *pgxpool.Pool
}

func NewTagRepository(pool *pgxpool.Pool) *TagRepository {
	return &TagRepository{pool: pool}
}

func (r *TagRepository) Create(ctx context.Context, name, slug string) (*model.Tag, error) {
	var t model.Tag
	err := r.pool.QueryRow(ctx, `
		INSERT INTO tags (name, slug) VALUES ($1, $2) RETURNING id, name, slug
	`, name, slug).Scan(&t.ID, &t.Name, &t.Slug)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TagRepository) List(ctx context.Context) ([]model.Tag, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, slug FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

type MediaRepository struct {
	pool *pgxpool.Pool
}

func NewMediaRepository(pool *pgxpool.Pool) *MediaRepository {
	return &MediaRepository{pool: pool}
}

func (r *MediaRepository) Create(ctx context.Context, uploaderID int64, filename, url, mime string, size int64) (*model.Media, error) {
	var m model.Media
	err := r.pool.QueryRow(ctx, `
		INSERT INTO media (uploader_id, filename, url, mime_type, size_bytes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, filename, url, mime_type, size_bytes, created_at
	`, uploaderID, filename, url, mime, size).Scan(&m.ID, &m.Filename, &m.URL, &m.MimeType, &m.SizeBytes, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create media: %w", err)
	}
	return &m, nil
}

func scanPostFromRows(rows pgx.Rows) (model.Post, error) {
	var p model.Post
	var categoryID *int64
	var cID *int64
	var cName, cSlug *string
	err := rows.Scan(
		&p.ID, &p.AuthorID, &categoryID, &p.Title, &p.Slug, &p.Excerpt, &p.Body, &p.Status,
		&p.CoverURL, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt,
		&cID, &cName, &cSlug,
	)
	if err != nil {
		return p, err
	}
	p.CategoryID = categoryID
	if cID != nil && cName != nil && cSlug != nil {
		p.Category = &model.Category{ID: *cID, Name: *cName, Slug: *cSlug}
	}
	return p, nil
}

func scanPostRow(row pgx.Row) (*model.Post, error) {
	var p model.Post
	var categoryID *int64
	var cID *int64
	var cName, cSlug *string
	err := row.Scan(
		&p.ID, &p.AuthorID, &categoryID, &p.Title, &p.Slug, &p.Excerpt, &p.Body, &p.Status,
		&p.CoverURL, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt,
		&cID, &cName, &cSlug,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.CategoryID = categoryID
	if cID != nil && cName != nil && cSlug != nil {
		p.Category = &model.Category{ID: *cID, Name: *cName, Slug: *cSlug}
	}
	return &p, nil
}
