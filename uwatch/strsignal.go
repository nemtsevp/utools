package main

import (
	"syscall"
)

var signals = [...]struct {
	code syscall.Signal
	name string
}{
	{syscall.SIGABRT, "SIGABRT"},
	{syscall.SIGALRM, "SIGALRM"},
	{syscall.SIGBUS, "SIGBUS"},
	{syscall.SIGCHLD, "SIGCHLD"},
	{syscall.SIGCLD, "SIGCLD"},
	{syscall.SIGCONT, "SIGCONT"},
	{syscall.SIGFPE, "SIGFPE"},
	{syscall.SIGHUP, "SIGHUP"},
	{syscall.SIGILL, "SIGILL"},
	{syscall.SIGINT, "SIGINT"},
	{syscall.SIGIO, "SIGIO"},
	{syscall.SIGIOT, "SIGIOT"},
	{syscall.SIGKILL, "SIGKILL"},
	{syscall.SIGPIPE, "SIGPIPE"},
	{syscall.SIGPOLL, "SIGPOLL"},
	{syscall.SIGPROF, "SIGPROF"},
	{syscall.SIGPWR, "SIGPWR"},
	{syscall.SIGQUIT, "SIGQUIT"},
	{syscall.SIGSEGV, "SIGSEGV"},
	{syscall.SIGSTKFLT, "SIGSTKFLT"},
	{syscall.SIGSTOP, "SIGSTOP"},
	{syscall.SIGSYS, "SIGSYS"},
	{syscall.SIGTERM, "SIGTERM"},
	{syscall.SIGTRAP, "SIGTRAP"},
	{syscall.SIGTSTP, "SIGTSTP"},
	{syscall.SIGTTIN, "SIGTTIN"},
	{syscall.SIGTTOU, "SIGTTOU"},
	{syscall.SIGURG, "SIGURG"},
	{syscall.SIGUSR1, "SIGUSR1"},
	{syscall.SIGUSR2, "SIGUSR2"},
	{syscall.SIGVTALRM, "SIGVTALRM"},
	{syscall.SIGWINCH, "SIGWINCH"},
	{syscall.SIGXCPU, "SIGXCPU"},
	{syscall.SIGXFSZ, "SIGXFSZ"},
	{syscall.SIGUNUSED, "SIGUNUSED"},
}

func strsignal(sig syscall.Signal) string {
	for _, s := range signals {
		if sig == s.code {
			return s.name
		}
	}
	return "UNKNOWN"
}
