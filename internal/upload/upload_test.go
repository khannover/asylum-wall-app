package upload

import (
	"bytes"
	"strings"
	"testing"
)

func TestSanitizeExtension(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid png", "proof.png", ".png", false},
		{"valid pdf", "doc.PDF", ".pdf", false},
		{"path traversal", "../../../etc/passwd.png", ".png", false},
		{"no extension", "file", "", true},
		{"disallowed exe", "malware.exe", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeExtension(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateContentPNG(t *testing.T) {
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if err := ValidateContent(pngHeader, ".png"); err != nil {
		t.Fatalf("valid png header rejected: %v", err)
	}

	if err := ValidateContent([]byte("not a png"), ".png"); err == nil {
		t.Fatal("invalid png accepted")
	}
}

func TestValidateContentEMLRejectsScript(t *testing.T) {
	eml := []byte("Subject: test\n\n<script>alert(1)</script>")
	if err := ValidateContent(eml, ".eml"); err == nil {
		t.Fatal("script in eml should be rejected")
	}
}

func TestReadLimited(t *testing.T) {
	data := bytes.Repeat([]byte("a"), 100)
	got, err := ReadLimited(strings.NewReader(string(data)), 200)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 100 {
		t.Fatalf("got len %d, want 100", len(got))
	}

	_, err = ReadLimited(strings.NewReader(string(bytes.Repeat([]byte("b"), 300))), 200)
	if err == nil {
		t.Fatal("expected size limit error")
	}
}