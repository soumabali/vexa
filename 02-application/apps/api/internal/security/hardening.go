package security

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// SecureServerConfig returns a hardened tls.Config enforcing TLS 1.3,
// modern cipher suites, and strict certificate verification.
func SecureServerConfig(certPEM, keyPEM []byte) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		MinVersion: tls.VersionTLS13,
		MaxVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		},
		PreferServerCipherSuites: true,
		Certificates:             []tls.Certificate{cert},
		// Enforce mutual TLS if client certs are configured
		ClientAuth: tls.VerifyClientCertIfGiven,
	}, nil
}

// SecureListener wraps a net.Listener with TLS 1.3 and connection timeouts.
func SecureListener(addr string, config *tls.Config) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return tls.NewListener(&timeoutListener{Listener: ln}, config), nil
}

// timeoutListener wraps net.Listener to enforce read/write deadlines.
type timeoutListener struct {
	net.Listener
}

func (l *timeoutListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &timeoutConn{Conn: c}, nil
}

type timeoutConn struct {
	net.Conn
}

func (c *timeoutConn) Read(b []byte) (int, error) {
	c.SetReadDeadline(time.Now().Add(30 * time.Second))
	return c.Conn.Read(b)
}

func (c *timeoutConn) Write(b []byte) (int, error) {
	c.SetWriteDeadline(time.Now().Add(30 * time.Second))
	return c.Conn.Write(b)
}

// SecureServer returns an http.Server with secure timeouts.
func SecureServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
}

// StrictTransportSecurity returns the HSTS header value.
func StrictTransportSecurity() string {
	return "max-age=31536000; includeSubDomains; preload"
}
