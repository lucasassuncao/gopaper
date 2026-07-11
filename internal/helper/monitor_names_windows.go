//go:build windows

package helper

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// wmiMonitorID mirrors the fields we need from WMI's root\wmi WmiMonitorID
// class, populated from each connected monitor's EDID. Manufacturer/user
// friendly names are EDID char-code arrays (one ASCII byte per element,
// zero-padded), not strings.
type wmiMonitorID struct {
	InstanceName     string
	ManufacturerName []uint16
	UserFriendlyName []uint16
}

// monitorNames queries WmiMonitorID for a human-readable name per monitor,
// keyed by its PNP instance path (e.g.
// "DISPLAY\AUS32E0\5&19f84e22&1&UID4354_0") so callers can correlate it
// back to an IDesktopWallpaper device path — see matchMonitorName. Prefers
// EDID's UserFriendlyName (e.g. "ASUS VG32VQ1B"); falls back to the 3-letter
// manufacturer PNP ID (e.g. "BOE") when a monitor's EDID doesn't set one, as
// is common for laptop panels.
func monitorNames() (map[string]string, error) {
	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		"Get-CimInstance -Namespace root/wmi -ClassName WmiMonitorID | "+
			"Select-Object InstanceName,ManufacturerName,UserFriendlyName | ConvertTo-Json -Compress").Output()
	if err != nil {
		return nil, fmt.Errorf("could not query monitor EDID data via WMI: %w", err)
	}

	var entries []wmiMonitorID
	if err := json.Unmarshal(out, &entries); err != nil {
		// A single result serializes as a bare object, not a one-element array.
		var single wmiMonitorID
		if err2 := json.Unmarshal(out, &single); err2 != nil {
			return nil, fmt.Errorf("could not parse monitor EDID data: %w", err)
		}
		entries = []wmiMonitorID{single}
	}

	names := make(map[string]string, len(entries))
	for _, m := range entries {
		name := decodeEDIDString(m.UserFriendlyName)
		if name == "" {
			name = decodeEDIDString(m.ManufacturerName)
		}
		if name != "" {
			names[m.InstanceName] = name
		}
	}
	return names, nil
}

// decodeEDIDString converts an EDID char-code array (one ASCII byte per
// element, zero-terminated/padded) into a string. Elements outside byte
// range shouldn't occur in real EDID data, but are treated as the end of
// the string rather than truncated, so the uint16->byte narrowing below is
// always in range.
func decodeEDIDString(codes []uint16) string {
	b := make([]byte, 0, len(codes))
	for _, c := range codes {
		if c == 0 || c > 0xff {
			break
		}
		b = append(b, byte(c))
	}
	return strings.TrimSpace(string(b))
}
