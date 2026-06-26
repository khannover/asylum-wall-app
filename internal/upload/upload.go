package upload

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
)

const MaxProofSize = 5 * 1024 * 1024

var allowedExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
	".pdf":  true,
	".eml":  true,
}

var allowedMIMEs = map[string]bool{
	"image/png":       true,
	"image/jpeg":      true,
	"image/gif":       true,
	"image/webp":      true,
	"application/pdf": true,
	"message/rfc822":  true,
	"text/plain":      true, // some .eml files are served as text/plain
}

func SanitizeExtension(filename string) (string, error) {
	base := filepath.Base(filename)
	if base == "." || base == ".." || base == "" {
		return "", fmt.Errorf("invalid filename")
	}

	ext := strings.ToLower(filepath.Ext(base))
	if ext == "" {
		return "", fmt.Errorf("file must have an extension")
	}
	if !allowedExtensions[ext] {
		return "", fmt.Errorf("unsupported file type: %s (allowed: png, jpg, jpeg, gif, webp, pdf, eml)", ext)
	}

	return ext, nil
}

func ValidateContent(data []byte, ext string) error {
	if len(data) == 0 {
		return fmt.Errorf("empty file")
	}
	if int64(len(data)) > MaxProofSize {
		return fmt.Errorf("file exceeds maximum size of 5MB")
	}

	detected := mime.TypeByExtension(ext)
	if detected != "" {
		detected = strings.Split(detected, ";")[0]
		if !allowedMIMEs[detected] {
			return fmt.Errorf("extension %s does not match allowed MIME types", ext)
		}
	}

	switch ext {
	case ".png":
		if !bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
			return fmt.Errorf("file content does not match PNG format")
		}
	case ".jpg", ".jpeg":
		if !bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
			return fmt.Errorf("file content does not match JPEG format")
		}
	case ".gif":
		if !bytes.HasPrefix(data, []byte("GIF87a")) && !bytes.HasPrefix(data, []byte("GIF89a")) {
			return fmt.Errorf("file content does not match GIF format")
		}
	case ".pdf":
		if !bytes.HasPrefix(data, []byte("%PDF-")) {
			return fmt.Errorf("file content does not match PDF format")
		}
	case ".webp":
		if len(data) < 12 || !bytes.HasPrefix(data, []byte("RIFF")) || !bytes.Equal(data[8:12], []byte("WEBP")) {
			return fmt.Errorf("file content does not match WebP format")
		}
	case ".eml":
		// EML is plain text RFC822; reject obvious script payloads
		lower := strings.ToLower(string(data[:min(512, len(data))]))
		if strings.Contains(lower, "<script") || strings.Contains(lower, "<?php") {
			return fmt.Errorf("suspicious content in EML file")
		}
	}

	return nil
}

func ReadLimited(r io.Reader, max int64) ([]byte, error) {
	limited := io.LimitReader(r, max+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read upload: %w", err)
	}
	if int64(len(data)) > max {
		return nil, fmt.Errorf("file exceeds maximum size of 5MB")
	}
	return data, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}