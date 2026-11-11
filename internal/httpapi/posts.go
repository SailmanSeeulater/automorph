package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/SailmanSeeulater/automorph/internal/db"
	"github.com/SailmanSeeulater/automorph/internal/posts"
	"github.com/SailmanSeeulater/automorph/internal/store"
)

// ---- Response shapes ----

type PostSummary struct {
	ID           int64      `json:"id"`
	Slug         string     `json:"slug"`
	Title        string     `json:"title"`
	Excerpt      string     `json:"excerpt"`
	Status       string     `json:"status" enum:"draft,published"`
	CoverMediaID *int64     `json:"coverMediaId,omitempty"`
	Cover        *ImageDTO  `json:"cover,omitempty"`
	PublishedAt  *time.Time `json:"publishedAt,omitempty"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// PublicPost is what readers get: Markdown already rendered to HTML.
type PublicPost struct {
	PostSummary
	BodyHTML string `json:"bodyHtml"`
}

// AdminPost is what the editor gets: the raw Markdown.
type AdminPost struct {
	PostSummary
	BodyMd    string    `json:"bodyMd"`
	CreatedAt time.Time `json:"createdAt"`
}

func toSummary(p store.Post) PostSummary {
	s := PostSummary{
		ID:        p.ID,
		Slug:      p.Slug,
		Title:     p.Title,
		Excerpt:   p.Excerpt,
		Status:    p.Status,
		UpdatedAt: time.Unix(p.UpdatedAt, 0).UTC(),
	}
	if p.CoverMediaID.Valid {
		s.CoverMediaID = &p.CoverMediaID.Int64
	}
	if p.PublishedAt.Valid {
		t := time.Unix(p.PublishedAt.Int64, 0).UTC()
		s.PublishedAt = &t
	}
	return s
}

func toAdminPost(ctx context.Context, q *store.Queries, p store.Post) AdminPost {
	a := AdminPost{PostSummary: toSummary(p), BodyMd: p.BodyMd, CreatedAt: time.Unix(p.CreatedAt, 0).UTC()}
	withCover(ctx, q, &a.PostSummary)
	return a
}

// withCover fills in s.Cover. A missing image just means no cover, so errors are ignored.
// It's one extra query per post, which costs microseconds with SQLite in the same process.
func withCover(ctx context.Context, q *store.Queries, s *PostSummary) {
	if s.CoverMediaID == nil {
		return
	}
	if img, err := q.GetImage(ctx, *s.CoverMediaID); err == nil {
		dto := toImageDTO(img)
		s.Cover = &dto
	}
}

// postError turns database errors into the right HTTP status.
func postError(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return huma.Error404NotFound("post not found")
	case db.IsUniqueViolation(err):
		return huma.Error409Conflict("that slug is already used by another post")
	case db.IsForeignKeyViolation(err):
		return huma.Error422UnprocessableEntity("coverMediaId doesn't match an uploaded image")
	}
	return err
}

func nullInt(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

// ---- Request shapes ----

// PostFields are the editable fields, shared by create and update.
type PostFields struct {
	Title        string `json:"title" minLength:"1" maxLength:"200"`
	Excerpt      string `json:"excerpt,omitempty" maxLength:"300"`
	BodyMd       string `json:"bodyMd,omitempty" maxLength:"200000"`
	CoverMediaID *int64 `json:"coverMediaId,omitempty"`
}

type CreatePostInput struct {
	Body struct {
		PostFields
		Slug   string `json:"slug,omitempty" maxLength:"80" pattern:"^[a-z0-9]+(-[a-z0-9]+)*$" doc:"Generated from the title if left out"`
		Status string `json:"status,omitempty" enum:"draft,published" default:"draft"`
	}
}

type UpdatePostInput struct {
	ID   int64 `path:"id"`
	Body struct {
		PostFields
		Slug   string `json:"slug" maxLength:"80" pattern:"^[a-z0-9]+(-[a-z0-9]+)*$" doc:"Changing it breaks existing links to the post"`
		Status string `json:"status" enum:"draft,published"`
	}
}

type PostIDInput struct {
	ID int64 `path:"id"`
}

type ListPostsInput struct {
	Limit  int64 `query:"limit" default:"20" minimum:"1" maximum:"50"`
	Offset int64 `query:"offset" default:"0" minimum:"0"`
}

type ListPostsOutput struct {
	Body struct {
		Items   []PostSummary `json:"items"`
		HasMore bool          `json:"hasMore"`
	}
}

type PublicPostOutput struct {
	Body PublicPost
}

type AdminPostOutput struct {
	Body AdminPost
}

type AdminPostListOutput struct {
	Body struct {
		Items []PostSummary `json:"items"`
	}
}

type PreviewInput struct {
	Body struct {
		BodyMd string `json:"bodyMd" maxLength:"200000"`
	}
}

type PreviewOutput struct {
	Body struct {
		HTML string `json:"html"`
	}
}

func registerPosts(api huma.API, d Deps) {
	q := d.Queries

	// ---- Public ----

	huma.Register(api, huma.Operation{
		OperationID: "list-posts",
		Method:      http.MethodGet,
		Path:        "/api/posts",
		Summary:     "List published posts, newest first",
		Tags:        []string{"Posts"},
	}, func(ctx context.Context, in *ListPostsInput) (*ListPostsOutput, error) {
		// Ask for one extra row to find out whether there's another page.
		rows, err := q.ListPublishedPosts(ctx, store.ListPublishedPostsParams{Limit: in.Limit + 1, Offset: in.Offset})
		if err != nil {
			return nil, err
		}
		out := &ListPostsOutput{}
		out.Body.Items = []PostSummary{}
		for i, p := range rows {
			if int64(i) == in.Limit {
				out.Body.HasMore = true
				break
			}
			s := toSummary(p)
			withCover(ctx, q, &s)
			out.Body.Items = append(out.Body.Items, s)
		}
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-post",
		Method:      http.MethodGet,
		Path:        "/api/posts/{slug}",
		Summary:     "Get a published post",
		Tags:        []string{"Posts"},
	}, func(ctx context.Context, in *struct {
		Slug string `path:"slug"`
	}) (*PublicPostOutput, error) {
		p, err := q.GetPublishedPostBySlug(ctx, in.Slug)
		if err != nil {
			return nil, postError(err)
		}
		html, err := posts.RenderMarkdown(p.BodyMd)
		if err != nil {
			return nil, err
		}
		post := PublicPost{PostSummary: toSummary(p), BodyHTML: html}
		withCover(ctx, q, &post.PostSummary)
		return &PublicPostOutput{Body: post}, nil
	})

	// ---- Admin (drafts included) ----

	huma.Register(api, huma.Operation{
		OperationID: "admin-list-posts",
		Method:      http.MethodGet,
		Path:        "/api/admin/posts",
		Summary:     "List all posts, including drafts",
		Tags:        []string{"Admin: Posts"},
	}, func(ctx context.Context, _ *struct{}) (*AdminPostListOutput, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}
		rows, err := q.ListAllPosts(ctx)
		if err != nil {
			return nil, err
		}
		out := &AdminPostListOutput{}
		out.Body.Items = make([]PostSummary, 0, len(rows))
		for _, p := range rows {
			s := toSummary(p)
			withCover(ctx, q, &s)
			out.Body.Items = append(out.Body.Items, s)
		}
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "admin-get-post",
		Method:      http.MethodGet,
		Path:        "/api/admin/posts/{id}",
		Summary:     "Get any post for editing",
		Tags:        []string{"Admin: Posts"},
	}, func(ctx context.Context, in *PostIDInput) (*AdminPostOutput, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}
		p, err := q.GetPost(ctx, in.ID)
		if err != nil {
			return nil, postError(err)
		}
		return &AdminPostOutput{Body: toAdminPost(ctx, q, p)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "admin-create-post",
		Method:        http.MethodPost,
		Path:          "/api/admin/posts",
		Summary:       "Create a post",
		Tags:          []string{"Admin: Posts"},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *CreatePostInput) (*AdminPostOutput, error) {
		user, err := requireAdmin(ctx)
		if err != nil {
			return nil, err
		}
		slug := in.Body.Slug
		if slug == "" {
			slug, err = posts.UniqueSlug(ctx, q.SlugTaken, posts.Slugify(in.Body.Title))
			if err != nil {
				return nil, err
			}
		}
		p, err := q.CreatePost(ctx, store.CreatePostParams{
			AuthorID:     user.ID,
			Slug:         slug,
			Title:        in.Body.Title,
			Excerpt:      in.Body.Excerpt,
			BodyMd:       in.Body.BodyMd,
			CoverMediaID: nullInt(in.Body.CoverMediaID),
			Status:       in.Body.Status,
		})
		if err != nil {
			return nil, postError(err)
		}
		return &AdminPostOutput{Body: toAdminPost(ctx, q, p)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "admin-update-post",
		Method:      http.MethodPut,
		Path:        "/api/admin/posts/{id}",
		Summary:     "Replace a post's editable fields",
		Description: "Send every field: anything left out is cleared. Setting status to published for the first time sets publishedAt.",
		Tags:        []string{"Admin: Posts"},
	}, func(ctx context.Context, in *UpdatePostInput) (*AdminPostOutput, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}
		p, err := q.UpdatePost(ctx, store.UpdatePostParams{
			ID:           in.ID,
			Slug:         in.Body.Slug,
			Title:        in.Body.Title,
			Excerpt:      in.Body.Excerpt,
			BodyMd:       in.Body.BodyMd,
			CoverMediaID: nullInt(in.Body.CoverMediaID),
			Status:       in.Body.Status,
		})
		if err != nil {
			return nil, postError(err)
		}
		return &AdminPostOutput{Body: toAdminPost(ctx, q, p)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "admin-delete-post",
		Method:      http.MethodDelete,
		Path:        "/api/admin/posts/{id}",
		Summary:     "Delete a post",
		Tags:        []string{"Admin: Posts"},
	}, func(ctx context.Context, in *PostIDInput) (*struct{}, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}
		n, err := q.DeletePost(ctx, in.ID)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, huma.Error404NotFound("post not found")
		}
		return nil, nil
	})

	// The editor's live preview uses the same renderer as the public site,
	// so what you see while writing is exactly what readers get.
	huma.Register(api, huma.Operation{
		OperationID: "admin-preview-markdown",
		Method:      http.MethodPost,
		Path:        "/api/admin/preview",
		Summary:     "Render Markdown to HTML",
		Tags:        []string{"Admin: Posts"},
	}, func(ctx context.Context, in *PreviewInput) (*PreviewOutput, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}
		html, err := posts.RenderMarkdown(in.Body.BodyMd)
		if err != nil {
			return nil, err
		}
		out := &PreviewOutput{}
		out.Body.HTML = html
		return out, nil
	})
}
