package news

import (
	"testing"

	"wecheckin/backend/internal/model"
)

func TestPopulateFieldsUsesFirstImageFromJSONList(t *testing.T) {
	list := PopulateFields([]model.News{{Pic: `["/upload/a.png","/upload/b.png"]`}})

	if len(list) != 1 {
		t.Fatalf("expected one news item")
	}
	if list[0].Img != "http://localhost:8083/upload/a.png" {
		t.Fatalf("expected first image with default static domain, got %q", list[0].Img)
	}
}

func TestPopulateFieldsKeepsAbsoluteImageURL(t *testing.T) {
	list := PopulateFields([]model.News{{Pic: "https://cdn.example.com/a.png"}})

	if list[0].Img != "https://cdn.example.com/a.png" {
		t.Fatalf("absolute image URL should stay unchanged, got %q", list[0].Img)
	}
}
