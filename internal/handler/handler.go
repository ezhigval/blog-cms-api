package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/ezhigval/blog-cms-api/internal/middleware"
	"github.com/ezhigval/blog-cms-api/internal/model"
	"github.com/ezhigval/blog-cms-api/internal/repository"
	"github.com/ezhigval/blog-cms-api/internal/service"
	"github.com/ezhigval/go-toolkit/httputil"
	"github.com/ezhigval/blog-cms-api/internal/auth"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *service.CMS
}

func New(svc *service.CMS) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.svc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		writeSvcErr(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			httputil.WriteError(w, httputil.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials", err))
			return
		}
		writeSvcErr(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, httputil.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "no claims", nil))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"id": claims.UserID, "email": claims.Email, "role": claims.Role,
	})
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserID(r.Context())
	if !ok {
		return
	}
	var req model.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	p, err := h.svc.CreatePost(r.Context(), uid, req)
	writeResult(w, http.StatusCreated, p, err)
}

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var req model.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	p, err := h.svc.UpdatePost(r.Context(), id, req)
	writeResult(w, http.StatusOK, p, err)
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.svc.DeletePost(r.Context(), id); err != nil {
		writeSvcErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetPostAdmin(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	p, err := h.svc.GetPost(r.Context(), id)
	writeResult(w, http.StatusOK, p, err)
}

func (h *Handler) ListPostsAdmin(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.ListPosts(r.Context(), parseListFilter(r, ""))
	writeResult(w, http.StatusOK, resp, err)
}

func (h *Handler) ListPostsPublic(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.ListPosts(r.Context(), parseListFilter(r, model.StatusPublished))
	writeResult(w, http.StatusOK, resp, err)
}

func (h *Handler) GetPostPublic(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	p, err := h.svc.GetPostBySlug(r.Context(), slug, true)
	writeResult(w, http.StatusOK, p, err)
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	c, err := h.svc.CreateCategory(r.Context(), req.Name)
	writeResult(w, http.StatusCreated, c, err)
}

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListCategories(r.Context())
	writeResult(w, http.StatusOK, list, err)
}

func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	t, err := h.svc.CreateTag(r.Context(), req.Name)
	writeResult(w, http.StatusCreated, t, err)
}

func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListTags(r.Context())
	writeResult(w, http.StatusOK, list, err)
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserID(r.Context())
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	file.Close()
	m, err := h.svc.Upload(r.Context(), uid, header)
	writeResult(w, http.StatusCreated, m, err)
}

func parseListFilter(r *http.Request, status model.PostStatus) repository.ListFilter {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	return repository.ListFilter{
		Status:  status,
		Query:   r.URL.Query().Get("q"),
		Page:    page,
		PerPage: perPage,
	}
}

func writeResult(w http.ResponseWriter, status int, v any, err error) {
	if err != nil {
		writeSvcErr(w, err)
		return
	}
	httputil.WriteJSON(w, status, v)
}

func writeSvcErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		httputil.WriteError(w, httputil.NewAppError(http.StatusNotFound, "NOT_FOUND", err.Error(), err))
	case errors.Is(err, service.ErrInvalidInput), errors.Is(err, service.ErrFileTooLarge):
		httputil.WriteError(w, httputil.NewAppError(http.StatusBadRequest, "BAD_REQUEST", err.Error(), err))
	default:
		httputil.WriteError(w, httputil.NewAppError(http.StatusInternalServerError, "INTERNAL", err.Error(), err))
	}
}

func writeErr(w http.ResponseWriter, status int, err error) {
	httputil.WriteError(w, httputil.NewAppError(status, "BAD_REQUEST", err.Error(), err))
}
