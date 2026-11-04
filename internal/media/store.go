package media

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // register decoders for image.DecodeConfig
	_ "image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gen2brain/webp" // also registers the WebP decoder
)

const (
	MaxUploadBytes = 20 << 20   // 20 MB
	maxPixels      = 50_000_000 // ~8K×6K; anything bigger is probably a decompression bomb
	webpQuality    = 80
)

var (
	ErrUnsupported = errors.New("unsupported image: upload a JPEG, PNG or WebP")
	ErrTooLarge    = errors.New("image is too large (max 50 megapixels)")
)

// Store resizes uploads into WebP variants on disk and serves them.
type Store struct {
	dir string
	// One image at a time keeps peak memory around 100 MB, which matters on a 1 GB VPS.
	sem chan struct{}
}

func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	return &Store{dir: dir, sem: make(chan struct{}, 1)}, nil
}

// Saved describes the stored image. Width and Height are of the largest variant.
type Saved struct {
	Key    string
	Width  int
	Height int
	Bytes  int64
}

// Save decodes an upload, fixes phone rotation, and writes a WebP per width.
// Re-encoding drops all EXIF metadata, including GPS location; the original is not kept.
func (s *Store) Save(ctx context.Context, data []byte) (Saved, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png" && format != "webp") {
		return Saved{}, ErrUnsupported
	}
	if cfg.Width*cfg.Height > maxPixels {
		return Saved{}, ErrTooLarge
	}

	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	case <-ctx.Done():
		return Saved{}, ctx.Err()
	}

	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return Saved{}, ErrUnsupported
	}

	saved := Saved{Key: newKey()}
	widths := Widths(min(img.Bounds().Dx(), MaxWidth))
	var written []string
	// Largest first, then each smaller size from the previous one: faster than resizing from the original every time.
	for i := len(widths) - 1; i >= 0; i-- {
		w := widths[i]
		if img.Bounds().Dx() != w {
			img = imaging.Resize(img, w, 0, imaging.Lanczos)
		}
		if saved.Width == 0 {
			saved.Width, saved.Height = img.Bounds().Dx(), img.Bounds().Dy()
		}
		path := s.path(saved.Key, w)
		n, err := writeWebP(path, img)
		if err != nil {
			for _, p := range written {
				os.Remove(p)
			}
			return Saved{}, err
		}
		written = append(written, path)
		saved.Bytes += n
	}
	return saved, nil
}

// Delete removes every variant of an image. Missing files are not an error.
func (s *Store) Delete(key string, maxWidth int) error {
	var errs []error
	for _, w := range Widths(maxWidth) {
		if err := os.Remove(s.path(key, w)); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Handler serves /media/… files. Variant files never change (new uploads get new keys),
// so browsers and Cloudflare may cache them for a year.
func (s *Store) Handler() http.Handler {
	files := http.FileServerFS(os.DirFS(s.dir))
	return http.StripPrefix("/media/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, ".webp") { // also blocks directory listings
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		files.ServeHTTP(w, r)
	}))
}

func (s *Store) path(key string, width int) string {
	return filepath.Join(s.dir, filepath.FromSlash(fmt.Sprintf("%s-%d.webp", key, width)))
}

// newKey returns e.g. "2026/11/k3j4h5…": a folder per month and a random, unguessable name.
func newKey() string {
	return time.Now().UTC().Format("2006/01") + "/" + strings.ToLower(rand.Text())
}

func writeWebP(path string, img image.Image) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return 0, err
	}
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, webp.Options{Quality: webpQuality}); err != nil {
		return 0, err
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o640); err != nil {
		return 0, err
	}
	return int64(buf.Len()), nil
}
