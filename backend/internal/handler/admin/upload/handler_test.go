package upload

import "testing"

func TestAllowedUploadExtension(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want bool
	}{
		{name: "avatar.JPG", want: true},
		{name: "cover.webp", want: true},
		{name: "intro.mp4", want: true},
		{name: "script.svg", want: false},
		{name: "payload.html", want: false},
		{name: "missing-extension", want: false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := allowedUploadExtension(tc.name); got != tc.want {
				t.Fatalf("allowedUploadExtension(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestValidateUploadContent(t *testing.T) {
	t.Parallel()
	pngHeader := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0, 'I', 'H', 'D', 'R'}
	if !validateUploadContent("avatar.png", pngHeader) {
		t.Fatal("valid PNG content should be accepted")
	}
	if validateUploadContent("avatar.png", []byte("<html><script>alert(1)</script></html>")) {
		t.Fatal("HTML content disguised as a PNG must be rejected")
	}
	if validateUploadContent("payload.svg", []byte("<svg></svg>")) {
		t.Fatal("unsupported extensions must be rejected even when the content is an image")
	}
}
