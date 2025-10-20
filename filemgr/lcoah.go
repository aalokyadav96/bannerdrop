package filemgr

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

const (
	defaultThumbWidth = 500
	maxUploadSize     = 10 << 20 // 10 MB
	defaultQuality    = 85
)

// -------------------------
// Public Save Functions
// -------------------------

func SaveFileForEntity(file multipart.File, header *multipart.FileHeader, entity EntityType, picType PictureType) (string, string, error) {
	defer file.Close()
	log.Println("->[SaveFileForEntity] : no error yet")
	filename, ext, err := saveFileAndProcess(file, header, entity, picType, defaultThumbWidth, "")
	log.Println("[SaveFileForEntity]->")
	return filename, ext, err
}

func SaveImageWithThumb(file multipart.File, header *multipart.FileHeader, entity EntityType, picType PictureType, thumbWidth int, userid string) (string, string, error) {
	defer file.Close()
	filename, ext, err := saveFileAndProcess(file, header, entity, picType, thumbWidth, userid)
	if err != nil {
		return filename + ext, "", err
	}

	// If thumbnail not already created, return empty string
	thumbName := ""
	fullPath := filepath.Join(ResolvePath(entity, picType), filename)
	if img, _, err := openImage(fullPath); err == nil {
		if img.Bounds().Dx() > thumbWidth || img.Bounds().Dy() > thumbWidth {
			thumbName = userid + ".jpg"
			if err := generateThumbnail(img, entity, thumbName, thumbWidth); err != nil {
				return filename, "", fmt.Errorf("thumbnail failed: %w", err)
			}
		}
	}
	return filename, thumbName, nil
}

// -------------------------
// Internal helper
// -------------------------

func saveMultipartFile(hdr *multipart.FileHeader, entity EntityType, picType PictureType) (string, string, error) {
	file, err := hdr.Open()
	if err != nil {
		return "", "", fmt.Errorf("open %s: %w", hdr.Filename, err)
	}
	defer file.Close()

	return saveFileAndProcess(file, hdr, entity, picType, defaultThumbWidth, "")
}

// -------------------------
// Multipart Form Helpers
// -------------------------

// SaveFormFile saves a single file from a multipart.Form.
// Returns the saved filename or error.
func SaveFormFile(form *multipart.Form, formKey string, entity EntityType, picType PictureType, required bool) (string, error) {
	files := form.File[formKey]
	if len(files) == 0 {
		if required {
			return "", fmt.Errorf("missing required file: %s", formKey)
		}
		return "", nil
	}

	filename, ext, err := saveMultipartFile(files[0], entity, picType)
	return filename + ext, err
}

// SaveFormFiles saves multiple files under the same form key.
// Returns list of saved filenames or partial errors.
func SaveFormFiles(form *multipart.Form, formKey string, entity EntityType, picType PictureType, required bool) ([]string, error) {
	files := form.File[formKey]
	if len(files) == 0 {
		if required {
			return nil, fmt.Errorf("missing required files: %s", formKey)
		}
		return nil, nil
	}

	var saved []string
	var errs []string
	for _, hdr := range files {
		filename, ext, err := saveMultipartFile(hdr, entity, picType)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", hdr.Filename, err))
			continue
		}
		saved = append(saved, filename+ext)
	}

	if len(errs) > 0 {
		return saved, fmt.Errorf("errors saving files: %s", strings.Join(errs, "; "))
	}
	return saved, nil
}

// SaveFormFilesByKeys saves files for multiple keys in the form.
// Returns all successfully saved filenames and any partial errors.
func SaveFormFilesByKeys(form *multipart.Form, keys []string, entity EntityType, picType PictureType, required bool) ([]string, error) {
	var allSaved []string
	var errs []string

	for _, key := range keys {
		saved, err := SaveFormFiles(form, key, entity, picType, required)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", key, err))
		}
		allSaved = append(allSaved, saved...)
	}

	if len(errs) > 0 {
		return allSaved, fmt.Errorf("errors: %s", strings.Join(errs, "; "))
	}
	return allSaved, nil
}

// -------------------------
// Core DRY Helper
// -------------------------

func saveFileAndProcess(file multipart.File, header *multipart.FileHeader, entity EntityType, picType PictureType, thumbWidth int, userid string) (string, string, error) {
	path := ResolvePath(entity, picType)

	log.Println("->[saveFileAndProcess] : no error yet")
	filename, ext, fullPath, err := writeValidatedFile(file, header, path, picType, maxUploadSize)
	if err != nil {
		log.Println("[saveFileAndProcess]->")
		return "", "", err
	}

	log.Println("->[saveFileAndProcess 1] : no error yet")

	if isImageType(picType) {
		thumbName := userid
		if thumbName == "" {
			thumbName = filename
		}
		if err := processImage(fullPath, entity, picType, thumbWidth, thumbName, ext); err != nil {
			return filename, ext, err
		}
		// if err := processImage(fullPath, entity, picType, thumbWidth, filename, ext); err != nil {
		// 	return filename, ext, err
		// }
	} else if picType == PicVideo || isVideoExt(ext) {
		go func(vpath string, ent EntityType, fname string) {
			if thumb, err := generateVideoPoster(vpath, ent, fname); err != nil {
				if LogFunc != nil {
					LogFunc(fmt.Sprintf("warning: video poster generation failed for %s: %v", fname, err), 0, "")
				}
			} else if LogFunc != nil {
				LogFunc(thumb, 0, "image/jpeg")
			}
		}(fullPath, entity, filename+ext)
	}

	return filename, ext, nil
}

// -------------------------
// Image/Video Processing
// -------------------------

func processImage(fullPath string, entity EntityType, picType PictureType, thumbWidth int, filename, ext string) error {
	_ = picType
	img, _, err := openImage(fullPath)
	if err != nil {
		if LogFunc != nil {
			LogFunc(fullPath, 0, "unknown")
		}
		return nil // best-effort
	}

	newPath, err := normalizeImageFormat(fullPath, ext, img)
	if err != nil {
		return err
	}
	if newPath != fullPath {
		fullPath = newPath
	}

	// Thumbnail
	imgCopy := imaging.Clone(img)
	go func() {
		// thumbName := filepath.Base(fullPath)
		thumbName := filename + ".jpg"
		if err := generateThumbnail(imgCopy, entity, thumbName, thumbWidth); err != nil && LogFunc != nil {
			LogFunc(fmt.Sprintf("warning: thumbnail failed for %s: %v", thumbName, err), 0, "")
		}
	}()

	// Metadata extraction
	go func() {
		if err := ExtractImageMetadata(imaging.Clone(img), generateUniqueID()); err != nil && LogFunc != nil {
			LogFunc(fmt.Sprintf("warning: metadata extraction failed for %s: %v", filepath.Base(fullPath), err), 0, "")
		}
	}()

	if LogFunc != nil {
		LogFunc(filepath.Base(fullPath), 0, "image/png")
	}
	return nil
}

func openImage(path string) (image.Image, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("open image: %w", err)
	}
	defer f.Close()
	img, format, err := image.Decode(f)
	return img, format, err
}

// -------------------------
// Thumbnail & Poster
// -------------------------

func generateThumbnail(img image.Image, entity EntityType, baseFilename string, thumbWidth int) error {
	resized := imaging.Resize(img, thumbWidth, 0, imaging.Lanczos)
	name := strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename)) + ".jpg"
	path := filepath.Join(ResolvePath(entity, PicThumb), name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create thumbnail: %w", err)
	}
	defer out.Close()
	if err := jpeg.Encode(out, resized, &jpeg.Options{Quality: defaultQuality}); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("encode thumbnail: %w", err)
	}
	if LogFunc != nil {
		LogFunc(path, 0, "image/jpeg")
	}
	return nil
}

func generateVideoPoster(videoPath string, entity EntityType, baseFilename string) (string, error) {
	thumbName := strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename)) + ".jpg"
	thumbDir := ResolvePath(entity, PicThumb)
	thumbPath := filepath.Join(thumbDir, thumbName)
	if err := os.MkdirAll(thumbDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", thumbDir, err)
	}

	ts := 0.5
	if out, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1", videoPath).Output(); err == nil {
		if d, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64); err == nil && d > 0 {
			if d >= 0.5 {
				ts = d / 2.0
			} else {
				ts = 0
			}
		}
	}

	ss := fmt.Sprintf("%.3f", ts)
	cmd := exec.Command("ffmpeg", "-y", "-i", videoPath, "-ss", ss, "-vframes", "1", thumbPath)
	if err := cmd.Run(); err != nil {
		fallback := exec.Command("ffmpeg", "-y", "-i", videoPath, "-ss", "0", "-vframes", "1", thumbPath)
		if ferr := fallback.Run(); ferr != nil {
			return "", fmt.Errorf("ffmpeg poster generation failed (primary: %v, fallback: %v)", err, ferr)
		}
	}

	if LogFunc != nil {
		LogFunc(thumbPath, 0, "image/jpeg")
	}
	return thumbName, nil
}

// -------------------------
// Image Normalization
// -------------------------

func normalizeImageFormat(fullPath, ext string, img image.Image) (string, error) {
	if ext == ".png" {
		return fullPath, nil
	}
	pngPath := strings.TrimSuffix(fullPath, ext) + ".png"
	out, err := os.Create(pngPath)
	if err != nil {
		return fullPath, fmt.Errorf("create png %s: %w", pngPath, err)
	}
	defer out.Close()
	if err := png.Encode(out, img); err != nil {
		_ = os.Remove(pngPath)
		return fullPath, fmt.Errorf("encode png: %w", err)
	}
	_ = os.Remove(fullPath)
	return pngPath, nil
}

// -------------------------
// File Validation & Writing
// -------------------------

func writeValidatedFile(reader io.Reader, header *multipart.FileHeader, destDir string, picType PictureType, maxSize int64) (string, string, string, error) {
	log.Println("->[writeValidatedFile] : no error yet")
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !isExtensionAllowed(ext, picType) {
		log.Println("[writeValidatedFile]->")
		return "", "", "", fmt.Errorf("%w: %s for %s", ErrInvalidExtension, ext, picType)
	}
	log.Println("->[writeValidatedFile] : no error yet")

	buf := make([]byte, 512)
	n, err := io.ReadFull(io.LimitReader(reader, 512), buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", "", "", fmt.Errorf("read header: %w", err)
	}

	mimeType := strings.ToLower(http.DetectContentType(buf[:n]))
	if mimeType == "application/octet-stream" {
		formMime := strings.ToLower(header.Header.Get("Content-Type"))
		if formMime != "" && isMIMEAllowed(formMime, picType) {
			mimeType = formMime
		}
	}

	if !isMIMEAllowed(mimeType, picType) {
		return "", "", "", fmt.Errorf("%w: %s for %s", ErrInvalidMIME, mimeType, picType)
	}
	if !extMatchesMIME(ext, mimeType, picType) {
		return "", "", "", fmt.Errorf("extension %s does not match MIME type %s for %s", ext, mimeType, picType)
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", "", "", fmt.Errorf("mkdir %s: %w", destDir, err)
	}

	// --- updated part ---
	filenameOnly, safeExt := getSafeFilename(header.Filename, ext, nil)
	fullPath := filepath.Join(destDir, filenameOnly+safeExt)
	// --- end update ---

	out, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return "", "", "", fmt.Errorf("create %s: %w", fullPath, err)
	}
	defer out.Close()

	if _, err := out.Write(buf[:n]); err != nil {
		return "", "", "", fmt.Errorf("write header: %w", err)
	}
	written, err := io.Copy(out, io.LimitReader(reader, maxSize-int64(n)))
	if err != nil {
		return "", "", "", fmt.Errorf("write body: %w", err)
	}
	totalWritten := written + int64(n)
	if maxSize > 0 && totalWritten > maxSize {
		_ = os.Remove(fullPath)
		return "", "", "", ErrFileTooLarge
	}

	if err := ScanForViruses(fullPath); err != nil {
		_ = os.Remove(fullPath)
		return "", "", "", fmt.Errorf("virus scan failed: %w", err)
	}

	if LogFunc != nil {
		LogFunc(filenameOnly+safeExt, totalWritten, mimeType)
	}
	log.Println("\t-------------------\t------------------\t--------------\t---------", filenameOnly, safeExt, fullPath)
	return filenameOnly, safeExt, fullPath, nil
}

// -------------------------
// Utilities
// -------------------------

func generateUniqueID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func isVideoExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".mp4", ".mov", ".mkv", ".webm", ".avi", ".flv", ".m4v":
		return true
	default:
		return false
	}
}
