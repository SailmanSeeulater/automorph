package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/SailmanSeeulater/automorph/internal/media"
	"github.com/SailmanSeeulater/automorph/internal/store"
)

type ImageDTO struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url" doc:"Largest version"`
	Srcset    string    `json:"srcset"`
	Width     int64     `json:"width" doc:"Of the largest version"`
	Height    int64     `json:"height"`
	Alt       string    `json:"alt"`
	Markdown  string    `json:"markdown" doc:"Paste this into a post body"`
	CreatedAt time.Time `json:"createdAt"`
}

func toImageDTO(img store.Image) ImageDTO {
	url := media.URL(img.StorageKey, int(img.Width))
	return ImageDTO{
		ID:        img.ID,
		URL:       url,
		Srcset:    media.Srcset(img.StorageKey, int(img.Width)),
		Width:     img.Width,
		Height:    img.Height,
		Alt:       img.Alt,
		Markdown:  "![" + img.Alt + "](" + url + ")",
		CreatedAt: time.Unix(img.CreatedAt, 0).UTC(),
	}
}

type UploadImageInput struct {
	RawBody huma.MultipartFormFiles[struct {
		File huma.FormFile `form:"file" contentType:"image/jpeg,image/png,image/webp" required:"true"`
		Alt  string        `form:"alt" required:"false" maxLength:"300" doc:"Describe the image for screen readers"`
	}]
}

type ImageOutput struct {
	Body ImageDTO
}

type ImageListOutput struct {
	Body struct {
		Items []ImageDTO `json:"items"`
	}
}

func registerMedia(api huma.API, d Deps) {
	q := d.Queries

	huma.Register(api, huma.Operation{
		OperationID:   "admin-upload-image",
		Method:        http.MethodPost,
		Path:          "/api/admin/media",
		Summary:       "Upload an image",
		Description:   "Resized to WebP at up to 1920px wide. EXIF data (including GPS) is removed.",
		Tags:          []string{"Admin: Media"},
		DefaultStatus: http.StatusCreated,
		MaxBodyBytes:  media.MaxUploadBytes + 1<<20, // + room for the multipart wrapper
	}, func(ctx context.Context, in *UploadImageInput) (*ImageOutput, error) {
		user, err := requireAdmin(ctx)
		if err != nil {
			return nil, err
		}
		form := in.RawBody.Data()
		data, err := io.ReadAll(io.LimitReader(form.File, media.MaxUploadBytes+1))
		if err != nil {
			return nil, err
		}
		if len(data) > media.MaxUploadBytes {
			return nil, huma.Error413RequestEntityTooLarge("image is larger than 20 MB")
		}

		saved, err := d.Media.Save(ctx, data)
		if errors.Is(err, media.ErrUnsupported) || errors.Is(err, media.ErrTooLarge) {
			return nil, huma.Error422UnprocessableEntity(err.Error())
		}
		if err != nil {
			return nil, err
		}

		img, err := q.CreateImage(ctx, store.CreateImageParams{
			UploaderID: sql.NullInt64{Int64: user.ID, Valid: true},
			StorageKey: saved.Key,
			Mime:       "image/webp",
			Width:      int64(saved.Width),
			Height:     int64(saved.Height),
			Bytes:      saved.Bytes,
			Alt:        form.Alt,
		})
		if err != nil {
			d.Media.Delete(saved.Key, saved.Width) // don't leave orphaned files behind
			return nil, err
		}
		return &ImageOutput{Body: toImageDTO(img)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "admin-list-images",
		Method:      http.MethodGet,
		Path:        "/api/admin/media",
		Summary:     "List uploaded images, newest first",
		Tags:        []string{"Admin: Media"},
	}, func(ctx context.Context, _ *struct{}) (*ImageListOutput, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}
		rows, err := q.ListImages(ctx)
		if err != nil {
			return nil, err
		}
		out := &ImageListOutput{}
		out.Body.Items = make([]ImageDTO, 0, len(rows))
		for _, img := range rows {
			out.Body.Items = append(out.Body.Items, toImageDTO(img))
		}
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "admin-delete-image",
		Method:      http.MethodDelete,
		Path:        "/api/admin/media/{id}",
		Summary:     "Delete an image",
		Description: "Posts using it as a cover lose their cover. Links to it inside post bodies will break.",
		Tags:        []string{"Admin: Media"},
	}, func(ctx context.Context, in *struct {
		ID int64 `path:"id"`
	}) (*struct{}, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}
		img, err := q.DeleteImage(ctx, in.ID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, huma.Error404NotFound("image not found")
		}
		if err != nil {
			return nil, err
		}
		// The row is gone either way; a leftover file is only wasted disk space.
		if err := d.Media.Delete(img.StorageKey, int(img.Width)); err != nil {
			d.Logger.Warn("delete image files", "key", img.StorageKey, "err", err)
		}
		return nil, nil
	})
}
