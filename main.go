package main

import "C"

import (
	"fmt"
	"github.com/AllenDang/w32"
	"golang.org/x/sys/windows"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

const (
	MAX_TOUCH_COUNT         = 10
	TOUCH_FEEDBACK_DEFAULT  = 0x1
	TOUCH_FEEDBACK_INDIRECT = 0x2
	TOUCH_FEEDBACK_NONE     = 0x3
)

const (
	TouchPointOffsetRelativeToMouse = 100
	MoveDistancePerUpdate           = 5
	UpdateCount                     = 100
)

type MouseHookStruct struct {
	pt           POINT
	hwnd         w32.HWND
	wHitTestCode uint
	dwExtraInfo  uintptr
}

type KeyboardHookStruct struct {
	vkCode      uint32
	scanCode    uint32
	flags       uint32
	time        uint32
	dwExtraInfo uintptr
}

var (
	user32DLL                = windows.NewLazyDLL("user32.dll")
	initializeTouchInjection = user32DLL.NewProc("InitializeTouchInjection")
	injectTouchInput         = user32DLL.NewProc("InjectTouchInput")
	getCursorPos             = user32DLL.NewProc("GetCursorPos")
)

var firstEvent = true

func LowLevelMouseProc(nCode int, wParam w32.WPARAM, lParam w32.LPARAM) w32.LRESULT {
	mhs := (*MouseHookStruct)(unsafe.Pointer(lParam))
	_ = mhs
	if nCode == 0 && (wParam == w32.WM_MOUSEWHEEL) {
		if firstEvent {
			var offset int32
			if mhs.hwnd == 0x00780000 {
				offset = TouchPointOffsetRelativeToMouse
			} else if mhs.hwnd == 0xFF880000 {
				offset = TouchPointOffsetRelativeToMouse * 3
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
			//fmt.Printf("Zoom in\n")
			touches[0].UpdateStart()
			touches[1].UpdateStart()
			touches[0].PointerInfo.PixelLocation.X += MoveDistancePerUpdate
			touches[0].PointerInfo.PixelLocation.Y -= MoveDistancePerUpdate

			touches[1].PointerInfo.PixelLocation.X -= MoveDistancePerUpdate
			touches[1].PointerInfo.PixelLocation.Y += MoveDistancePerUpdate

			injectTouchInput.Call(uintptr(touchLength), uintptr(unsafe.Pointer(&touches[0])))

			// Zoom in
		} else if mhs.hwnd == 0xFF880000 {
			//fmt.Printf("Zoom out\n")

			touches[0].UpdateStart()
			touches[1].UpdateStart()
			touches[0].PointerInfo.PixelLocation.X -= MoveDistancePerUpdate
			touches[0].PointerInfo.PixelLocation.Y += MoveDistancePerUpdate

			touches[1].PointerInfo.PixelLocation.X += MoveDistancePerUpdate
			touches[1].PointerInfo.PixelLocation.Y -= MoveDistancePerUpdate

			injectTouchInput.Call(uintptr(touchLength), uintptr(unsafe.Pointer(&touches[0])))

			// Zoom out
		}
		return 1
	}

	return w32.CallNextHookEx(keyboardHook, nCode, wParam, lParam)
}

func KeyboardProc(nCode int, wParam w32.WPARAM, lParam w32.LPARAM) w32.LRESULT {
	khs := (*KeyboardHookStruct)(unsafe.Pointer(lParam))

	// we only care about F21
	if khs.vkCode != w32.VK_F21 {
		return w32.CallNextHookEx(keyboardHook, nCode, wParam, lParam)
	}

	if nCode == 0 && (wParam == w32.WM_KEYDOWN || wParam == w32.WM_SYSKEYDOWN) {
		fmt.Printf("[%12d]F21 Pressed Scan : %X Flags : %b\n", time.Now().Unix(), khs.scanCode, khs.flags)

		//Hook for mouse events so we can check if the mouse is moving
		mouseHook = w32.SetWindowsHookEx(w32.WH_MOUSE_LL, LowLevelMouseProc, moduleInstance, 0)
		if mouseHook == 0 {
			fmt.Printf("SetWindowsHookEx for mouse failed: %d\n", w32.GetLastError())
		}

	} else if nCode == 0 && (wParam == w32.WM_KEYUP || wParam == w32.WM_SYSKEYUP) {
		fmt.Printf("[%12d]F21 Released Scan : %X Flags : %b\n", time.Now().Unix(), khs.scanCode, khs.flags)

		touches[0].Release()
		touches[1].Release()
		injectTouchInput.Call(uintptr(touchLength), uintptr(unsafe.Pointer(&touches[0])))
		firstEvent = true

		// Unhook the mouse hook
		w32.UnhookWindowsHookEx(mouseHook)
	}

	return w32.CallNextHookEx(keyboardHook, nCode, wParam, lParam)
}

var keyboardHook w32.HHOOK
var mouseHook w32.HHOOK
var moduleInstance w32.HINSTANCE
var p POINT

var touches = make([]PointerTouchInfo, 2)
var touchLength = 2

func main() {
	runtime.LockOSThread()
	runtime.GOMAXPROCS(1)
	// define the signals we want to handle
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	stop := make(chan struct{})

	// start a goroutine to handle signals
	go func() {
		sig := <-signals
		fmt.Printf("Housekeeping : %d\n", sig)
		w32.UnhookWindowsHookEx(keyboardHook)

		stop <- struct{}{}

		os.Exit(0)
	}()

	moduleInstance = w32.GetModuleHandle("")

	keyboardHook = w32.SetWindowsHookEx(w32.WH_KEYBOARD_LL, KeyboardProc, moduleInstance, 0)
	if keyboardHook == 0 {
		errCode := w32.GetLastError()
		fmt.Printf("SetWindowsHookEx failed: %d\n", errCode)
		// You might want to print a more detailed error message based on the error code.
		//fmt.Printf("Error description: %s\n", w32.FormatMessage(errCode))
		return
	}

	fmt.Printf("keyboardHook : %d\n", keyboardHook)

	result, _, _ := initializeTouchInjection.Call(MAX_TOUCH_COUNT, TOUCH_FEEDBACK_DEFAULT)
	r := (*bool)(unsafe.Pointer(&result))
	fmt.Printf("initializeTouchInjection : %t\n", *r)

	getCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	fmt.Printf("CursorPos : %+v\n", p)

	go func() {
		var i = 0
		for {
			select {
			case <-time.After(1 * time.Second):
				fmt.Printf("[%012d] %d : %d %d\n", time.Now().Unix(), i, keyboardHook, mouseHook)
				i++
			}
		}
	}()

	go func() {
		for {
			// Process pending messages without blocking
			w32.GetMessage(nil, 0, w32.WM_QUIT, w32.WM_QUIT+1)

			// Check if stop signal received
			select {
			case <-stop:
				return // Stop the goroutine
			default:
			}
		}
	}()

	for w32.GetMessage(nil, 0, w32.WM_QUIT, w32.WM_QUIT+1) != 0 {
		// NOP while not WM_QUIT
		fmt.Printf("NOP\n")
	}
	fmt.Printf("Housekeeping\n")
	w32.UnhookWindowsHookEx(keyboardHook)
}

func ZoomIn(p POINT) {
	contacts := make([]PointerTouchInfo, 2)
	length := len(contacts)

	contacts[0].Init(1, p.X+TouchPointOffsetRelativeToMouse, p.Y-TouchPointOffsetRelativeToMouse)
	contacts[1].Init(2, p.X-TouchPointOffsetRelativeToMouse, p.Y+TouchPointOffsetRelativeToMouse)

	contacts[0].Press()
	contacts[1].Press()

	injectTouchInput.Call(uintptr(length), uintptr(unsafe.Pointer(&contacts[0])))

	contacts[0].UpdateStart()
	contacts[1].UpdateStart()
	for i := 0; i < UpdateCount; i += 4 {
		contacts[0].PointerInfo.PixelLocation.X += MoveDistancePerUpdate
		contacts[0].PointerInfo.PixelLocation.Y -= MoveDistancePerUpdate

		contacts[1].PointerInfo.PixelLocation.X -= MoveDistancePerUpdate
		contacts[1].PointerInfo.PixelLocation.Y += MoveDistancePerUpdate

		injectTouchInput.Call(uintptr(length), uintptr(unsafe.Pointer(&contacts[0])))
		<-time.After(1 * time.Microsecond)
	}

	contacts[0].Release()
	contacts[1].Release()
	//contacts.PointerInfo.PointerFlags = PointerFlagUp
	//fmt.Printf("%+v\n", contacts[0])

	injectTouchInput.Call(uintptr(length), uintptr(unsafe.Pointer(&contacts[0])))
}

func ZoomOut(p POINT) {
	contacts := make([]PointerTouchInfo, 2)
	length := len(contacts)

	contacts[0].Init(1, p.X+MoveDistancePerUpdate*UpdateCount, p.Y-MoveDistancePerUpdate*UpdateCount)
	contacts[1].Init(2, p.X-MoveDistancePerUpdate*UpdateCount, p.Y+MoveDistancePerUpdate*UpdateCount)

	contacts[0].Press()
	contacts[1].Press()

	injectTouchInput.Call(uintptr(length), uintptr(unsafe.Pointer(&contacts[0])))

	contacts[0].UpdateStart()
	contacts[1].UpdateStart()
	for i := 0; i < UpdateCount; i += 3 {
		contacts[0].PointerInfo.PixelLocation.X -= MoveDistancePerUpdate
		contacts[0].PointerInfo.PixelLocation.Y += MoveDistancePerUpdate

		contacts[1].PointerInfo.PixelLocation.X += MoveDistancePerUpdate
		contacts[1].PointerInfo.PixelLocation.Y -= MoveDistancePerUpdate

		injectTouchInput.Call(uintptr(length), uintptr(unsafe.Pointer(&contacts[0])))

		<-time.After(1 * time.Microsecond)
	}

	contacts[0].Release()
	contacts[1].Release()
	//contacts.PointerInfo.PointerFlags = PointerFlagUp
	//fmt.Printf("%+v\n", contacts[0])

	injectTouchInput.Call(uintptr(length), uintptr(unsafe.Pointer(&contacts[0])))
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
