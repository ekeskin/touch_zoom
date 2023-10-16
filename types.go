package main

import "github.com/AllenDang/w32"

type PointerTouchInfo struct {
	PointerInfo PointerInfo
	TouchFlags  TouchFlags
	TouchMask   TouchMask
	Contact     w32.RECT
	ContactRaw  w32.RECT
	Orientation uint32
	Pressure    uint32
}

type PointerInfo struct {
	PointerType         PointerInputType
	PointerId           uint32
	FrameId             uint32
	PointerFlags        PointerFlags
	SourceDevice        HANDLE
	WindowTarget        HWND
	PixelLocation       w32.POINT
	HimetricLocation    w32.POINT
	PixelLocationRaw    w32.POINT
	HimetricLocationRaw w32.POINT
	Time                uint32
	HistoryCount        uint32
	InputData           uint32
	KeyStates           uint32
	PerformanceCount    uint64
	ButtonChangeType    PointerButtonChangeType
}

type MouseHookStruct struct {
	pt           w32.POINT
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
