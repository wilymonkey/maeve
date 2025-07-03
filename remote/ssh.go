package remote

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wilymonkey/maeve/cfg"
	"github.com/wilymonkey/maeve/local/hashsums"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func NewSSHClient(address string) (*ssh.Client, error) {
	user, host, port, err := parseAddress(address)
	if err != nil {
		return nil, fmt.Errorf("parse node address ⇒  %w", err)
	}

	key, err := os.ReadFile(cfg.Global.SSHKey)
	if err != nil {
		return nil, fmt.Errorf("read ssh key ⇒  %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parse ssh key ⇒  %w", err)
	}

	hostKeyCallback, err := knownhosts.New(cfg.Global.SSHKnownHosts)
	if err != nil {
		return nil, fmt.Errorf("create host key callback ⇒  %w", err)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback,
	}

	client, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		return nil, fmt.Errorf("dial ssh ⇒  %w", err)
	}
	return client, nil
}

func sendDir(client *ssh.Client, address string) error {
	// File transfer tasks
	tasks := []struct {
		localPath  string
		remotePath string
	}{
		{"/path/to/local/file1.txt", "/path/to/remote/file1.txt"},
		{"/path/to/local/file2.txt", "/path/to/remote/file2.txt"},
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go func(local, remote string) {
			defer wg.Done()
			if err := sendFile(client, local, remote); err != nil {
				errs <- fmt.Errorf("transfer %s: %v", local, err)
			}
		}(task.localPath, task.remotePath)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	return nil
}

func sendFile(client *ssh.Client, localPath, remotePath string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("session: %v", err)
	}
	defer session.Close()

	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat file: %v", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %v", err)
	}
	defer stdin.Close()

	if err := session.Start(fmt.Sprintf("scp -t %s", remotePath)); err != nil {
		return fmt.Errorf("start scp: %v", err)
	}

	fmt.Fprintf(stdin, "C0644 %d %s\n", stat.Size(), remotePath)
	if _, err := io.Copy(stdin, file); err != nil {
		return fmt.Errorf("copy file: %v", err)
	}
	fmt.Fprint(stdin, "\x00")

	return session.Wait()
}

func parseAddress(address string) (user, host, port string, err error) {
	parts := strings.Split(address, "@")
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid format: missing @")
	}
	user = parts[0]

	hostPort := strings.Split(parts[1], ":")
	if len(hostPort) != 2 {
		hostPort = append(hostPort, "22")
	}
	host = hostPort[0]
	port = hostPort[1]

	return user, host, port, nil
}

func remoteHashes(client *ssh.Client) ([]hashsums.FileHash, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("new ssh session ⇒  %w", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("getting stdOut pipe for session ⇒  %w", err)
	}

	err = session.Start(fmt.Sprintf("maeve --latest-hashes %s", cfg.Global.Name))
	if err != nil {
		return nil, fmt.Errorf("start maeve through ssh ⇒  %w", err)
	}

	var hashes []hashsums.FileHash
	err = gob.NewDecoder(stdout).Decode(&hashes)
	if err != nil {
		return nil, fmt.Errorf("decode stdout ⇒  %w", err)
	}

	if err = session.Wait(); err != nil {
		return nil, fmt.Errorf("session wait for cmd ⇒  %w", err)
	}

	return hashes, nil
}

// MaxSessions tries to open as many "session" channels as possible
// on the given SSH client, and returns the maximum before failure.
func maxSessions(client *ssh.Client) (int, error) {
	var sessions []*ssh.Session
	defer func() {
		for _, s := range sessions {
			s.Close()
		}
	}()

	count := 0
	for {
		sess, err := client.NewSession()
		if err != nil {
			break
		}
		sessions = append(sessions, sess)
		count++
	}

	if count == 0 {
		return 0, fmt.Errorf("unable to open any SSH sessions")
	}
	return count, nil
}

type ProgressWriter struct {
	ID           int
	Writer       io.Writer
	Total        int64
	Transferred  int64
	Percent      int
	LastReported time.Time
}

func NewProgressWriter(id int, total int64, writer io.Writer) ProgressWriter {
	return ProgressWriter{ID: id, Writer: writer, Total: total}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	pw.Transferred += int64(n)

	now := time.Now()
	if now.Sub(pw.LastReported) > time.Second || pw.Transferred == pw.Total {
		pw.Percent = int(float64(pw.Transferred) / float64(pw.Total) * 100)
		pw.LastReported = now
	}

	return n, err
}
