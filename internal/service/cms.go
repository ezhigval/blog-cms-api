package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/ezhigval/blog-cms-api/internal/auth"
	"github.com/ezhigval/blog-cms-api/internal/config"
	"github.com/ezhigval/blog-cms-api/internal/model"
	"github.com/ezhigval/blog-cms-api/internal/repository"
	"github.com/ezhigval/blog-cms-api/internal/slug"
	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrForbidden     = errors.New("forbidden")
	ErrInvalidInput  = errors.New("invalid input")
	ErrFileTooLarge  = errors.New("file too large")
)

type CMS struct {
	cfg        config.Config
	users      *repository.UserRepository
	posts      *repository.PostRepository
	categories *repository.CategoryRepository
	tags       *repository.TagRepository
	media      *repository.MediaRepository
	tokens     *auth.TokenManager
}

func NewCMS(cfg config.Config, users *repository.UserRepository, posts *repository.PostRepository,
	categories *repository.CategoryRepository, tags *repository.TagRepository,
	media *repository.MediaRepository, tokens *auth.TokenManager) *CMS {
	return &CMS{cfg: cfg, users: users, posts: posts, categories: categories, tags: tags, media: media, tokens: tokens}
}

func (s *CMS) Register(ctx context.Context, email, password string) (*model.AuthResponse, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(password) < 8 {
		return nil, ErrInvalidInput
	}
	count, err := s.users.Count(ctx)
	if err != nil {
		return nil, err
	}
	role := model.RoleEditor
	if count == 0 {
		role = model.RoleAdmin
	}
	hash, err := auth.HashPassword(password, s.cfg.BcryptCost)
	if err != nil {
		return nil, err
	}
	user, err := s.users.Create(ctx, email, hash, role)
	if err != nil {
		return nil, err
	}
	token, err := s.tokens.Issue(*user)
	if err != nil {
		return nil, err
	}
	return &model.AuthResponse{AccessToken: token, User: *user}, nil
}

func (s *CMS) Login(ctx context.Context, email, password string) (*model.AuthResponse, error) {
	user, hash, err := s.users.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(email)))
	if err != nil {
		return nil, err
	}
	if user == nil || auth.CheckPassword(hash, password) != nil {
		return nil, auth.ErrInvalidCredentials
	}
	token, err := s.tokens.Issue(*user)
	if err != nil {
		return nil, err
	}
	return &model.AuthResponse{AccessToken: token, User: *user}, nil
}

func (s *CMS) CreatePost(ctx context.Context, authorID int64, req model.CreatePostRequest) (*model.Post, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	if req.Title == "" || req.Body == "" {
		return nil, ErrInvalidInput
	}
	if req.Slug == "" {
		req.Slug = slug.Make(req.Title)
	} else {
		req.Slug = slug.Make(req.Slug)
	}
	if req.Status == "" {
		req.Status = model.StatusDraft
	}
	return s.posts.Create(ctx, authorID, req)
}

func (s *CMS) UpdatePost(ctx context.Context, id int64, req model.UpdatePostRequest) (*model.Post, error) {
	if req.Slug != nil {
		slugVal := slug.Make(*req.Slug)
		req.Slug = &slugVal
	}
	return s.posts.Update(ctx, id, req)
}

func (s *CMS) DeletePost(ctx context.Context, id int64) error {
	return s.posts.Delete(ctx, id)
}

func (s *CMS) GetPost(ctx context.Context, id int64) (*model.Post, error) {
	p, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *CMS) GetPostBySlug(ctx context.Context, slug string, publicOnly bool) (*model.Post, error) {
	p, err := s.posts.GetBySlug(ctx, slug, publicOnly)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *CMS) ListPosts(ctx context.Context, f repository.ListFilter) (*model.PostListResponse, error) {
	return s.posts.List(ctx, f)
}

func (s *CMS) CreateCategory(ctx context.Context, name string) (*model.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidInput
	}
	return s.categories.Create(ctx, name, slug.Make(name))
}

func (s *CMS) ListCategories(ctx context.Context) ([]model.Category, error) {
	list, err := s.categories.List(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return []model.Category{}, nil
	}
	return list, nil
}

func (s *CMS) CreateTag(ctx context.Context, name string) (*model.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidInput
	}
	return s.tags.Create(ctx, name, slug.Make(name))
}

func (s *CMS) ListTags(ctx context.Context) ([]model.Tag, error) {
	list, err := s.tags.List(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return []model.Tag{}, nil
	}
	return list, nil
}

func (s *CMS) Upload(ctx context.Context, uploaderID int64, header *multipart.FileHeader) (*model.Media, error) {
	if header.Size > s.cfg.MaxUploadMB*1024*1024 {
		return nil, ErrFileTooLarge
	}
	mime := header.Header.Get("Content-Type")
	if !strings.HasPrefix(mime, "image/") {
		return nil, ErrInvalidInput
	}

	if err := os.MkdirAll(s.cfg.UploadDir, 0o755); err != nil {
		return nil, err
	}

	ext := filepath.Ext(header.Filename)
	name := uuid.New().String() + ext
	path := filepath.Join(s.cfg.UploadDir, name)

	src, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/uploads/%s", strings.TrimRight(s.cfg.PublicURL, "/"), name)
	return s.media.Create(ctx, uploaderID, header.Filename, url, mime, header.Size)
}
