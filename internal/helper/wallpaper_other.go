//go:build !windows

package helper

import "errors"

// setWallpaperFade and setWallpaperPosition back the fade transition, which
// is only implemented on Windows via IDesktopWallpaper. Returning an error
// here makes SetWallpaperFromPath/SetWallpaperMode fall back to the
// standard cross-platform behavior unchanged.
func setWallpaperFade(fullPath string) error {
	return errors.New("wallpaper fade transition is only supported on Windows")
}

func setWallpaperPosition(mode string) error {
	return errors.New("wallpaper fade transition is only supported on Windows")
}

// monitorDevicePaths and setWallpaperOnMonitor back the per-monitor mode,
// which relies on IDesktopWallpaper and is therefore Windows-only. The
// caller treats an error here as "fall back to the single-wallpaper flow".
func monitorDevicePaths() ([]string, error) {
	return nil, errors.New("per-monitor wallpapers are only supported on Windows")
}

func setWallpaperOnMonitor(devicePath, fullPath string) error {
	return errors.New("per-monitor wallpapers are only supported on Windows")
}

func monitorDetails() ([]MonitorDetail, error) {
	return nil, errors.New("monitor enumeration is only supported on Windows")
}

func monitorNames() (map[string]string, error) {
	return nil, errors.New("monitor enumeration is only supported on Windows")
}
