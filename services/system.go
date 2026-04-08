package services

import (
	"fmt"
	"runtime"
	"sort"
	"time"
	"os/exec"
	"strings"

	"dardcor-agent/models"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type SystemService struct{}

func NewSystemService() *SystemService {
	return &SystemService{}
}

func (ss *SystemService) GetSystemInfo() (*models.SystemInfo, error) {
	info := &models.SystemInfo{
		CollectedAt: time.Now(),
		Bios:        make(map[string]string),
	}

	hostInfo, err := host.Info()
	if err == nil {
		info.OS = models.OSInfo{
			Platform: hostInfo.Platform,
			Family:   hostInfo.PlatformFamily,
			Version:  hostInfo.PlatformVersion,
			Arch:     runtime.GOARCH,
		}
		info.Uptime = hostInfo.Uptime
		info.HostName = hostInfo.Hostname
		
		info.Bios["kernel_version"] = hostInfo.KernelVersion
		info.Bios["os_release"] = hostInfo.OS
	}


	if runtime.GOOS == "windows" {
		out, err := exec.Command("wmic", "path", "win32_VideoController", "get", "name").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 1 {
				gpu := strings.TrimSpace(lines[1])
				if gpu != "" {
					info.Bios["gpu"] = gpu
				}
			}
		}
		
		out, err = exec.Command("wmic", "baseboard", "get", "product,manufacturer").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 1 {
				info.Bios["motherboard"] = strings.TrimSpace(lines[1])
			}
		}
	}

	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		info.CPU = models.CPUInfo{
			ModelName: cpuInfo[0].ModelName,
			Cores:     int(cpuInfo[0].Cores),
			Frequency: cpuInfo[0].Mhz,
		}
		info.CPU.Threads = runtime.NumCPU()
	}

	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		info.CPU.UsagePercent = cpuPercent[0]
	}

	memInfo, err := mem.VirtualMemory()
	if err == nil {
		info.Memory = models.MemoryInfo{
			Total:       memInfo.Total,
			Used:        memInfo.Used,
			Free:        memInfo.Free,
			UsedPercent: memInfo.UsedPercent,
		}
	}

	partitions, err := disk.Partitions(false)
	if err == nil {
		for _, p := range partitions {
			usage, err := disk.Usage(p.Mountpoint)
			if err != nil {
				continue
			}
			info.Disks = append(info.Disks, models.DiskInfo{
				Device:      p.Device,
				MountPoint:  p.Mountpoint,
				FSType:      p.Fstype,
				Total:       usage.Total,
				Used:        usage.Used,
				Free:        usage.Free,
				UsedPercent: usage.UsedPercent,
			})
		}
	}

	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			var addrs []string
			for _, addr := range iface.Addrs {
				addrs = append(addrs, addr.Addr)
			}
			info.Network = append(info.Network, models.NetInfo{
				Name:       iface.Name,
				MacAddress: iface.HardwareAddr,
				Addresses:  addrs,
			})
		}
	}

	netIO, err := net.IOCounters(true)
	if err == nil {
		for _, io := range netIO {
			for j := range info.Network {
				if info.Network[j].Name == io.Name {
					info.Network[j].BytesSent = io.BytesSent
					info.Network[j].BytesRecv = io.BytesRecv
					break
				}
			}
		}
	}

	return info, nil
}

func (ss *SystemService) GetProcesses(sortBy string, limit int) ([]models.ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var processList []models.ProcessInfo
	for _, p := range procs {
		name, _ := p.Name()
		status, _ := p.Status()
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()
		memInfo, _ := p.MemoryInfo()
		username, _ := p.Username()
		createTime, _ := p.CreateTime()
		cmdline, _ := p.Cmdline()

		memMB := float64(0)
		if memInfo != nil {
			memMB = float64(memInfo.RSS) / 1024 / 1024
		}

		statusStr := ""
		if len(status) > 0 {
			statusStr = status[0]
		}

		processList = append(processList, models.ProcessInfo{
			PID:        p.Pid,
			Name:       name,
			Status:     statusStr,
			CPUPercent: cpuPercent,
			MemPercent: memPercent,
			MemoryMB:   memMB,
			Username:   username,
			CreateTime: createTime,
			CmdLine:    cmdline,
		})
	}

	switch sortBy {
	case "cpu":
		sort.Slice(processList, func(i, j int) bool {
			return processList[i].CPUPercent > processList[j].CPUPercent
		})
	case "memory", "mem":
		sort.Slice(processList, func(i, j int) bool {
			return processList[i].MemoryMB > processList[j].MemoryMB
		})
	case "name":
		sort.Slice(processList, func(i, j int) bool {
			return processList[i].Name < processList[j].Name
		})
	default:
		sort.Slice(processList, func(i, j int) bool {
			return processList[i].PID < processList[j].PID
		})
	}

	if limit > 0 && limit < len(processList) {
		processList = processList[:limit]
	}

	return processList, nil
}

func (ss *SystemService) KillProcess(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}
	return p.Kill()
}

func (ss *SystemService) GetCPUUsageRealtime() ([]float64, error) {
	return cpu.Percent(time.Second, true)
}

func (ss *SystemService) GetMemoryUsage() (*models.MemoryInfo, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return &models.MemoryInfo{
		Total:       memInfo.Total,
		Used:        memInfo.Used,
		Free:        memInfo.Free,
		UsedPercent: memInfo.UsedPercent,
	}, nil
}
