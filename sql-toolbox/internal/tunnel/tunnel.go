package tunnel

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
)

// SSHTunnel manages an SSH tunnel for database connections
type SSHTunnel struct {
	config       *models.SSHTunnelConfig
	sshClient    *ssh.Client
	listener     net.Listener
	localAddress string
	dbHost       string
	dbPort       int
	done         chan error
	closeOnce    sync.Once
}

// NewSSHTunnel creates and establishes a new SSH tunnel
func NewSSHTunnel(config *models.SSHTunnelConfig, dbHost string, dbPort int) (*SSHTunnel, error) {
	if config == nil || !config.Enabled {
		return nil, fmt.Errorf("SSH tunnel config is nil or disabled")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SSH tunnel config: %w", err)
	}

	tunnel := &SSHTunnel{
		config: config,
		dbHost: dbHost,
		dbPort: dbPort,
		done:   make(chan error, 1),
	}

	if err := tunnel.establish(); err != nil {
		return nil, err
	}

	return tunnel, nil
}

// establish connects to SSH server and creates local listener
func (t *SSHTunnel) establish() error {
	// Get authentication method
	authMethod, err := t.getAuthMethod()
	if err != nil {
		return fmt.Errorf("failed to get SSH auth method: %w", err)
	}

	// SSH client configuration
	sshConfig := &ssh.ClientConfig{
		User: t.config.User,
		Auth: []ssh.AuthMethod{authMethod},
		// TODO: Add host key verification for production use
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to SSH server
	sshAddr := fmt.Sprintf("%s:%d", t.config.Host, t.config.GetEffectivePort())
	sshClient, err := ssh.Dial("tcp", sshAddr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to establish SSH tunnel to %s: %w", sshAddr, err)
	}
	t.sshClient = sshClient

	// Create local listener
	localAddr := fmt.Sprintf("127.0.0.1:%d", t.config.LocalPort)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		sshClient.Close()
		return fmt.Errorf("failed to create local listener on %s: %w", localAddr, err)
	}
	t.listener = listener
	t.localAddress = listener.Addr().String()

	// Start forwarding in background
	go t.forward()

	return nil
}

// getAuthMethod returns the appropriate SSH authentication method based on config
func (t *SSHTunnel) getAuthMethod() (ssh.AuthMethod, error) {
	switch t.config.AuthMethod {
	case "key":
		return t.getKeyAuth()
	case "password":
		return ssh.Password(t.config.Password), nil
	case "agent":
		return t.getAgentAuth()
	default:
		return nil, fmt.Errorf("unsupported auth method: %s", t.config.AuthMethod)
	}
}

// getKeyAuth reads private key file and returns PublicKeys authentication
func (t *SSHTunnel) getKeyAuth() (ssh.AuthMethod, error) {
	keyPath, err := t.config.ExpandKeyFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to expand key file path: %w", err)
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("SSH key file not found: %s: %w", t.config.KeyFile, err)
	}

	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH private key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

// getAgentAuth connects to SSH agent and returns agent authentication
func (t *SSHTunnel) getAgentAuth() (ssh.AuthMethod, error) {
	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSock == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK environment variable not set")
	}

	conn, err := net.Dial("unix", sshAuthSock)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH agent: %w", err)
	}

	agentClient := agent.NewClient(conn)
	return ssh.PublicKeysCallback(agentClient.Signers), nil
}

// forward accepts local connections and forwards them through SSH tunnel
func (t *SSHTunnel) forward() {
	for {
		localConn, err := t.listener.Accept()
		if err != nil {
			// Listener closed, exit forwarding loop
			select {
			case t.done <- err:
			default:
			}
			return
		}

		go t.handleConnection(localConn)
	}
}

// handleConnection handles a single connection by forwarding through SSH tunnel
func (t *SSHTunnel) handleConnection(localConn net.Conn) {
	defer localConn.Close()

	// Dial remote database through SSH tunnel
	remoteAddr := fmt.Sprintf("%s:%d", t.dbHost, t.dbPort)
	remoteConn, err := t.sshClient.Dial("tcp", remoteAddr)
	if err != nil {
		// Database unreachable from bastion
		return
	}
	defer remoteConn.Close()

	// Bidirectional copy
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(remoteConn, localConn)
	}()

	go func() {
		defer wg.Done()
		io.Copy(localConn, remoteConn)
	}()

	wg.Wait()
}

// GetLocalAddress returns the local address string (e.g., "127.0.0.1:54321")
func (t *SSHTunnel) GetLocalAddress() string {
	return t.localAddress
}

// GetLocalPort returns the local port number
func (t *SSHTunnel) GetLocalPort() int {
	if t.listener == nil {
		return 0
	}
	_, portStr, err := net.SplitHostPort(t.localAddress)
	if err != nil {
		return 0
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return port
}

// HealthCheck sends SSH keepalive to verify tunnel is alive
func (t *SSHTunnel) HealthCheck() error {
	if t.sshClient == nil {
		return fmt.Errorf("SSH client is not connected")
	}

	// Send a keepalive request
	_, _, err := t.sshClient.SendRequest("keepalive@openssh.com", true, nil)
	if err != nil {
		return fmt.Errorf("SSH tunnel health check failed: %w", err)
	}

	return nil
}

// Close closes the tunnel listener and SSH client
func (t *SSHTunnel) Close() error {
	var closeErr error

	t.closeOnce.Do(func() {
		// Close listener first to stop accepting new connections
		if t.listener != nil {
			if err := t.listener.Close(); err != nil {
				closeErr = fmt.Errorf("failed to close listener: %w", err)
			}
		}

		// Close SSH client
		if t.sshClient != nil {
			if err := t.sshClient.Close(); err != nil {
				if closeErr == nil {
					closeErr = fmt.Errorf("failed to close SSH client: %w", err)
				}
			}
		}
	})

	return closeErr
}
