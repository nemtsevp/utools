package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mattn/go-isatty"
)

const (
	envVar = "UWATCH_CHILD"
)

var (
	verbose bool
)

func main() {
	var (
		pollInterval time.Duration
		deathTimeout time.Duration
	)

	flag.DurationVar(&pollInterval, "p", time.Millisecond*100, "poll interval")
	flag.DurationVar(&deathTimeout, "t", time.Millisecond*1000, "death timeout")
	flag.BoolVar(&verbose, "v", false, "verbose mode")
	flag.Parse()

	if flag.NArg() == 0 {
		die("invalid arguments: missing command")
	}

	var child *exec.Cmd
	if _, ok := os.LookupEnv(envVar); !ok {
		if err := os.Setenv(envVar, ""); err != nil {
			die("setenv: %v", err)
		}
		child = exec.Command(os.Args[0], os.Args[1:]...)
		deathTimeout = 0
	} else {
		if err := os.Unsetenv(envVar); err != nil {
			die("unsetenv: %v", err)
		}
		if err := syscall.Setpgid(os.Getpid(), os.Getpid()); err != nil {
			die("setpgid: %v", err)
		}
		child = exec.Command(flag.Args()[0], flag.Args()[1:]...)
	}

	sigs := make(chan os.Signal)
	signal.Notify(sigs,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM)

	child.Stdin = os.Stdin
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr

	if isatty.IsTerminal(os.Stdin.Fd()) {
		child.Stdin = rpipe(os.Stdin)
	}

	if isatty.IsTerminal(os.Stdout.Fd()) {
		child.Stdout = wpipe(os.Stdout)
	}

	if isatty.IsTerminal(os.Stderr.Fd()) {
		child.Stderr = wpipe(os.Stderr)
	}

	msginfo("spawn: %s", strings.Join(child.Args, " "))

	if err := child.Start(); err != nil {
		die("spawn: %v", err)
	}

	msginfo("spawn: child pid is %v", child.Process.Pid)

	child.Stdin.(io.Closer).Close()
	child.Stdout.(io.Closer).Close()
	child.Stderr.(io.Closer).Close()

	done := make(chan error)
	go func() {
		done <- child.Wait()
	}()

	var (
		pollTick <-chan time.Time
		hardKill <-chan time.Time
	)

	if pollInterval > 0 {
		pollTick = time.NewTicker(pollInterval).C
	}

	killAll := func(sig syscall.Signal) {
		signal.Ignore(sig)
		for _, pid := range []int{child.Process.Pid, -os.Getpid()} {
			msginfo("sending %v to pid %v", strsignal(sig), pid)
			if err := syscall.Kill(pid, sig); err != nil && err != syscall.ESRCH {
				die("kill(%v, %v): %v", pid, strsignal(sig), err)
			}
		}
	}

	softKillAll := func(sig syscall.Signal) {
		killAll(syscall.SIGHUP)
		if hardKill == nil && deathTimeout > 0 {
			hardKill = time.After(deathTimeout)
		}
	}

	for {
		select {
		case err := <-done:
			msginfo("received SIGCHLD")
			killAll(syscall.SIGTERM)
			if err == nil {
				exit(0)
			} else {
				exit(1)
			}

		case sig := <-sigs:
			msginfo("received %v", strsignal(sig.(syscall.Signal)))
			softKillAll(sig.(syscall.Signal))

		case <-pollTick:
			if os.Getppid() == 1 {
				msginfo("detected parent death")
				softKillAll(syscall.SIGHUP)
			}

		case <-hardKill:
			killAll(syscall.SIGKILL)
			exit(1)
		}
	}
}

func die(f string, args ...interface{}) {
	msg(f, args...)
	exit(1)
}

func exit(code int) {
	msginfo("existing with code %v", code)
	os.Exit(code)
}

func msginfo(f string, args ...interface{}) {
	if verbose {
		msg(f, args...)
	}
}

func msg(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmt.Sprintf("%s[%v]: %s\n", os.Args[0], os.Getpid(), f), args...)
}

func rpipe(std io.Reader) io.Reader {
	r, w, err := os.Pipe()
	if err != nil {
		die("pipe: %v", err)
	}

	go func() {
		io.Copy(w, std)
		w.Close()
	}()

	return r
}

func wpipe(std io.Writer) io.Writer {
	r, w, err := os.Pipe()
	if err != nil {
		die("pipe: %v", err)
	}

	go func() {
		io.Copy(os.Stdout, r)
		r.Close()
	}()

	return w
}
