package model

import "time"

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
)

type PostStatus string

const (
	StatusDraft     PostStatus = "draft"
	StatusPublished PostStatus = "published"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Post struct {
	ID          int64      `json:"id"`
	AuthorID    int64      `json:"author_id"`
	CategoryID  *int64     `json:"category_id,omitempty"`
	Category    *Category  `json:"category,omitempty"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     string     `json:"excerpt,omitempty"`
	Body        string     `json:"body"`
	Status      PostStatus `json:"status"`
	CoverURL    string     `json:"cover_url,omitempty"`
	Tags        []Tag      `json:"tags,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Media struct {
	ID        int64     `json:"id"`
	Filename  string    `json:"filename"`
	URL       string    `json:"url"`
	MimeType  string    `json:"mime_type"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
}

type PostListResponse struct {
	Items      []Post `json:"items"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	TotalPages int    `json:"total_pages"`
}

type CreatePostRequest struct {
	Title      string     `json:"title"`
	Slug       string     `json:"slug"`
	Excerpt    string     `json:"excerpt"`
	Body       string     `json:"body"`
	Status     PostStatus `json:"status"`
	CategoryID *int64     `json:"category_id"`
	CoverURL   string     `json:"cover_url"`
	TagIDs     []int64    `json:"tag_ids"`
}

type UpdatePostRequest struct {
	Title      *string     `json:"title"`
	Slug       *string     `json:"slug"`
	Excerpt    *string     `json:"excerpt"`
	Body       *string     `json:"body"`
	Status     *PostStatus `json:"status"`
	CategoryID *int64      `json:"category_id"`
	CoverURL   *string     `json:"cover_url"`
	TagIDs     []int64     `json:"tag_ids"`
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}
