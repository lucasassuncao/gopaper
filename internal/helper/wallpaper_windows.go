//go:build windows

package helper

import (
	"errors"
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/reujab/wallpaper"
	"golang.org/x/sys/windows"
)

// Explorer switches into slideshow mode asynchronously after
// AdvanceSlideshow returns, anywhere from under a second to a few seconds
// later. These bound how long setWallpaperFade polls GetStatus while
// waiting for that flip and while verifying its static-picture pin stuck.
const (
	slideshowFlipTimeout = 4 * time.Second
	pinVerifyWindow      = 2500 * time.Millisecond
	statusPollInterval   = 250 * time.Millisecond
)

// guid mirrors the binary layout of the Win32 GUID struct.
type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

var (
	clsidDesktopWallpaper = guid{0xC2CF3110, 0x460E, 0x4fc1, [8]byte{0xB9, 0xD0, 0x8A, 0x1C, 0x0C, 0x9C, 0xC4, 0xBD}}
	iidIDesktopWallpaper  = guid{0xB92B56A9, 0x8B55, 0x4E14, [8]byte{0x9A, 0x89, 0x01, 0x99, 0xBB, 0xB6, 0xF9, 0x3B}}
)

// IDesktopWallpaper vtable slots (after the 3 IUnknown slots).
const (
	idwSetWallpaper              = 3
	idwGetWallpaper              = 4
	idwGetMonitorDevicePathAt    = 5
	idwGetMonitorDevicePathCount = 6
	idwSetPosition               = 10
	idwSetSlideshow              = 12
	idwSetSlideshowOptions       = 14
	idwAdvanceSlideshow          = 16
	idwGetStatus                 = 17
)

const (
	clsctxLocalServer       = 0x4
	coinitApartmentThreaded = 0x2
	dsdForward              = 0
	dssSlideshow            = 0x2 // DESKTOP_SLIDESHOW_STATE: slideshow mode active
	hrEUnexpected           = 0x8000FFFF

	dwposCenter  = 0
	dwposTile    = 1
	dwposStretch = 2
	dwposFit     = 3
	dwposFill    = 4
	dwposSpan    = 5
)

var (
	ole32   = syscall.NewLazyDLL("ole32.dll")
	shell32 = syscall.NewLazyDLL("shell32.dll")

	procCoInitializeEx                    = ole32.NewProc("CoInitializeEx")
	procCoUninitialize                    = ole32.NewProc("CoUninitialize")
	procCoCreateInstance                  = ole32.NewProc("CoCreateInstance")
	procCoTaskMemFree                     = ole32.NewProc("CoTaskMemFree")
	procSHParseDisplayName                = shell32.NewProc("SHParseDisplayName")
	procSHCreateShellItemArrayFromIDLists = shell32.NewProc("SHCreateShellItemArrayFromIDLists")
)

func coUninitialize() {
	_, _, _ = procCoUninitialize.Call() //nolint:dogsled // stdcall with no return value
}

func coTaskMemFree(p uintptr) {
	_, _, _ = procCoTaskMemFree.Call(p) //nolint:dogsled // stdcall with no return value
}

// hresultError is a failed COM/shell call, keeping the raw HRESULT so
// callers can react to specific codes.
type hresultError struct {
	op   string
	code uint32
}

func (e *hresultError) Error() string {
	return fmt.Sprintf("%s failed: HRESULT 0x%08X", e.op, e.code)
}

func hrErr(name string, r uintptr) error {
	if int32(r) < 0 { //#nosec G115 -- r is a raw stdcall return register; an HRESULT always fits in the low 32 bits, so truncating to check its sign bit is correct, not an overflow
		return &hresultError{op: name, code: uint32(r)} //#nosec G115 -- same HRESULT truncation as above
	}
	return nil
}

// isHResult reports whether err is a COM failure with the given HRESULT.
func isHResult(err error, code uint32) bool {
	var he *hresultError
	return errors.As(err, &he) && he.code == code
}

// vtblCall invokes the COM method at the given vtable slot on `this`.
func vtblCall(this unsafe.Pointer, index int, args ...uintptr) error {
	vtbl := *(*unsafe.Pointer)(this)
	fn := *(*uintptr)(unsafe.Pointer(uintptr(vtbl) + uintptr(index)*unsafe.Sizeof(uintptr(0)))) //#nosec G115 -- index is always one of this file's small hardcoded vtable-slot constants, never external input
	all := append([]uintptr{uintptr(this)}, args...)
	r, _, _ := syscall.SyscallN(fn, all...)
	return hrErr(fmt.Sprintf("vtbl[%d]", index), r)
}

func release(this unsafe.Pointer) {
	if this != nil {
		_ = vtblCall(this, 2)
	}
}

// newDesktopWallpaper initializes COM on the current (locked) OS thread and
// creates the shell's IDesktopWallpaper instance. It must run out-of-process
// (CLSCTX_LOCAL_SERVER) since the object lives in explorer.exe.
func newDesktopWallpaper() (unsafe.Pointer, error) {
	hr, _, _ := procCoInitializeEx.Call(0, coinitApartmentThreaded)
	// S_OK (0) and S_FALSE (1, already initialized) are both success.
	if hr != 0 && hr != 1 {
		return nil, hrErr("CoInitializeEx", hr)
	}

	var obj unsafe.Pointer
	hr, _, _ = procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidDesktopWallpaper)),
		0,
		clsctxLocalServer,
		uintptr(unsafe.Pointer(&iidIDesktopWallpaper)),
		uintptr(unsafe.Pointer(&obj)),
	)
	if err := hrErr("CoCreateInstance", hr); err != nil {
		coUninitialize()
		return nil, err
	}
	return obj, nil
}

func closeDesktopWallpaper(obj unsafe.Pointer) {
	release(obj)
	coUninitialize()
}

// parseDisplayName resolves a filesystem path to an absolute PIDL, as
// required by SHCreateShellItemArrayFromIDLists. The caller must free the
// result with CoTaskMemFree.
func parseDisplayName(path string) (unsafe.Pointer, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	var pidl unsafe.Pointer
	hr, _, _ := procSHParseDisplayName.Call(
		uintptr(unsafe.Pointer(p)),
		0,
		uintptr(unsafe.Pointer(&pidl)),
		0,
		0,
	)
	if err := hrErr("SHParseDisplayName", hr); err != nil {
		return nil, err
	}
	return pidl, nil
}

// setWallpaperFade transitions the desktop background from whatever is
// currently set to fullPath using the same crossfade the Settings app and
// slideshow feature use, by driving IDesktopWallpaper directly: set a
// two-item slideshow ([current, new]) and advance it once. If anything
// here fails (e.g. DWM/Aero unavailable), the caller falls back to the
// instant SystemParametersInfoW swap.
func setWallpaperFade(fullPath string) error {
	if err := advanceToWallpaper(fullPath); err != nil {
		return err
	}

	// Explorer only flips into DSS_SLIDESHOW mode a moment after
	// advanceToWallpaper's COM session has fully closed, and the delay is
	// nondeterministic (sub-second to a few seconds; with multiple
	// monitors the flips are also staggered per monitor). While in
	// slideshow mode it owns a recurring timer, so we must exit that mode
	// before returning. Pinning the desktop back to a static picture via
	// the legacy API does that reliably — but only if it lands *after* the
	// flip; a pin that lands before is silently undone by it. So: wait for
	// the flip (bounded), pin, and verify the pin stuck, re-pinning if the
	// flip raced past us.
	waitForSlideshowState(true, slideshowFlipTimeout)
	for attempt := 0; attempt < 3; attempt++ {
		if err := wallpaper.SetFromFile(fullPath); err != nil {
			return err
		}
		if !waitForSlideshowState(true, pinVerifyWindow) {
			return nil
		}
	}
	return fmt.Errorf("could not exit slideshow mode after fade transition")
}

// waitForSlideshowState polls IDesktopWallpaper::GetStatus until the
// slideshow bit matches want or the timeout elapses. It reports whether the
// desired state was observed. Status-read errors count as "not observed".
func waitForSlideshowState(want bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if active, err := slideshowActive(); err == nil && active == want {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(statusPollInterval)
	}
}

// slideshowActive reports whether Explorer currently considers the desktop
// background to be in slideshow mode.
func slideshowActive() (bool, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	dw, err := newDesktopWallpaper()
	if err != nil {
		return false, err
	}
	defer closeDesktopWallpaper(dw)

	var status uint32
	if err := vtblCall(dw, idwGetStatus, uintptr(unsafe.Pointer(&status))); err != nil {
		return false, err
	}
	return status&dssSlideshow != 0, nil
}

// advanceToWallpaper drives the actual IDesktopWallpaper crossfade and
// closes its COM session before returning. It configures a slideshow whose
// list contains only the new image: with a single item, Explorer is forced
// to show the same image on every monitor (a multi-item list gets spread
// across monitors, leaving them out of sync), and the advance still gets
// the DWM crossfade.
func advanceToWallpaper(fullPath string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	dw, err := newDesktopWallpaper()
	if err != nil {
		return err
	}
	defer closeDesktopWallpaper(dw)

	pidlNew, err := parseDisplayName(fullPath)
	if err != nil {
		return err
	}
	defer func() { coTaskMemFree(uintptr(pidlNew)) }()

	var itemArray unsafe.Pointer
	hr, _, _ := procSHCreateShellItemArrayFromIDLists.Call(
		1,
		uintptr(unsafe.Pointer(&pidlNew)),
		uintptr(unsafe.Pointer(&itemArray)),
	)
	if err := hrErr("SHCreateShellItemArrayFromIDLists", hr); err != nil {
		return err
	}
	defer release(itemArray)

	if err := vtblCall(dw, idwSetSlideshow, uintptr(itemArray)); err != nil {
		return err
	}
	if err := vtblCall(dw, idwSetSlideshowOptions, 0, 1000); err != nil {
		return err
	}
	// AdvanceSlideshow intermittently returns E_UNEXPECTED right after
	// SetSlideshow while Explorer is still ingesting the new image list;
	// a short retry gets past it.
	var advErr error
	for attempt := 0; attempt < 4; attempt++ {
		if advErr = vtblCall(dw, idwAdvanceSlideshow, 0, dsdForward); advErr == nil {
			return nil
		}
		if !isHResult(advErr, hrEUnexpected) {
			return advErr
		}
		time.Sleep(250 * time.Millisecond)
	}
	return advErr
}

// utf16PtrToStringAndFree reads a NUL-terminated CoTaskMem-allocated wide
// string returned by a COM out-parameter, frees it, and returns its Go form.
func utf16PtrToStringAndFree(p *uint16) string {
	if p == nil {
		return ""
	}
	defer coTaskMemFree(uintptr(unsafe.Pointer(p)))
	return windows.UTF16PtrToString(p)
}

// monitorDevicePaths returns the device path of every monitor Explorer
// knows about, in IDesktopWallpaper enumeration order (index 0 is
// "monitor 1" in the configuration).
func monitorDevicePaths() ([]string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	dw, err := newDesktopWallpaper()
	if err != nil {
		return nil, err
	}
	defer closeDesktopWallpaper(dw)

	var count uint32
	if err := vtblCall(dw, idwGetMonitorDevicePathCount, uintptr(unsafe.Pointer(&count))); err != nil {
		return nil, err
	}

	paths := make([]string, 0, count)
	for i := uint32(0); i < count; i++ {
		var id *uint16
		if err := vtblCall(dw, idwGetMonitorDevicePathAt, uintptr(i), uintptr(unsafe.Pointer(&id))); err != nil {
			return nil, err
		}
		paths = append(paths, utf16PtrToStringAndFree(id))
	}
	return paths, nil
}

// setWallpaperOnMonitor applies fullPath to the monitor identified by
// devicePath via IDesktopWallpaper::SetWallpaper. This is always an instant
// swap: the crossfade trick in setWallpaperFade relies on a one-item
// slideshow, which forces the same image onto every monitor and therefore
// cannot target monitors individually.
func setWallpaperOnMonitor(devicePath, fullPath string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	dw, err := newDesktopWallpaper()
	if err != nil {
		return err
	}
	defer closeDesktopWallpaper(dw)

	monitor, err := syscall.UTF16PtrFromString(devicePath)
	if err != nil {
		return err
	}
	path, err := syscall.UTF16PtrFromString(fullPath)
	if err != nil {
		return err
	}
	return vtblCall(dw, idwSetWallpaper, uintptr(unsafe.Pointer(monitor)), uintptr(unsafe.Pointer(path)))
}

func desktopWallpaperPosition(mode string) (uintptr, bool) {
	switch mode {
	case "center":
		return dwposCenter, true
	case "tile":
		return dwposTile, true
	case "stretch":
		return dwposStretch, true
	case "fit":
		return dwposFit, true
	case "span":
		return dwposSpan, true
	case "crop":
		return dwposFill, true
	default:
		return 0, false
	}
}

// setWallpaperPosition applies the display mode via IDesktopWallpaper
// instead of the legacy registry + reapply dance, so it doesn't trigger a
// second, abrupt wallpaper swap right after setWallpaperFade's transition.
func setWallpaperPosition(mode string) error {
	pos, ok := desktopWallpaperPosition(mode)
	if !ok {
		pos = dwposFill
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	dw, err := newDesktopWallpaper()
	if err != nil {
		return err
	}
	defer closeDesktopWallpaper(dw)

	return vtblCall(dw, idwSetPosition, pos)
}
