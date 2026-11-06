package tracking

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Fix is one GPS point reported by a phone.
type Fix struct {
	Latitude   float64
	Longitude  float64
	RecordedAt time.Time
	SpeedKmh   *float32
	Heading    *float32
	AltitudeM  *float32
	AccuracyM  *float32
	BatteryPct *float32
	IsMoving   *bool
}

const maxBodyBytes = 1 << 20 // 1 MiB; enough for a large offline batch

var errNoDeviceID = errors.New("missing device id")

// parseTraccarRequest understands both formats Traccar Client has used:
//
//   - Current apps POST JSON: {"device_id": "...", "location": {...}} where
//     "location" may also be an array when the app flushes an offline batch.
//   - Older apps (and the OsmAnd protocol) send the values as URL query or
//     form parameters: id, lat, lon, timestamp, speed (knots), bearing, ...
func parseTraccarRequest(r *http.Request) (string, []Fix, error) {
	mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))

	if r.Method == http.MethodPost && mediaType == "application/json" {
		body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
		if err != nil {
			return "", nil, err
		}
		if len(body) > maxBodyBytes {
			return "", nil, errors.New("request body too large")
		}
		return parseJSON(body)
	}

	values := r.URL.Query()
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(nil, r.Body, maxBodyBytes)
		if err := r.ParseForm(); err == nil {
			for k, v := range r.PostForm {
				values[k] = v
			}
		}
	}
	return parseQuery(values)
}

type jsonCoords struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Accuracy  *float64 `json:"accuracy"`
	Speed     *float64 `json:"speed"` // metres per second, -1 = unknown
	Heading   *float64 `json:"heading"`
	Altitude  *float64 `json:"altitude"`
}

type jsonLocation struct {
	Timestamp string     `json:"timestamp"`
	Coords    jsonCoords `json:"coords"`
	IsMoving  *bool      `json:"is_moving"`
	Battery   *struct {
		Level *float64 `json:"level"` // 0..1, -1 = unknown
	} `json:"battery"`
}

type jsonPayload struct {
	DeviceID string          `json:"device_id"`
	Location json.RawMessage `json:"location"`
}

func parseJSON(body []byte) (string, []Fix, error) {
	var p jsonPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return "", nil, fmt.Errorf("invalid json: %w", err)
	}
	if strings.TrimSpace(p.DeviceID) == "" {
		return "", nil, errNoDeviceID
	}

	var locs []jsonLocation
	trimmed := strings.TrimSpace(string(p.Location))
	switch {
	case strings.HasPrefix(trimmed, "["):
		if err := json.Unmarshal(p.Location, &locs); err != nil {
			return "", nil, fmt.Errorf("invalid location list: %w", err)
		}
	case strings.HasPrefix(trimmed, "{"):
		var one jsonLocation
		if err := json.Unmarshal(p.Location, &one); err != nil {
			return "", nil, fmt.Errorf("invalid location: %w", err)
		}
		locs = []jsonLocation{one}
	default:
		return "", nil, errors.New("missing location")
	}

	fixes := make([]Fix, 0, len(locs))
	for _, l := range locs {
		if l.Coords.Latitude == nil || l.Coords.Longitude == nil {
			continue
		}
		ts, err := parseTimestamp(l.Timestamp)
		if err != nil {
			continue
		}
		f := Fix{
			Latitude:   *l.Coords.Latitude,
			Longitude:  *l.Coords.Longitude,
			RecordedAt: ts,
			Heading:    nonNegative(l.Coords.Heading, 1),
			AccuracyM:  nonNegative(l.Coords.Accuracy, 1),
			AltitudeM:  f32(l.Coords.Altitude),
			SpeedKmh:   nonNegative(l.Coords.Speed, 3.6),
			IsMoving:   l.IsMoving,
		}
		if l.Battery != nil {
			f.BatteryPct = nonNegative(l.Battery.Level, 100)
		}
		fixes = append(fixes, f)
	}
	return strings.TrimSpace(p.DeviceID), fixes, nil
}

func parseQuery(v url.Values) (string, []Fix, error) {
	deviceID := strings.TrimSpace(v.Get("id"))
	if deviceID == "" {
		deviceID = strings.TrimSpace(v.Get("deviceid"))
	}
	if deviceID == "" {
		return "", nil, errNoDeviceID
	}

	latStr, lonStr := v.Get("lat"), v.Get("lon")
	if loc := v.Get("location"); loc != "" && (latStr == "" || lonStr == "") {
		if parts := strings.SplitN(loc, ",", 2); len(parts) == 2 {
			latStr, lonStr = parts[0], parts[1]
		}
	}
	lat, errLat := strconv.ParseFloat(strings.TrimSpace(latStr), 64)
	lon, errLon := strconv.ParseFloat(strings.TrimSpace(lonStr), 64)
	if errLat != nil || errLon != nil {
		return "", nil, errors.New("missing or invalid lat/lon")
	}
	if valid := v.Get("valid"); valid == "false" || valid == "0" {
		return deviceID, nil, nil
	}

	ts := time.Now().UTC()
	if raw := v.Get("timestamp"); raw != "" {
		parsed, err := parseTimestamp(raw)
		if err != nil {
			return "", nil, err
		}
		ts = parsed
	}

	f := Fix{Latitude: lat, Longitude: lon, RecordedAt: ts}
	f.SpeedKmh = nonNegative(queryFloat(v, "speed"), 1.852) // knots -> km/h
	if h := queryFloat(v, "bearing"); h != nil {
		f.Heading = nonNegative(h, 1)
	} else {
		f.Heading = nonNegative(queryFloat(v, "heading"), 1)
	}
	f.AltitudeM = f32(queryFloat(v, "altitude"))
	f.AccuracyM = nonNegative(queryFloat(v, "accuracy"), 1)
	if b := queryFloat(v, "batt"); b != nil {
		f.BatteryPct = nonNegative(b, 1)
	} else {
		f.BatteryPct = nonNegative(queryFloat(v, "battery"), 1)
	}
	return deviceID, []Fix{f}, nil
}

// parseTimestamp accepts epoch seconds, epoch milliseconds, RFC 3339 and
// "2006-01-02 15:04:05" (UTC).
func parseTimestamp(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, errors.New("empty timestamp")
	}
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		if n > 1e11 { // milliseconds
			return time.UnixMilli(n).UTC(), nil
		}
		return time.Unix(n, 0).UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05", raw); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, fmt.Errorf("unrecognised timestamp %q", raw)
}

func queryFloat(v url.Values, key string) *float64 {
	raw := strings.TrimSpace(v.Get(key))
	if raw == "" {
		return nil
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &f
}

func f32(v *float64) *float32 {
	if v == nil {
		return nil
	}
	f := float32(*v)
	return &f
}

// nonNegative scales v, treating negative values (the apps' "unknown") as nil.
func nonNegative(v *float64, scale float64) *float32 {
	if v == nil || *v < 0 {
		return nil
	}
	f := float32(*v * scale)
	return &f
}
