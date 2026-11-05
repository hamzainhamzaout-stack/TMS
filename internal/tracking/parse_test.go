package tracking

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParseJSONSingleLocation(t *testing.T) {
	body := `{"device_id":"abc123","location":{"timestamp":"2026-11-05T10:00:00.000Z","is_moving":true,
		"coords":{"latitude":9.03,"longitude":38.74,"accuracy":12.5,"speed":10,"heading":-1,"altitude":2350},
		"battery":{"level":0.82,"is_charging":false}}}`
	r := httptest.NewRequest(http.MethodPost, "/track", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json; charset=utf-8")

	id, fixes, err := parseTraccarRequest(r)
	if err != nil {
		t.Fatal(err)
	}
	if id != "abc123" || len(fixes) != 1 {
		t.Fatalf("id=%q fixes=%d", id, len(fixes))
	}
	f := fixes[0]
	if f.Latitude != 9.03 || f.Longitude != 38.74 {
		t.Errorf("coords = %v,%v", f.Latitude, f.Longitude)
	}
	if !f.RecordedAt.Equal(time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)) {
		t.Errorf("time = %v", f.RecordedAt)
	}
	if f.SpeedKmh == nil || *f.SpeedKmh != 36 {
		t.Errorf("speed = %v, want 36 km/h", f.SpeedKmh)
	}
	if f.Heading != nil {
		t.Errorf("heading -1 should be unknown, got %v", *f.Heading)
	}
	if f.BatteryPct == nil || *f.BatteryPct < 81.9 || *f.BatteryPct > 82.1 {
		t.Errorf("battery = %v", f.BatteryPct)
	}
}

func TestParseJSONBatch(t *testing.T) {
	body := `{"device_id":"abc","location":[
		{"timestamp":"2026-11-05T10:00:00Z","coords":{"latitude":9,"longitude":38}},
		{"timestamp":"2026-11-05T10:01:00Z","coords":{"latitude":9.1,"longitude":38.1}}]}`
	r := httptest.NewRequest(http.MethodPost, "/track", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	_, fixes, err := parseTraccarRequest(r)
	if err != nil || len(fixes) != 2 {
		t.Fatalf("fixes=%d err=%v", len(fixes), err)
	}
}

func TestParseQueryOsmAnd(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost,
		"/track?id=dev1&lat=9.03&lon=38.74&timestamp=1793872800&speed=10&bearing=90&batt=55&accuracy=8", nil)
	id, fixes, err := parseTraccarRequest(r)
	if err != nil {
		t.Fatal(err)
	}
	if id != "dev1" || len(fixes) != 1 {
		t.Fatalf("id=%q fixes=%d", id, len(fixes))
	}
	f := fixes[0]
	if f.SpeedKmh == nil || *f.SpeedKmh < 18.5 || *f.SpeedKmh > 18.6 {
		t.Errorf("speed = %v, want 18.52 km/h (10 knots)", f.SpeedKmh)
	}
	if f.RecordedAt.Unix() != 1793872800 {
		t.Errorf("time = %v", f.RecordedAt)
	}
	if *f.Heading != 90 || *f.BatteryPct != 55 || *f.AccuracyM != 8 {
		t.Errorf("fix = %+v", f)
	}
}

func TestParseQueryMillisAndLocationParam(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/track?deviceid=d&location=9.5,38.5&timestamp=1793872800123", nil)
	_, fixes, err := parseTraccarRequest(r)
	if err != nil || len(fixes) != 1 {
		t.Fatalf("fixes=%d err=%v", len(fixes), err)
	}
	if fixes[0].Latitude != 9.5 || fixes[0].RecordedAt.UnixMilli() != 1793872800123 {
		t.Errorf("fix = %+v", fixes[0])
	}
}

func TestParseRejectsMissingDevice(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/track?lat=1&lon=2", nil)
	if _, _, err := parseTraccarRequest(r); err == nil {
		t.Fatal("expected error")
	}
}

func TestRejectReason(t *testing.T) {
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	bad := float32(5000)
	cases := map[string]Fix{
		"coordinates out of range": {Latitude: 95, Longitude: 38, RecordedAt: now},
		"null island":              {Latitude: 0, Longitude: 0, RecordedAt: now},
		"timestamp in the future":  {Latitude: 9, Longitude: 38, RecordedAt: now.Add(time.Hour)},
		"timestamp too old":        {Latitude: 9, Longitude: 38, RecordedAt: now.Add(-8 * 24 * time.Hour)},
		"accuracy too low":         {Latitude: 9, Longitude: 38, RecordedAt: now, AccuracyM: &bad},
		"":                         {Latitude: 9, Longitude: 38, RecordedAt: now.Add(-time.Hour)},
	}
	for want, f := range cases {
		if got := rejectReason(f, now); got != want {
			t.Errorf("rejectReason(%+v) = %q, want %q", f, got, want)
		}
	}
}

func TestDeviceToken(t *testing.T) {
	a, _ := newDeviceToken()
	b, _ := newDeviceToken()
	if len(a) != tokenLength || a == b || hashToken(a) == hashToken(b) {
		t.Fatalf("tokens %q %q", a, b)
	}
}
