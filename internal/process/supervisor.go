package process

import (
	"fmt"
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
}

func NewSupervisor() *Supervisor {
	return &Supervisor{
		processes: make(map[string]*managedProcess),
	}
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
	if dir == "" && filepath.Base(exe) != exe {
		dir = filepath.Dir(exe)
	}
	if dir != "" {
		if absDir, err := filepath.Abs(dir); err == nil {
			dir = absDir
		}
	}

	cmd := winutil.Command(exe, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if len(env) > 0 {
		cmd.Env = append(cmd.Environ(), env...)
	}
	if err := cmd.Start(); err != nil {
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
		s.mu.Lock()
		defer s.mu.Unlock()
		if current, ok := s.processes[name]; ok && current == p {
			current.status.Running = false
			current.status.PID = 0
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
