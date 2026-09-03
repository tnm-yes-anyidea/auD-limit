package player

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Minimal mpv controller (POSIX unix-socket based). This is intentionally small
// and used only for a first-pass implementation.

type MPV struct{
	proc *exec.Cmd
	socketPath string
	conn net.Conn
}

func StartMPV() (*MPV, error) {
	p := &MPV{}
	// use temp unix socket
	sock := filepath.Join(os.TempDir(), fmt.Sprintf("audlimiter-%d.sock", os.Getpid()))
	p.socketPath = sock
	cmd := exec.Command("mpv", "--idle=yes", "--no-terminal", fmt.Sprintf("--input-ipc-server=%s", sock))
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil { return nil, err }
	p.proc = cmd
	// wait for socket to appear
	for i:=0;i<100;i++ {
		if _, err := os.Stat(sock); err == nil { break }
		time.Sleep(10 * time.Millisecond)
	}
	// connect
	conn, err := net.DialTimeout("unix", sock, time.Second)
	if err != nil { return p, nil } // return without conn, playback will be disabled
	p.conn = conn
	return p, nil
}

func (p *MPV) Close() {
	if p.conn != nil { p.conn.Close() }
	if p.proc != nil {
		p.proc.Process.Kill()
		p.proc = nil
	}
}

func (p *MPV) send(cmd interface{}) error {
	if p.conn == nil { return fmt.Errorf("no mpv socket") }
	b, err := json.Marshal(cmd)
	if err != nil { return err }
	b = append(b, '\n')
	_, err = p.conn.Write(b)
	return err
}

func (p *MPV) Play(path string) error {
	cmd := map[string]interface{}{"command": []interface{}{"loadfile", path, "replace"}}
	return p.send(cmd)
}
