package srvmanager

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
)

type command struct {
	Name string
	Run  func() int
}

var commands = []*command{
	cmdStart,
	cmdStop,
	cmdRestart,
}

var (
	mutex       sync.Mutex
	osFile      *os.File
	closerFuncs []func()       = make([]func(), 0)
	doneChan    chan bool      = make(chan bool, 1)
	waitChan    chan bool      = make(chan bool, 1)
	signalChan  chan os.Signal = make(chan os.Signal, 1)
	appPath     string         = os.Args[0]
	appName     string         = path.Base(appPath)
	pidPath     string         = path.Dir(appPath) + "/" + appName + ".pid"
)

func InitManage() {
	if len(os.Args) > 1 {
		name := os.Args[1]
		if name == "start" {
			pidFile := newPIDFile(pidPath)
			if err := pidFile.Lock(); err != nil {
				Println(appName + " already running!")
				os.Exit(1)
				return
			}
			_ = pidFile.Unlock()
		} else {
			for _, cmd := range commands {
				if cmd.Name == name {
					os.Exit(cmd.Run())
				}
			}
		}
	}
	createPIDFileLockAndSet(pidPath, os.Getpid())
	//signalListen(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	signalListen(signalChan, syscall.SIGTERM, syscall.SIGINT)
}

func Bind(fn func()) {
	mutex.Lock()
	c := make([]func(), 0, len(closerFuncs)+1)
	c = append(c, fn)
	closerFuncs = append(c, closerFuncs...)
	mutex.Unlock()
}

func Wait() {
	<-waitChan
	Println(appName + " stopping... ")
	<-doneChan
	Println(appName + " stopped!\n")
	os.Exit(0)
}

func signalListen(signalChan chan os.Signal, siganls ...os.Signal) {
	signal.Notify(signalChan, siganls...)
	go func() {
		<-signalChan
		mutex.Lock()
		defer mutex.Unlock()
		waitChan <- true
		for _, fn := range closerFuncs {
			fn()
		}
		doneChan <- true
	}()
}

func createPIDFileLockAndSet(path string, pid int) {
	pidFile := newPIDFile(path)
	if err := pidFile.Lock(); err != nil {
		Fatalln("PID not lock!")
	}
	if err := pidFile.Set(pid); err != nil {
		Fatalln("PID not set!")
	}
}

type pidFile struct {
	*os.File
}

func newPIDFile(path string) *pidFile {
	var err error
	osFile, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		Fatalln("PID file not open!")
	}
	return &pidFile{osFile}
}

func (file *pidFile) Set(pid int) error {
	file.Truncate(0)
	file.Seek(0, os.SEEK_SET)
	_, err := fmt.Fprint(file, pid)
	if err != nil {
		return Errorln("PID not save!")
	}
	return nil
}

func (file *pidFile) Get() (int, error) {
	var pid int = 0
	_, err := fmt.Fscan(file, &pid)
	if err != nil {
		return 0, Errorln("PID not read!")
	}
	return pid, nil
}

func (file *pidFile) Lock() error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func (file *pidFile) Unlock() error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
