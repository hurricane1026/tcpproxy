package main
import (
    "net"
    "fmt"
    "bufio"
    "os"
    "sync"
    "time"
)

func openLogWithFd(fd *os.File) *bufio.Writer{
    //return log.New(fd, "", log.Ldate|log.Ltime|log.Lmicroseconds)
    return bufio.NewWriter(fd)
}

func openLog(path string) (logger *bufio.Writer, fd *os.File, err error) {
    if fd, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644); err == nil {
           logger = openLogWithFd(fd)
        }
    return
}

func main() {
        if len(os.Args) != 4 {
                fatal("usage: netfwd localIp:localPort remoteIp:remotePort log_path")
        }
        localAddr := os.Args[1]
        remoteAddr := os.Args[2]
        log_path := os.Args[3]
        local, err := net.Listen("tcp", localAddr)
        if local == nil {
                fatal("cannot listen: %v", err)
        }

        // start logging
        logger, _ , _ := openLog(log_path)
        for {
                conn, err := local.Accept()
                if conn == nil {
                        fatal("accept failed: %v", err)
                }
                go forward(conn, remoteAddr, logger)
        }
}

var l *sync.Mutex = new(sync.Mutex)

func Copy(r *bufio.Reader, w *bufio.Writer, addr_title string, need_log bool, logger *bufio.Writer) (e error) {

    for {
    buf := make([]byte, 1024)
    //l.Lock()
    //defer l.Unlock()
    var n int
    if n, e = r.Read(buf); e != nil {
        break
    }
    if need_log {
        log_string := fmt.Sprintf("%v %s %s\n", time.Now(), addr_title, string(buf))
        logger.WriteString(log_string)
    }

    if _, e = w.Write(buf[0:n]); e != nil {
        break
    }
    if e = w.Flush(); e != nil {
        break
    }
    }
    return e
}

func forward(local net.Conn, remoteAddr string, logger *bufio.Writer) {
    local_reader := bufio.NewReader(local)
    local_writer := bufio.NewWriter(local)
    remote, err := net.Dial("tcp",remoteAddr)
    if remote == nil {
        fmt.Fprintf(os.Stderr, "remote dial failed: %v\n", err)
        return
    }
    addr_title := local.RemoteAddr().String()
    remote_reader := bufio.NewReader(remote)
    remote_writer := bufio.NewWriter(remote)
    go Copy(local_reader, remote_writer, addr_title, true, logger)
    go Copy(remote_reader, local_writer, "", false, nil)
}

func fatal(s string, a ... interface{}) {
    fmt.Fprintf(os.Stderr, "netfwd: %s\n", fmt.Sprintf(s, a))
    os.Exit(2)
}
