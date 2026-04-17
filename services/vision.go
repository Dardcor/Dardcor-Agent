package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type VisionService struct {
	storageDir string
}

func NewVisionService() *VisionService {
	dir := filepath.Join("storage", "vision")
	os.MkdirAll(dir, 0755)
	return &VisionService{storageDir: dir}
}

func (s *VisionService) CaptureScreen() (string, error) {
	filename := fmt.Sprintf("screen_%d.png", time.Now().Unix())
	outputPath := filepath.Join(s.storageDir, filename)
	absPath, _ := filepath.Abs(outputPath)

	// PowerShell script to take a screenshot
	psScript := fmt.Sprintf(`
		Add-Type -AssemblyName System.Windows.Forms
		Add-Type -AssemblyName System.Drawing
		$screen = [System.Windows.Forms.Screen]::PrimaryScreen
		$top    = $screen.Bounds.Top
		$left   = $screen.Bounds.Left
		$width  = $screen.Bounds.Width
		$height = $screen.Bounds.Height
		$bitmap = New-Object System.Drawing.Bitmap($width, $height)
		$graphics = [System.Drawing.Graphics]::FromImage($bitmap)
		$graphics.CopyFromScreen($left, $top, 0, 0, $bitmap.Size)
		$bitmap.Save('%s', [System.Drawing.Imaging.ImageFormat]::Png)
		$graphics.Dispose()
		$bitmap.Dispose()
	`, absPath)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("powershell failed: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("screenshot file not created")
	}

	return absPath, nil
}

func (s *VisionService) CleanupOldScreenshots() {
	files, err := os.ReadDir(s.storageDir)
	if err != nil {
		return
	}

	// Keep last 10 screenshots
	if len(files) <= 10 {
		return
	}

	for i := 0; i < len(files)-10; i++ {
		os.Remove(filepath.Join(s.storageDir, files[i].Name()))
	}
}
