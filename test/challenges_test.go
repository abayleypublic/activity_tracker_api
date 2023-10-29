package test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/AustinBayley/activity_tracker_api/pkg/challenges"
	"github.com/monzo/typhon"
)

func TestGetChallenges(t *testing.T) {
	ctx := context.Background()
	req := typhon.NewRequest(ctx, http.MethodGet, fmt.Sprintf("%s/challenges", baseURI), nil)
	res := req.Send().Response()

	if res.StatusCode != http.StatusOK {
		if res.Error != nil {
			t.Fatalf("Failed to get challenges: %v", res.Error)
		}

		t.Fatalf("Failed to get challenges: %v", res.StatusCode)
	}

	c := []challenges.Challenge{}
	if err := res.Decode(&c); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(c) != 1 {
		t.Fatalf("Expected 1 challenge, got %d", len(c))
	}
}
