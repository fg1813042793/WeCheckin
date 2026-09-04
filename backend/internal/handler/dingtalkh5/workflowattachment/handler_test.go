package workflowattachment

import "testing"

func TestAllowedAttachmentExtension(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want bool
	}{
		{name: "photo.JPG", want: true},
		{name: "report.pdf", want: true},
		{name: "sheet.xlsx", want: true},
		{name: "archive.zip", want: true},
		{name: "script.svg", want: false},
		{name: "page.html", want: false},
		{name: "program.exe", want: false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := allowedAttachmentExtension(tc.name); got != tc.want {
				t.Fatalf("allowedAttachmentExtension(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestValidateAttachmentContent(t *testing.T) {
	t.Parallel()
	if !validateAttachmentContent("report.pdf", []byte("%PDF-1.7\n")) {
		t.Fatal("valid PDF content should be accepted")
	}
	if validateAttachmentContent("report.pdf", []byte("<html><script>alert(1)</script></html>")) {
		t.Fatal("HTML content disguised as a PDF must be rejected")
	}
	if !validateAttachmentContent("report.docx", []byte{'P', 'K', 0x03, 0x04}) {
		t.Fatal("OOXML zip content should be accepted")
	}
	if !validateAttachmentContent("notes.txt", []byte("plain text")) {
		t.Fatal("plain text content should be accepted")
	}
}
