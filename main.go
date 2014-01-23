package main

import "syscall"
import "errors"
import "os"

const (
	craw = "/dev/mmcblk0p7"
)

func main() {
	e := do()
	if e != nil {
		println(e.Error())
		println("Spawning sh")
		e = syscall.Exec("/bin/sh", nil, nil)
		println(e.Error())
	}
}

func do() error {
	os.Setenv("PATH", "/sbin:/bin:/usr/sbin:/usr/bin")
	os.Mkdir("/proc", 0755)
	syscall.Mount("proc", "/proc", "proc", 0, "")
	println("Spawning /bin/cryptsetup.cryptoinit")
	p, _, e := syscall.StartProcess("/bin/cryptsetup.cryptoinit",
		[]string{"/bin/cryptsetup.cryptoinit", "luksOpen", craw, "crypto"},
		&syscall.ProcAttr{"", nil, []uintptr{0, 1, 2}, &syscall.SysProcAttr{}},
	)
	if e != nil {
		return e
	}
	var ws syscall.WaitStatus
	syscall.Wait4(p, &ws, 0, nil)
	if ws.ExitStatus() != 0 {
		return errors.New("cryptsetup failed")
	}
	println("Mounting...")
	os.Mkdir("/mnt", 0755)
	e = Mount("/dev/mapper/crypto", "/mnt")
	if e != nil {
		return e
	}
	os.Chdir("/")
	//	println("And executing switch_root...")
	//	return syscall.Exec("/sbin/switch_root", []string{"/sbin/switch_root", "/mnt", "/sbin/init"}, nil)

	println("Pivoting...")
	e = syscall.PivotRoot("/mnt", "/mnt/usr")
	if e != nil {
		return e
	}
	os.Chdir("/")
	println("And executing real init (/bin/systemd)...")
	syscall.Exec("/bin/systemd", []string{"/bin/systemd"}, nil)
	println("Or executing real init (/sbin/init)...")
	return syscall.Exec("/sbin/init", []string{"/sbin/init"}, nil)
}

func Mount(from, to string) error {
	return syscall.Mount(from, to, "ext4", syscall.MS_RELATIME, "")
}
