package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/laragal/laragal/internal/winutil"
)

type ServiceStatus struct {
	Name    string `json:"name"`
	Running bool   `json:"running"`
	PID     int    `json:"pid"`
	Port    int    `json:"port"`
}

type managedProcess struct {
	cmd    *exec.Cmd
	status ServiceStatus
}

type Supervisor struct {
	mu        sync.RWMutex
	processes map[string]*managedProcess
	LogsDir   string
	OnExit    func(name string)
}

func NewSupervisor() *Supervisor {
	return &Supervisor{
		processes: make(map[string]*managedProcess),
		LogsDir:   filepath.Join("data", "logs"),
	}
}

func (s *Supervisor) getLogsDir() string {
	if s.LogsDir != "" {
		return s.LogsDir
	}
	return filepath.Join("data", "logs")
}

func (s *Supervisor) StartService(name string, exe string, args []string, port int) error {
	return s.StartServiceInDirWithEnv(name, exe, args, "", nil, port)
}

func (s *Supervisor) StartServiceWithEnv(name string, exe string, args []string, env []string, port int) error {
	return s.StartServiceInDirWithEnv(name, exe, args, "", env, port)
}

func (s *Supervisor) StartServiceInDir(name string, exe string, args []string, dir string, port int) error {
	return s.StartServiceInDirWithEnv(name, exe, args, dir, nil, port)
}

func (s *Supervisor) StartServiceInDirWithEnv(name string, exe string, args []string, dir string, env []string, port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p, exists := s.processes[name]; exists && p.status.Running {
		return fmt.Errorf("service %s is already running", name)
	}

	if filepath.Base(exe) == exe {
		if lp, err := exec.LookPath(exe); err == nil {
			exe = lp
		}
	} else if absExe, err := filepath.Abs(exe); err == nil {
		exe = absExe
	}
	if dir != "" {
		if absDir, err := filepath.Abs(dir); err == nil {
			dir = absDir
		}
	}

	logsDir := s.getLogsDir()
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}
	logPath := filepath.Join(logsDir, fmt.Sprintf("%s.log", name))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file for %s: %w", name, err)
	}

	cmd := winutil.Command(exe, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if dir != "" {
		cmd.Dir = dir
	}
	if len(env) > 0 {
		cmd.Env = append(cmd.Environ(), env...)
	}
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return fmt.Errorf("failed to start service %s: %w", name, err)
	}

	p := &managedProcess{
		cmd: cmd,
		status: ServiceStatus{
			Name:    name,
			Running: true,
			PID:     cmd.Process.Pid,
			Port:    port,
		},
	}
	s.processes[name] = p

	// Monitor termination in background goroutine
	go func() {
		_ = cmd.Wait()
		_ = logFile.Close()

		s.mu.Lock()
		isCurrent := false
		if current, ok := s.processes[name]; ok && current == p {
			current.status.Running = false
			current.status.PID = 0
			isCurrent = true
		}
		onExit := s.OnExit
		s.mu.Unlock()

		if isCurrent && onExit != nil {
			onExit(name)
		}
	}()

	return nil
}

func (s *Supervisor) StopService(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, exists := s.processes[name]
	if !exists {
		return nil
	}

	if p.cmd != nil && p.cmd.Process != nil && p.cmd.Process.Pid > 0 {
		_ = winutil.KillProcessTree(p.cmd.Process.Pid)
		_ = p.cmd.Process.Kill()
	}
	if p.status.Port > 0 {
		_ = winutil.KillProcessesOnPort(p.status.Port)
	}

	p.status.Running = false
	p.status.PID = 0
	return nil
}

func (s *Supervisor) StopAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.processes {
		if p.cmd != nil && p.cmd.Process != nil && p.cmd.Process.Pid > 0 {
			_ = winutil.KillProcessTree(p.cmd.Process.Pid)
			_ = p.cmd.Process.Kill()
		}
		if p.status.Port > 0 {
			_ = winutil.KillProcessesOnPort(p.status.Port)
		}
		p.status.Running = false
		p.status.PID = 0
	}
	return nil
}

func (s *Supervisor) GetServiceStatus(name string) (ServiceStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, exists := s.processes[name]
	if !exists {
		return ServiceStatus{Name: name, Running: false}, false
	}
	return p.status, true
}
