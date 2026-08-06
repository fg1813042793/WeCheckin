package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestV2RoutesAreRegistered(t *testing.T) {
	src, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatalf("read routes.go: %v", err)
	}
	if !strings.Contains(string(src), "registerV2Routes(h)") {
		t.Fatalf("registerRoutes must register the /api/v2 route suite")
	}
}

func TestV2AdminRoutesExposeRESTfulResources(t *testing.T) {
	src, err := os.ReadFile("routes_v2.go")
	if err != nil {
		t.Fatalf("read routes_v2.go: %v", err)
	}
	text := string(src)
	required := []string{
		`h.Group("/api/v2/admin", adminmw.AdminAuth(), adminmw.AdminPerm())`,
		`admin.GET("/users", aUser.GetUserList)`,
		`admin.POST("/users", aUser.AddUser)`,
		`admin.GET("/users/:id", withQueryID(aUser.GetUserByID))`,
		`admin.PUT("/users/:id", withFormID(aUser.EditUser))`,
		`admin.DELETE("/users/:id", withFormID(aUser.DelUser))`,
		`admin.GET("/surveys", aSurvey.List)`,
		`admin.POST("/surveys", aSurvey.Insert)`,
		`admin.PUT("/surveys/:id", withBodyOrFormID(aSurvey.Edit))`,
		`admin.GET("/exams", aExam.List)`,
		`admin.POST("/exams", aExam.Save)`,
		`admin.PUT("/exams/:id", withBodyOrFormID(aExam.Save))`,
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			t.Fatalf("routes_v2.go missing %s", want)
		}
	}
}

func TestV2ClientRoutesExposeRESTfulResources(t *testing.T) {
	src, err := os.ReadFile("routes_v2.go")
	if err != nil {
		t.Fatalf("read routes_v2.go: %v", err)
	}
	text := string(src)
	required := []string{
		`h.GET("/api/v2/home", hm.GetHomeList)`,
		`h.POST("/api/v2/auth/login", pp.Login)`,
		`h.GET("/api/v2/surveys", cSurvey.List)`,
		`h.GET("/api/v2/exams", cExam.List)`,
		`client.GET("/me", pp.GetMyDetail)`,
		`client.GET("/news", ns.GetNewsList)`,
		`client.GET("/enrollments/:id", withQueryID(el.ViewEnroll))`,
		`client.POST("/enrollments/:id/joins", withFormParam("enroll_id", "id", el.EnrollJoin))`,
		`client.POST("/events/:id/participants", withFormParam("event_id", "id", ev.EventParticipate))`,
		`client.POST("/exams/:id/start", withQueryParam("examId", "id", cExam.Start))`,
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			t.Fatalf("routes_v2.go missing %s", want)
		}
	}
}

func TestSwaggerIncludesAllV2RouteOperations(t *testing.T) {
	routeSrc, err := os.ReadFile("routes_v2.go")
	if err != nil {
		t.Fatalf("read routes_v2.go: %v", err)
	}
	want := strings.Count(string(routeSrc), `.GET("`) +
		strings.Count(string(routeSrc), `.POST("`) +
		strings.Count(string(routeSrc), `.PUT("`) +
		strings.Count(string(routeSrc), `.DELETE("`) +
		strings.Count(string(routeSrc), `.PATCH("`)

	docSrc, err := os.ReadFile("../docs/swagger/swagger.json")
	if err != nil {
		t.Fatalf("read swagger.json: %v", err)
	}
	var doc struct {
		Paths map[string]map[string]interface{} `json:"paths"`
	}
	if err := json.Unmarshal(docSrc, &doc); err != nil {
		t.Fatalf("parse swagger.json: %v", err)
	}
	methods := map[string]struct{}{
		"get": {}, "post": {}, "put": {}, "delete": {}, "patch": {},
	}
	got := 0
	for path, ops := range doc.Paths {
		if !strings.HasPrefix(path, "/api/v2") {
			continue
		}
		for method := range ops {
			if _, ok := methods[method]; ok {
				got++
			}
		}
	}
	if got != want {
		t.Fatalf("swagger v2 operation count = %d, want %d; rerun swag init after changing routes_v2.go", got, want)
	}
}
