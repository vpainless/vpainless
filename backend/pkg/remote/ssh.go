package remote

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Default timeout value for dialing into SSH servers.
// Change to 0 for no timeout.
var Timeout time.Duration = 10 * time.Second

// Dial creates a ssh client to the specified address.
// The returned client should be closed when done.
func Dial(ipv4 net.IP, privateKey []byte, username string) (*ssh.Client, error) {
	key, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("error parsing ssh private key: %w", err)
	}

	config := ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(key)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         Timeout,
	}

	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", ipv4, 22), &config)
}

// Execute executes a command through ssh. It waits for the command to finish
// and returns the output. Make sure that the pass command exists.
func Execute(conn *ssh.Client, cmd string) (string, error) {
	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("error creating ssh session: %w", err)
	}
	defer session.Close()

	b, err := session.Output(cmd)
	if err != nil {
		return "", fmt.Errorf("error executing command `%s`: %w", cmd, err)
	}

	return string(b), nil
}

// Start starts a command through ssh. It does not waits for the command to finish.
func Start(conn *ssh.Client, cmd string) error {
	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("error creating ssh session: %w", err)
	}
	defer session.Close()

	err = session.Start(cmd)
	if err != nil {
		return fmt.Errorf("error starting command `%s`: %w", cmd, err)
	}

	return nil
}

// UploadFile uploads a file to the remote machine using sftp protocol
// over an SSH connection
func UploadFile(conn *ssh.Client, filePath string, content io.Reader) error {
	client, err := sftp.NewClient(conn)
	if err != nil {
		return fmt.Errorf("error creating sftp client: %w", err)
	}
	defer client.Close()

	f, err := client.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file through sftp: %w", err)
	}

	if _, err := io.Copy(f, content); err != nil {
		return fmt.Errorf("error writing file through sftp: %w", err)
	}

	return nil
}
