package application

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestWorkflowSearchRejectsLongStarterName(t *testing.T) {
	service := &Service{store: &fakeStore{}}
	keyword := strings.Repeat("名", 51)

	if _, err := service.ListInstances(context.Background(), InstanceQuery{StarterName: keyword}); !errors.Is(err, ErrStarterNameSearchTooLong) {
		t.Fatalf("ListInstances error = %v, want %v", err, ErrStarterNameSearchTooLong)
	}
	if _, err := service.ListTasks(context.Background(), TaskQuery{StarterName: keyword}); !errors.Is(err, ErrStarterNameSearchTooLong) {
		t.Fatalf("ListTasks error = %v, want %v", err, ErrStarterNameSearchTooLong)
	}
}

func TestWorkflowSearchRejectsLongDefinitionName(t *testing.T) {
	service := &Service{store: &fakeStore{}}
	keyword := strings.Repeat("流", 51)

	if _, err := service.ListInstances(context.Background(), InstanceQuery{DefinitionName: keyword}); !errors.Is(err, ErrDefinitionNameSearchTooLong) {
		t.Fatalf("ListInstances error = %v, want %v", err, ErrDefinitionNameSearchTooLong)
	}
	if _, err := service.ListTasks(context.Background(), TaskQuery{DefinitionName: keyword}); !errors.Is(err, ErrDefinitionNameSearchTooLong) {
		t.Fatalf("ListTasks error = %v, want %v", err, ErrDefinitionNameSearchTooLong)
	}
}

func TestListPublishedDefinitionCategoriesReturnsSortedUniqueValues(t *testing.T) {
	service := &Service{store: &fakeStore{publishedDefinitions: []PublishedDefinition{
		{Category: " hr "},
		{Category: "finance"},
		{Category: "hr"},
		{Category: " "},
	}}}

	categories, err := service.ListPublishedDefinitionCategories(context.Background())
	if err != nil {
		t.Fatalf("ListPublishedDefinitionCategories() error = %v", err)
	}
	if got, want := strings.Join(categories, ","), "finance,hr"; got != want {
		t.Fatalf("categories = %q, want %q", got, want)
	}
}
