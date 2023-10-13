package main

type POINT struct {
	X, Y int32
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

type PointerTouchInfo struct {
	PointerInfo PointerInfo
	TouchFlags  TouchFlags
	TouchMask   TouchMask
	Contact     RECT
	ContactRaw  RECT
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
	PixelLocation       POINT
	HimetricLocation    POINT
	PixelLocationRaw    POINT
	HimetricLocationRaw POINT
	Time                uint32
	HistoryCount        uint32
	InputData           uint32
	KeyStates           uint32
	PerformanceCount    uint64
	ButtonChangeType    PointerButtonChangeType
}
