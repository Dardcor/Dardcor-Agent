package services

import (
	"strings"
	"syscall"
	"unsafe"
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	procSendInput       = user32.NewProc("SendInput")
	procGetMetrics      = user32.NewProc("GetSystemMetrics")
	procGetCursorPos    = user32.NewProc("GetCursorPos")
	procSetCursorPos    = user32.NewProc("SetCursorPos")
	procWindowFromPt    = user32.NewProc("WindowFromPoint")
	procEnumWindows     = user32.NewProc("EnumWindows")
	procGetWindowText   = user32.NewProc("GetWindowTextW")
	procIsWindowVisible = user32.NewProc("IsWindowVisible")
)

const (
	INPUT_MOUSE    = 0
	INPUT_KEYBOARD = 1

	MOUSEEVENTF_MOVE      = 0x0001
	MOUSEEVENTF_LEFTDOWN  = 0x0002
	MOUSEEVENTF_LEFTUP    = 0x0004
	MOUSEEVENTF_RIGHTDOWN = 0x0008
	MOUSEEVENTF_RIGHTUP   = 0x0010
	MOUSEEVENTF_ABSOLUTE  = 0x8000

	KEYEVENTF_KEYUP = 0x0002

	SM_CXSCREEN = 0
	SM_CYSCREEN = 1
)

type MOUSEINPUT struct {
	DX        int32
	DY        int32
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type KEYBDINPUT struct {
	Vk        uint16
	Scan      uint16
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type INPUT struct {
	Type uint32
	Data [24]byte // Large enough for MOUSEINPUT or KEYBDINPUT
}

type POINT struct {
	X int32
	Y int32
}

type AutomationService struct{}

func NewAutomationService() *AutomationService {
	return &AutomationService{}
}

func (s *AutomationService) GetScreenSize() (int, int) {
	width, _, _ := procGetMetrics.Call(uintptr(SM_CXSCREEN))
	height, _, _ := procGetMetrics.Call(uintptr(SM_CYSCREEN))
	return int(width), int(height)
}

func (s *AutomationService) GetCursorPos() (int, int) {
	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return int(pt.X), int(pt.Y)
}

func (s *AutomationService) MouseMove(x, y int) error {
	w, h := s.GetScreenSize()
	// Absolute coordinates are 0-65535
	ax := int32(x * 65535 / w)
	ay := int32(y * 65535 / h)

	var input INPUT
	input.Type = INPUT_MOUSE
	mi := (*MOUSEINPUT)(unsafe.Pointer(&input.Data))
	mi.DX = ax
	mi.DY = ay
	mi.Flags = MOUSEEVENTF_MOVE | MOUSEEVENTF_ABSOLUTE

	ret, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), uintptr(unsafe.Sizeof(input)))
	if ret == 0 {
		return err
	}
	return nil
}

func (s *AutomationService) MouseClick(x, y int, button string) error {
	if err := s.MouseMove(x, y); err != nil {
		return err
	}

	var down, up uint32
	if strings.ToLower(button) == "right" {
		down = MOUSEEVENTF_RIGHTDOWN
		up = MOUSEEVENTF_RIGHTUP
	} else {
		down = MOUSEEVENTF_LEFTDOWN
		up = MOUSEEVENTF_LEFTUP
	}

	var inputs [2]INPUT
	inputs[0].Type = INPUT_MOUSE
	miD := (*MOUSEINPUT)(unsafe.Pointer(&inputs[0].Data))
	miD.Flags = down

	inputs[1].Type = INPUT_MOUSE
	miU := (*MOUSEINPUT)(unsafe.Pointer(&inputs[1].Data))
	miU.Flags = up

	ret, _, err := procSendInput.Call(2, uintptr(unsafe.Pointer(&inputs)), uintptr(unsafe.Sizeof(inputs[0])))
	if ret == 0 {
		return err
	}
	return nil
}

func (s *AutomationService) Type(text string) error {
	for _, char := range text {
		// This is a simplified version using Vk
		// For a real supreme agent, we should use ScanCode for Unicode
		vk := s.getVkFromChar(char)
		if vk == 0 {
			continue
		}

		var inputs [2]INPUT
		inputs[0].Type = INPUT_KEYBOARD
		kiD := (*KEYBDINPUT)(unsafe.Pointer(&inputs[0].Data))
		kiD.Vk = vk

		inputs[1].Type = INPUT_KEYBOARD
		kiU := (*KEYBDINPUT)(unsafe.Pointer(&inputs[1].Data))
		kiU.Vk = vk
		kiU.Flags = KEYEVENTF_KEYUP

		procSendInput.Call(2, uintptr(unsafe.Pointer(&inputs)), uintptr(unsafe.Sizeof(inputs[0])))
	}
	return nil
}

func (s *AutomationService) getVkFromChar(r rune) uint16 {
	// Simplified mapping
	if r >= 'a' && r <= 'z' {
		return uint16(r - 'a' + 0x41)
	}
	if r >= 'A' && r <= 'Z' {
		return uint16(r - 'A' + 0x41)
	}
	if r >= '0' && r <= '9' {
		return uint16(r - '0' + 0x30)
	}
	if r == ' ' {
		return 0x20
	}
	if r == '\n' || r == '\r' {
		return 0x0D
	}
	if r == '.' {
		return 0xBE
	}
	return 0
}

func (s *AutomationService) PressKey(vk uint16) error {
	var inputs [2]INPUT
	inputs[0].Type = INPUT_KEYBOARD
	kiD := (*KEYBDINPUT)(unsafe.Pointer(&inputs[0].Data))
	kiD.Vk = vk

	inputs[1].Type = INPUT_KEYBOARD
	kiU := (*KEYBDINPUT)(unsafe.Pointer(&inputs[1].Data))
	kiU.Vk = vk
	kiU.Flags = KEYEVENTF_KEYUP

	ret, _, err := procSendInput.Call(2, uintptr(unsafe.Pointer(&inputs)), uintptr(unsafe.Sizeof(inputs[0])))
	if ret == 0 {
		return err
	}
	return nil
}

func (s *AutomationService) ListWindows() []string {
	var windows []string
	cb := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		visible, _, _ := procIsWindowVisible.Call(hwnd)
		if visible != 0 {
			buf := make([]uint16, 255)
			procGetWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
			name := syscall.UTF16ToString(buf)
			if name != "" {
				windows = append(windows, name)
			}
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return windows
}
