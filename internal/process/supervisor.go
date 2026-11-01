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

	// Check if already running
	if p, exists := s.processes[name]; exists && p.status.Running {
		return nil
	}

	cmd := winutil.Command(exe, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	// Prepare log file
	logsDir := s.getLogsDir()
	_ = os.MkdirAll(logsDir, 0755)
	logFile, err := os.OpenFile(filepath.Join(logsDir, fmt.Sprintf("%s.log", name)), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	// Set process group on Windows so children can be killed
	winutil.HideWindow(cmd)

	if err := cmd.Start(); err != nil {
		if logFile != nil {
			_ = logFile.Close()
		}
		return fmt.Errorf("failed to start %s: %w", name, err)
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

	// Monitor in background
	go func() {
		_ = cmd.Wait()
		if logFile != nil {
			_ = logFile.Close()
		}

		s.mu.Lock()
		p.status.Running = false
		p.status.PID = 0
		onExit := s.OnExit
		s.mu.Unlock()

		if onExit != nil {
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

func (s *Supervisor) GetAllServices() []ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]ServiceStatus, 0, len(s.processes))
	for _, p := range s.processes {
		res = append(res, p.status)
	}
	return res
}

func (s *Supervisor) RemoveService(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.processes, name)
}

func (s *Supervisor) SetServiceStatus(name string, status ServiceStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.processes == nil {
		s.processes = make(map[string]*managedProcess)
	}
	s.processes[name] = &managedProcess{
		status: status,
	}
}
