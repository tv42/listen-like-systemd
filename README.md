# listen-like-systemd -- Open a listening socket and pass it to subprocess

This program performs the parent side of [systemd's socket activation](https://0pointer.de/blog/projects/socket-activation.html) dance.
This is useful for e.g. development, where you want to run a program like it was socket activated, but not create a systemd service for it.

`listen-like-system` has pretty much been obsoleted by systemd's own [systemd-socket-activate](https://www.freedesktop.org/software/systemd/man/systemd-socket-activate.html) command, though I still prefer the calling convention used here.

## Bugs

### can overwrite actively used fds 3, 4, ...

- systemd design forces passed fds to be numbered starting from 3
- Go runtime on BSD derivatives uses an fd around 3-5 to hold open kqueue fd
- there's a race in where runtime could try to use the supposedly-kqueue fd after our dup2
- could push dup2 later, to minimize window; seems impossible to actually fix, with pure Go
- could fork and leave parent running, and avoid the issue by using <https://pkg.go.dev/os#ProcAttr.Files> to adjust the FD numbers, but leaving the parent around is ugly and would mess up `LISTEN_PID`
