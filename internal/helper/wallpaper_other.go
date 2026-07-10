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
