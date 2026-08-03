package dingtalkh5

import (
	"testing"

	"wecheckin/backend/internal/model"
)

func cardValue(cards []WorkbenchStatCardDTO, key string) int {
	for _, card := range cards {
		if card.Key == key {
			return card.Value
		}
	}
	return -1
}

func TestWorkbenchStatsFromReviewsReturnsCardsOnly(t *testing.T) {
	user := &model.DingTalkH5PerfUser{Account: "lip"}
	stats := workbenchStatsFromReviews(user, []model.DingTalkH5PerfReview{
		{EmployeeAccount: "lip", Status: ReviewStatusDraft},
		{EmployeeAccount: "lip", Status: ReviewStatusCompleted},
		{EmployeeAccount: "cube", ManagerAccount: "lip", Status: ReviewStatusManagerReview},
	})

	if len(stats.Cards) != 5 {
		t.Fatalf("workbench cards = %d, want 5", len(stats.Cards))
	}
	if cardValue(stats.Cards, "queue") != 2 {
		t.Fatalf("queue card = %d, want 2", cardValue(stats.Cards, "queue"))
	}
	if cardValue(stats.Cards, "all") != 3 {
		t.Fatalf("all card = %d, want 3", cardValue(stats.Cards, "all"))
	}
	if cardValue(stats.Cards, "completed") != 1 {
		t.Fatalf("completed card = %d, want 1", cardValue(stats.Cards, "completed"))
	}
}
