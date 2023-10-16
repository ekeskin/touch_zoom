package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"unsafe"

	"fyne.io/systray"
	"fyne.io/systray/example/icon"
	"github.com/AllenDang/w32"
	"golang.org/x/sys/windows"
)

// Configuration Variables
var (
	TouchPointOffsetRelativeToMouse   int32  = 100
	TouchPointOffsetZoomOutMultiplier int32  = 3
	ZoomDistancePerUpdate             int32  = 3
	MoveHandlingEnabled               bool   = false
	MoveHandlingNatural               bool   = true
	TriggerKey                        uint32 = w32.VK_F21
)

var (
	user32DLL                = windows.NewLazyDLL("user32.dll")
	initializeTouchInjection = user32DLL.NewProc("InitializeTouchInjection")
	injectTouchInput         = user32DLL.NewProc("InjectTouchInput")
)

var firstEvent = true
var keyboardHook w32.HHOOK
var mouseHook w32.HHOOK
var moduleInstance w32.HINSTANCE

var touches = make([]PointerTouchInfo, 2)
var touchLength = 2
var signals = make(chan os.Signal, 1)

var lastMovePosition w32.POINT

func main() {
	runtime.LockOSThread()
	runtime.GOMAXPROCS(1)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	go signalHandler()

	moduleInstance = w32.GetModuleHandle("")
	keyboardHook = w32.SetWindowsHookEx(w32.WH_KEYBOARD_LL, keyboardHookCallback, moduleInstance, 0)
	if keyboardHook == 0 {
		errCode := w32.GetLastError()
		log.Fatalf("SetWindowsHookEx failed: %d\n", errCode)
		return
	}

	result, _, _ := initializeTouchInjection.Call(MAX_TOUCH_COUNT, TOUCH_FEEDBACK_DEFAULT)
	if *(*bool)(unsafe.Pointer(&result)) == false {
		log.Fatalf("initializeTouchInjection failed\n")
		return
	}

	systray.Run(onReady, onExit)
}

func mouseHookCallback(nCode int, wParam w32.WPARAM, lParam w32.LPARAM) w32.LRESULT {
	if nCode != 0 {
		return w32.CallNextHookEx(keyboardHook, nCode, wParam, lParam)
	}
	mhs := (*MouseHookStruct)(unsafe.Pointer(lParam))
	switch wParam {
	case w32.WM_MOUSEWHEEL:
		if firstEvent {
			var offset int32
			if mhs.hwnd == 0x00780000 {
				offset = TouchPointOffsetRelativeToMouse
			} else if mhs.hwnd == 0xFF880000 {
				offset = TouchPointOffsetRelativeToMouse * TouchPointOffsetZoomOutMultiplier
			}
			firstEvent = false
			touches[0].Init(0, mhs.pt.X+offset, mhs.pt.Y-offset)
			touches[1].Init(1, mhs.pt.X-offset, mhs.pt.Y+offset)

			touches[0].Press()
			touches[1].Press()
			injectTouchInput.Call(uintptr(touchLength), uintptr(unsafe.Pointer(&touches[0])))
			return 1
		}
		if mhs.hwnd == 0x00780000 {
			touches[0].UpdateStart()
			touches[0].MoveRel(ZoomDistancePerUpdate, ZoomDistancePerUpdate)

			touches[1].UpdateStart()
			touches[1].MoveRel(-ZoomDistancePerUpdate, -ZoomDistancePerUpdate)

			injectTouchInput.Call(uintptr(touchLength), uintptr(unsafe.Pointer(&touches[0])))

		} else if mhs.hwnd == 0xFF880000 {

			touches[0].UpdateStart()
			touches[0].MoveRel(-ZoomDistancePerUpdate, -ZoomDistancePerUpdate)

			touches[1].UpdateStart()
			touches[1].MoveRel(ZoomDistancePerUpdate, ZoomDistancePerUpdate)

			injectTouchInput.Call(uintptr(touchLength), uintptr(unsafe.Pointer(&touches[0])))
		}
		return 1
	case w32.WM_MOUSEMOVE:
		if !MoveHandlingEnabled {
			return w32.CallNextHookEx(keyboardHook, nCode, wParam, lParam)
		}
		if firstEvent {
			firstEvent = false
			touches[0].Init(0, mhs.pt.X, mhs.pt.Y)
			touches[0].Press()
			injectTouchInput.Call(1, uintptr(unsafe.Pointer(&touches[0])))

			lastMovePosition = mhs.pt
			return 1
		}

		touches[0].UpdateStart()
		if MoveHandlingNatural {
			diff := w32.POINT{X: mhs.pt.X - lastMovePosition.X, Y: mhs.pt.Y - lastMovePosition.Y}
			touches[0].MoveRel(-diff.X, -diff.Y)
		} else {
			touches[0].MoveAbs(mhs.pt.X, mhs.pt.Y)
		}

		injectTouchInput.Call(1, uintptr(unsafe.Pointer(&touches[0])))

		lastMovePosition = mhs.pt
		return 0
		break
	default:
		return w32.CallNextHookEx(keyboardHook, nCode, wParam, lParam)
	}

	return w32.CallNextHookEx(keyboardHook, nCode, wParam, lParam)
}

func keyboardHookCallback(nCode int, wParam w32.WPARAM, lParam w32.LPARAM) w32.LRESULT {
	if nCode < 0 {
		return w32.CallNextHookEx(keyboardHook, nCode, wParam, lParam)
	}

	khs := (*KeyboardHookStruct)(unsafe.Pointer(lParam))

	if khs.vkCode != TriggerKey {
		return w32.CallNextHookEx(keyboardHook, nCode, wParam, lParam)
	}

	if nCode == 0 && (wParam == w32.WM_KEYDOWN || wParam == w32.WM_SYSKEYDOWN) {
		log.Printf("F21 Pressed\n")
		mouseHook = w32.SetWindowsHookEx(w32.WH_MOUSE_LL, mouseHookCallback, moduleInstance, 0)
		if mouseHook == 0 {
			log.Printf("SetWindowsHookEx for mouse failed: %d\n", w32.GetLastError())
		}

	} else if nCode == 0 && (wParam == w32.WM_KEYUP || wParam == w32.WM_SYSKEYUP) {
		log.Printf("F21 Released\n")

		touches[0].Release()
		touches[1].Release()
		injectTouchInput.Call(uintptr(touchLength), uintptr(unsafe.Pointer(&touches[0])))
		firstEvent = true

		success := w32.UnhookWindowsHookEx(mouseHook)
		if !success {
			log.Printf("UnhookWindowsHookEx failed: %d\n", w32.GetLastError())
		}
	}

	return w32.CallNextHookEx(keyboardHook, nCode, wParam, lParam)
}

func signalHandler() {
	sig := <-signals
	log.Printf("Received signal %s\n", sig)
	housekeeping()
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("Touch Zoom")
	systray.SetTooltip("Touch Zoom")

	mMoveHandlingToggle := systray.AddMenuItem("✘ Move Handling Disabled", "enable or disable mouse movement handling")
	mMoveHandlingNatural := systray.AddMenuItem("Natural Move Handling", "When enabled, mouse movements will act as if you are touching the screen directly")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		for {
			select {
			case <-mMoveHandlingToggle.ClickedCh:
				MoveHandlingEnabled = !MoveHandlingEnabled
				if MoveHandlingEnabled {
					mMoveHandlingToggle.SetTitle("✓ Move Handling Enabled")
					mMoveHandlingNatural.Enable()
				} else {
					mMoveHandlingToggle.SetTitle("✘ Move Handling Disabled")
					mMoveHandlingNatural.Disabled()
				}
			case <-mMoveHandlingNatural.ClickedCh:
				MoveHandlingNatural = !MoveHandlingNatural
				if MoveHandlingNatural {
					mMoveHandlingNatural.SetTitle("Natural Move Handling")
				} else {
					mMoveHandlingNatural.SetTitle("Inverted Move Handling")
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	housekeeping()
}

func housekeeping() {
	success := w32.UnhookWindowsHookEx(keyboardHook)
	if !success {
		log.Printf("UnhookWindowsHookEx failed: %d\n", w32.GetLastError())
	}

	os.Exit(0)
}

func (p *PointerTouchInfo) Init(id uint32, x, y int32) {
	p.PointerInfo.PointerType = PointerInputTouch
	p.PointerInfo.PointerId = id
	p.PointerInfo.PixelLocation.X = x
	p.PointerInfo.PixelLocation.Y = y

	p.TouchFlags = TouchFlagNone
	p.TouchMask = TouchMaskContactArea | TouchMaskOrientation | TouchMaskPressure
	p.Orientation = 90
	p.Pressure = 32000

	p.Contact.Left = p.PointerInfo.PixelLocation.Y - 2
	p.Contact.Top = p.PointerInfo.PixelLocation.Y + 2
	p.Contact.Right = p.PointerInfo.PixelLocation.X - 2
	p.Contact.Bottom = p.PointerInfo.PixelLocation.X + 2
}

func (p *PointerTouchInfo) Press() {
	p.PointerInfo.PointerFlags = PointerFlagDown | PointerFlagInContact | PointerFlagInRange
}

func (p *PointerTouchInfo) Release() {
	p.PointerInfo.PointerFlags = PointerFlagUp
}

func (p *PointerTouchInfo) UpdateStart() {
	p.PointerInfo.PointerFlags = PointerFlagUpdate | PointerFlagInContact | PointerFlagInRange
}

func (p *PointerTouchInfo) MoveRel(x, y int32) {
	p.PointerInfo.PixelLocation.X += x
	p.PointerInfo.PixelLocation.Y -= y
}

func (p *PointerTouchInfo) MoveAbs(x, y int32) {
	p.PointerInfo.PixelLocation.X = x
	p.PointerInfo.PixelLocation.Y = y
}
