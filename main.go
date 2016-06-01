// listen-like-systemd is a standalone command that emulates the
// "socket activation" TCP file descriptor passing of Systemd. It is
// intended for testing software that will be deployed via Systemd in
// production.
//
// We don't handle D-Bus or the lovecraftian aspects of systemd, but
// just the subset that makes TCP listening work with
// github.com/coreos/go-systemd/activation; see
// http://www.freedesktop.org/software/systemd/man/sd_listen_fds.html
// for more.
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var prog = filepath.Base(os.Args[0])

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s [IP]:PORT[,...] CMD [ARG..]\n", prog)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 2 {
		usage()
		os.Exit(2)
	}
	addrs := strings.Split(flag.Arg(0), ",")
	for i, addr := range addrs {
		network := "tcp"
		if strings.HasPrefix(addr, "/") {
			network = "unix"
			os.Remove(addr)
		}
		l, err := net.Listen(network, addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: cannot listen: %v\n", prog, err)
			os.Exit(1)
		}

		var f *os.File
		switch lt := l.(type) {
		case *net.TCPListener:
			f, err = lt.File()
		case *net.UnixListener:
			f, err = lt.File()
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: cannot get listening FD: %v\n", prog, err)
			os.Exit(1)
		}
		if err := syscall.Dup2(int(f.Fd()), 3+i); err != nil {
			fmt.Fprintf(os.Stderr, "%s: cannot duplicate listening FD: %v\n", prog, err)
			os.Exit(1)
		}
	}
	os.Setenv("LISTEN_FDS", strconv.FormatUint(uint64(len(addrs)), 10))
	os.Setenv("LISTEN_PID", strconv.FormatUint(uint64(os.Getpid()), 10))

	exe, err := exec.LookPath(flag.Arg(1))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: cannot find executable: %v\n", prog, err)
		os.Exit(1)
	}
	args := flag.Args()[1:]
	args[0] = exe
	err = syscall.Exec(exe, args, os.Environ())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: cannot execute command: %v\n", prog, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "%s: successful exec left us running!\n", prog)
	os.Exit(3)
}
