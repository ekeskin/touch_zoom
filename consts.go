package main

type HANDLE uintptr
type HWND uintptr

type PointerInputType uint32

const (
	PointerInputPointer  PointerInputType = 1
	PointerInputTouch                     = 2
	PointerInputPen                       = 3
	PointerInputMouse                     = 4
	PointerInputTouchpad                  = 5
)

type PointerFlags uint32 // Could be something else
const (
	PointerFlagNone           PointerFlags = 0x00000000
	PointerFlagNew                         = 0x00000001
	PointerFlagInRange                     = 0x00000002
	PointerFlagInContact                   = 0x00000004
	PointerFlagFirstButton                 = 0x00000010
	PointerFlagSecondButton                = 0x00000020
	PointerFlagThirdButton                 = 0x00000040
	PointerFlagFourthButton                = 0x00000080
	PointerFlagFifthButton                 = 0x00000100
	PointerFlagPrimary                     = 0x00002000
	PointerFlagConfidence                  = 0x00004000
	PointerFlagCanceled                    = 0x00008000
	PointerFlagDown                        = 0x00010000
	PointerFlagUpdate                      = 0x00020000
	PointerFlagUp                          = 0x00040000
	PointerFlagWheel                       = 0x00080000
	PointerFlagHWheel                      = 0x00100000
	PointerFlagCaptureChanged              = 0x00200000
	PointerFlagHasTransform                = 0x00400000
)

type PointerButtonChangeType uint32

const (
	PointerChangeNone PointerButtonChangeType = iota
	PointerChangeFirstButtonDown
	PointerChangeFirstButtonUp
	PointerChangeSecondButtonDown
	PointerChangeSecondButtonUp
	PointerChangeThirdButtonDown
	PointerChangeThirdButtonUp
	PointerChangeFourthButtonDown
	PointerChangeFourthButtonUp
	PointerChangeFifthButtonDown
	PointerChangeFifthButtonUp
)

type ModifierKeyState uint32

const (
	PointerModifierShift ModifierKeyState = 0x0004
	PointerModifierCtrl                   = 0x0008
)

type TouchFlags uint32

const (
	TouchFlagNone TouchFlags = 0x00000000
)

type TouchMask uint32

const (
	TouchMaskNone        TouchMask = 0x00000000
	TouchMaskContactArea           = 0x00000001
	TouchMaskOrientation           = 0x00000002
	TouchMaskPressure              = 0x00000004
)
