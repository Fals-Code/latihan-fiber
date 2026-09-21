package helper

import (
	"testing"
	"time"
)

func TestCursorRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	encoded, err := EncodeCursor(now, 7)
	if err != nil {
		t.Fatalf("encode cursor failed: %v", err)
	}
	gotTime, gotID, err := DecodeCursor(encoded)
	if err != nil || !gotTime.Equal(now) || gotID != 7 {
		t.Fatalf("cursor round trip failed: %v %v %v", gotTime, gotID, err)
	}
}

func TestDecodeCursorRejectsInvalid(t *testing.T) {
	if _, _, err := DecodeCursor("bad"); err == nil {
		t.Fatal("expected invalid cursor error")
	}
}
