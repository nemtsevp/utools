package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func die(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmt.Sprintf("%s[%v]: %s\n", os.Args[0], os.Getpid(), f), args...)
	os.Exit(1)
}

func killall(pid int, sig syscall.Signal) {
	signal.Ignore(sig)
	for _, p := range []int{pid, -os.Getpid()} {
		if err := syscall.Kill(p, sig); err != nil && err != syscall.ESRCH {
			die("kill(%v, %v): %v", p, sig, err)
		}
	}
}

const (
	procType   = "PROC_TYPE"
	parentProc = "1"
)

func main() {
	var deathTimeout time.Duration
	flag.DurationVar(&deathTimeout, "t", time.Millisecond*200, "death timeout")
	flag.Parse()

	sigs := make(chan os.Signal)
	signal.Notify(sigs,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM)

	if flag.NArg() == 0 {
		die("command was not specified")
	}

	var child *exec.Cmd
	if os.Getenv(procType) == parentProc {
		// it's a child process
		if err := syscall.Setpgid(os.Getpid(), os.Getpid()); err != nil {
			die("setpgid: %v", err)
		}

		child = exec.Command(flag.Args()[0], flag.Args()[1:]...)
	} else {
		// it's a parent process
		child = exec.Command(os.Args[0], os.Args[1:]...)
		if err := os.Setenv(procType, parentProc); err != nil {
			die("setenv: %v", err)
		}
	}

	defer func() {
		if typ := os.Getenv(procType); typ == parentProc {
			if err := os.Unsetenv(procType); err != nil {
				die("unsetenv: %v", err)
			}
		}
	}()

	child.Stdin = os.Stdin
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr

	if err := child.Start(); err != nil {
		die("start: %v", err)
	}

	done := make(chan error)
	go func() {
		done <- child.Wait()
	}()

	os.Stdin.Close()
	os.Stdout.Close()

	ticker := time.NewTicker(deathTimeout)
	waitdeath := false

	for {
		select {
		case err := <-done:
			killall(child.Process.Pid, syscall.SIGTERM)
			if err == nil {
				os.Exit(0)
			} else {
				os.Exit(1)
			}

		case sig := <-sigs:
			killall(child.Process.Pid, sig.(syscall.Signal))
			waitdeath = true

		case <-ticker.C:
			if waitdeath {
				killall(child.Process.Pid, syscall.SIGKILL)
				os.Exit(1)
			} else {
				if os.Getppid() == 1 {
					killall(child.Process.Pid, syscall.SIGHUP)
					waitdeath = true
				}
			}
		}
	}
}
