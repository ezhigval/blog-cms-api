package repository

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ezhigval/blog-cms-api/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostRepository struct {
	pool *pgxpool.Pool
}

func NewPostRepository(pool *pgxpool.Pool) *PostRepository {
	return &PostRepository{pool: pool}
}

type ListFilter struct {
	Status  model.PostStatus
	Query   string
	Page    int
	PerPage int
}

func (r *PostRepository) Create(ctx context.Context, authorID int64, req model.CreatePostRequest) (*model.Post, error) {
	var p model.Post
	var pubAt *time.Time
	if req.Status == model.StatusPublished {
		now := time.Now().UTC()
		pubAt = &now
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO posts (author_id, category_id, title, slug, excerpt, body, status, cover_url, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, author_id, category_id, title, slug, COALESCE(excerpt,''), body, status,
			COALESCE(cover_url,''), published_at, created_at, updated_at
	`, authorID, req.CategoryID, req.Title, req.Slug, req.Excerpt, req.Body, req.Status, req.CoverURL, pubAt).Scan(
		&p.ID, &p.AuthorID, &p.CategoryID, &p.Title, &p.Slug, &p.Excerpt, &p.Body, &p.Status,
		&p.CoverURL, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}
	if err := r.setTags(ctx, p.ID, req.TagIDs); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, p.ID)
}

func (r *PostRepository) Update(ctx context.Context, id int64, req model.UpdatePostRequest) (*model.Post, error) {
	current, err := r.GetByID(ctx, id)
	if err != nil || current == nil {
		return current, err
	}

	title := current.Title
	slug := current.Slug
	excerpt := current.Excerpt
	body := current.Body
	status := current.Status
	categoryID := current.CategoryID
	coverURL := current.CoverURL

	if req.Title != nil {
		title = *req.Title
	}
	if req.Slug != nil {
		slug = *req.Slug
	}
	if req.Excerpt != nil {
		excerpt = *req.Excerpt
	}
	if req.Body != nil {
		body = *req.Body
	}
	if req.Status != nil {
		status = *req.Status
	}
	if req.CategoryID != nil {
		categoryID = req.CategoryID
	}
	if req.CoverURL != nil {
		coverURL = *req.CoverURL
	}

	var pubAt *time.Time = current.PublishedAt
	if status == model.StatusPublished && current.PublishedAt == nil {
		now := time.Now().UTC()
		pubAt = &now
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE posts SET title=$1, slug=$2, excerpt=$3, body=$4, status=$5,
			category_id=$6, cover_url=$7, published_at=$8, updated_at=NOW()
		WHERE id=$9
	`, title, slug, excerpt, body, status, categoryID, coverURL, pubAt, id)
	if err != nil {
		return nil, err
	}
	if req.TagIDs != nil {
		if err := r.setTags(ctx, id, req.TagIDs); err != nil {
			return nil, err
		}
	}
	return r.GetByID(ctx, id)
}

func (r *PostRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM posts WHERE id = $1`, id)
	return err
}

func (r *PostRepository) GetByID(ctx context.Context, id int64) (*model.Post, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT p.id, p.author_id, p.category_id, p.title, p.slug, COALESCE(p.excerpt,''), p.body,
			p.status, COALESCE(p.cover_url,''), p.published_at, p.created_at, p.updated_at,
			c.id, c.name, c.slug
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1
	`, id)
	p, err := scanPostRow(row)
	if err != nil || p == nil {
		return p, err
	}
	tags, _ := r.tagsForPost(ctx, p.ID)
	p.Tags = tags
	return p, nil
}

func (r *PostRepository) GetBySlug(ctx context.Context, slug string, publicOnly bool) (*model.Post, error) {
	q := `
		SELECT p.id, p.author_id, p.category_id, p.title, p.slug, COALESCE(p.excerpt,''), p.body,
			p.status, COALESCE(p.cover_url,''), p.published_at, p.created_at, p.updated_at,
			c.id, c.name, c.slug
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.slug = $1`
	if publicOnly {
		q += ` AND p.status = 'published'`
	}
	row := r.pool.QueryRow(ctx, q, slug)
	p, err := scanPostRow(row)
	if err != nil || p == nil {
		return p, err
	}
	tags, _ := r.tagsForPost(ctx, p.ID)
	p.Tags = tags
	return p, nil
}

func (r *PostRepository) List(ctx context.Context, f ListFilter) (*model.PostListResponse, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage <= 0 || f.PerPage > 50 {
		f.PerPage = 10
	}
	offset := (f.Page - 1) * f.PerPage

	where := []string{"1=1"}
	args := []any{}
	i := 1

	if f.Status != "" {
		where = append(where, fmt.Sprintf("p.status = $%d", i))
		args = append(args, f.Status)
		i++
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		where = append(where, fmt.Sprintf("(p.title %% $%d OR p.body %% $%d OR p.title ILIKE $%d)", i, i, i+1))
		args = append(args, q, "%"+q+"%")
		i += 2
	}

	whereSQL := strings.Join(where, " AND ")
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM posts p WHERE %s`, whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}

	listSQL := fmt.Sprintf(`
		SELECT p.id, p.author_id, p.category_id, p.title, p.slug, COALESCE(p.excerpt,''), p.body,
			p.status, COALESCE(p.cover_url,''), p.published_at, p.created_at, p.updated_at,
			c.id, c.name, c.slug
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE %s
		ORDER BY COALESCE(p.published_at, p.created_at) DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, i, i+1)
	args = append(args, f.PerPage, offset)

	rows, err := r.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Post
	for rows.Next() {
		p, err := scanPostFromRows(rows)
		if err != nil {
			return nil, err
		}
		tags, _ := r.tagsForPost(ctx, p.ID)
		p.Tags = tags
		items = append(items, p)
	}
	if items == nil {
		items = []model.Post{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(f.PerPage)))
	return &model.PostListResponse{
		Items: items, Total: total, Page: f.Page, PerPage: f.PerPage, TotalPages: totalPages,
	}, rows.Err()
}

func (r *PostRepository) setTags(ctx context.Context, postID int64, tagIDs []int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM post_tags WHERE post_id = $1`, postID)
	if err != nil {
		return err
	}
	for _, tid := range tagIDs {
		_, err := r.pool.Exec(ctx, `INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2)`, postID, tid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostRepository) tagsForPost(ctx context.Context, postID int64) ([]model.Tag, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.id, t.name, t.slug FROM tags t
		JOIN post_tags pt ON pt.tag_id = t.id WHERE pt.post_id = $1
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}
