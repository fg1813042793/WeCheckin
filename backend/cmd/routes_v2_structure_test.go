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
	if !strings.Contains(string(src), "routes.Register(h)") {
		t.Fatalf("registerRoutes must delegate to internal route registration")
	}
	v2Src, err := os.ReadFile("../internal/routes/v2/routes.go")
	if err != nil {
		t.Fatalf("read internal/routes/v2/routes.go: %v", err)
	}
	for _, want := range []string{
		"clientroutes.Register(h)",
		"adminroutes.Register(h)",
		"dingtalkh5routes.Register(h)",
	} {
		if !strings.Contains(string(v2Src), want) {
			t.Fatalf("v2 route suite missing %s", want)
		}
	}
}

func TestV2AdminRoutesExposeRESTfulResources(t *testing.T) {
	src, err := os.ReadFile("../internal/routes/v2/admin/routes.go")
	if err != nil {
		t.Fatalf("read admin v2 routes: %v", err)
	}
	text := string(src)
	required := []string{
		`h.Group("/api/v2/admin", adminmw.AdminAuth(), adminmw.AdminPerm())`,
		`admin.GET("/users", aUser.GetUserList)`,
		`admin.POST("/users", aUser.AddUser)`,
		`admin.GET("/users/:id", routeparam.WithQueryID(aUser.GetUserByID))`,
		`admin.PUT("/users/:id", routeparam.WithFormID(aUser.EditUser))`,
		`admin.DELETE("/users/:id", routeparam.WithFormID(aUser.DelUser))`,
		`admin.GET("/surveys", aSurvey.List)`,
		`admin.POST("/surveys", aSurvey.Insert)`,
		`admin.PUT("/surveys/:id", routeparam.WithBodyOrFormID(aSurvey.Edit))`,
		`admin.GET("/exams", aExam.List)`,
		`admin.POST("/exams", aExam.Save)`,
		`admin.PUT("/exams/:id", routeparam.WithBodyOrFormID(aExam.Save))`,
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			t.Fatalf("admin v2 routes missing %s", want)
		}
	}
}

func TestV2ClientRoutesExposeRESTfulResources(t *testing.T) {
	src, err := os.ReadFile("../internal/routes/v2/client/routes.go")
	if err != nil {
		t.Fatalf("read client v2 routes: %v", err)
	}
	text := string(src)
	required := []string{
		`h.GET("/api/v2/home", hm.GetHomeList)`,
		`h.POST("/api/v2/auth/login", pp.Login)`,
		`h.GET("/api/v2/surveys", cSurvey.List)`,
		`h.GET("/api/v2/exams", cExam.List)`,
		`client.GET("/me", pp.GetMyDetail)`,
		`client.GET("/news", ns.GetNewsList)`,
		`client.GET("/enrollments/:id", routeparam.WithQueryID(el.ViewEnroll))`,
		`client.POST("/enrollments/:id/joins", routeparam.WithFormParam("enroll_id", "id", el.EnrollJoin))`,
		`client.POST("/events/:id/participants", routeparam.WithFormParam("event_id", "id", ev.EventParticipate))`,
		`client.POST("/exams/:id/start", routeparam.WithQueryParam("examId", "id", cExam.Start))`,
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			t.Fatalf("client v2 routes missing %s", want)
		}
	}
}

func TestSwaggerIncludesAllV2RouteOperations(t *testing.T) {
	routeFiles := []string{
		"../internal/routes/v2/admin/routes.go",
		"../internal/routes/v2/client/routes.go",
	}
	var routeSrc strings.Builder
	for _, path := range routeFiles {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		routeSrc.Write(src)
	}
	routeText := routeSrc.String()
	want := strings.Count(routeText, `.GET("`) +
		strings.Count(routeText, `.POST("`) +
		strings.Count(routeText, `.PUT("`) +
		strings.Count(routeText, `.DELETE("`) +
		strings.Count(routeText, `.PATCH("`)

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
		t.Fatalf("swagger v2 operation count = %d, want %d; rerun swag init after changing v2 routes", got, want)
	}
}
