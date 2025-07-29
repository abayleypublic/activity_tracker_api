package targets_test

import (
	"encoding/json"
	"testing"

	"github.com/AustinBayley/activity_tracker_api/pkg/targets"
)

func TestMarshal(t *testing.T) {
	target := targets.RouteMovingTarget{
		BaseTarget: targets.BaseTarget{
			TargetType: targets.RouteMovingTargetType,
		},
	}

	data, err := json.Marshal(&target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var raw targets.RawTarget
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if raw.TargetType != targets.RouteMovingTargetType {
		t.Errorf("expected target type %s, got %s", targets.RouteMovingTargetType, raw.TargetType)
	}

	if raw.RealTarget == nil {
		t.Error("expected RealTarget to be set")
	}

	if _, ok := raw.RealTarget.(*targets.RouteMovingTarget); !ok {
		t.Error("expected RealTarget to be of type RouteMovingTarget")
	}
}
