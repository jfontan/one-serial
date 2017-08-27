package main

// Code based on the following examples:
//
//   https://github.com/gliderlabs/ssh/blob/master/_examples/ssh-pty/pty.go
//   https://github.com/golang/crypto/blob/master/ssh/example_test.go

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/OpenNebula/goca"
	"github.com/gliderlabs/ssh"
	"github.com/kr/pty"
)

func GetHostAndKeys(id uint) (hostname, keys string) {
	vm := goca.NewVM(id)

	err := vm.Info()
	if err != nil {
		log.Print(err)
		return "", ""
	}

	keys, _ = vm.XPath("/VM/TEMPLATE/CONTEXT/SSH_PUBLIC_KEY")
	hostname, _ = vm.XPath("/VM/HISTORY_RECORDS/HISTORY/HOSTNAME")

	return hostname, keys
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func sessionHandler(s ssh.Session) {
	user := s.User()
	id, _ := strconv.Atoi(user)
	hostname, _ := GetHostAndKeys(uint(id))

	vm_name := fmt.Sprintf("one-%s", user)

	cmd := exec.Command("ssh", "-t", hostname, "virsh", "-c", "qemu:///system", "console", vm_name)

	ptyReq, winCh, isPty := s.Pty()
	if isPty {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
		f, err := pty.Start(cmd)
		if err != nil {
			panic(err)
		}
		go func() {
			for win := range winCh {
				setWinsize(f, win.Width, win.Height)
			}
		}()
		go func() {
			io.Copy(f, s) // stdin
		}()
		io.Copy(s, f) // stdout
	} else {
		io.WriteString(s, "No PTY requested.\n")
		s.Exit(1)
	}
}

func publicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	user := ctx.User()
	id, _ := strconv.Atoi(user)

	log.Println("User: ", user)

	_, ssh_keys := GetHostAndKeys(uint(id))

	bytes := []byte(ssh_keys)
	for len(bytes) > 0 {
		pubkey, _, _, rest, err := ssh.ParseAuthorizedKey(bytes)
		if err != nil {
			log.Print(err)
		}

		bytes = rest

		if ssh.KeysEqual(key, pubkey) {
			return true
		}
	}

	return false
}

func main() {
	hostKey := ssh.HostKeyFile("./id_rsa")

	ssh.Handle(sessionHandler)
	auth := ssh.PublicKeyAuth(publicKeyHandler)

	log.Fatal(ssh.ListenAndServe(":2222", nil, hostKey, auth))
}
