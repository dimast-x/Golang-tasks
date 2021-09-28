package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		panic("wat should I do")
	}
}

func parent() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
}



func child() {
	must(syscall.Sethostname([]byte("container")))
	must(syscall.Mount("ubuntu_fs", "ubuntu_fs", "", syscall.MS_BIND, ""))
	must(os.MkdirAll("ubuntu_fs/oldrootfs", 0700))
	//must(syscall.PivotRoot("ubuntu_fs", "ubuntu_fs/oldrootfs"))
	must(os.Chdir("/"))
	//must(syscall.Mount("proc", "proc", "proc", 0, ""))
	//must(syscall.Mount("something", "mytemp", "tmpfs", 0, ""))

	


	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}