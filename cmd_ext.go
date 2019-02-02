package server

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

// commandMd5 responds to the MD5 FTP command.
// It returns the Md5 in hex-string.
type commandMd5 struct{}

func (cmd commandMd5) IsExtend() bool {
	return false
}

func (cmd commandMd5) RequireParam() bool {
	return true
}

func (cmd commandMd5) RequireAuth() bool {
	return true
}

func (cmd commandMd5) Execute(conn *Conn, param string) {
	path := conn.buildPath(param)
	_, data, err := conn.driver.GetFile(path, 0)
	if err == nil {
		defer data.Close()
		h := md5.New()
		_, err := io.Copy(h, data)
		if err != nil {
			conn.writeMessage(551, "File not available")
		} else {
			conn.writeMessage(213, fmt.Sprintf("%x", h.Sum(nil)))
		}
	} else {
		conn.writeMessage(450, "File not found")
	}

}

// commandSite responds to the SITE EXEC FTP command
type commandSite struct{}

func (cmd commandSite) IsExtend() bool {
	return true
}

func (cmd commandSite) RequireParam() bool {
	return true
}

func (cmd commandSite) RequireAuth() bool {
	return true
}

func (cmd commandSite) Execute(conn *Conn, param string) {
	sub, params := conn.parseLine(param)
	switch strings.ToUpper(sub) {
	case "EXEC":
		out := popen(exec.Command("bash", "-c", params), nil)
		//out =bytes.Replace(out,[]byte("\n"),[]byte(" "),-1)
		conn.writeMessageMultiline(200, "\n"+string(out))
	default:
		conn.writeMessage(200, sub+":"+params)
	}

}
func popen(cmd *exec.Cmd, input []byte) []byte {

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		stdin.Write(input)
	}()

	out, err := cmd.Output()
	if err != nil {
		log.Print(err, out)
	}

	return out
}
