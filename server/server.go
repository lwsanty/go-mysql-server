package server // import "github.com/src-d/go-mysql-server/server"

import (
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/src-d/go-mysql-server"
	"github.com/src-d/go-mysql-server/auth"

	"vitess.io/vitess/go/mysql"
)

// Server is a MySQL server for SQLe engines.
type Server struct {
	Listener *mysql.Listener
}

// Config for the mysql server.
type Config struct {
	// Protocol for the connection.
	Protocol string
	// Address of the server.
	Address string
	// Auth of the server.
	Auth auth.Auth
	// Tracer to use in the server. By default, a noop tracer will be used if
	// no tracer is provided.
	Tracer opentracing.Tracer

	ConnReadTimeout  time.Duration
	ConnWriteTimeout time.Duration
}

// NewDefaultServer creates a Server with the default session builder.
func NewDefaultServer(cfg Config, e *sqle.Engine) (*Server, error) {
	return NewServer(cfg, e, DefaultSessionBuilder)
}

// NewServer creates a server with the given protocol, address, authentication
// details given a SQLe engine and a session builder.
func NewServer(cfg Config, e *sqle.Engine, sb SessionBuilder) (*Server, error) {
	var tracer opentracing.Tracer
	if cfg.Tracer != nil {
		tracer = cfg.Tracer
	} else {
		tracer = opentracing.NoopTracer{}
	}

	if cfg.ConnReadTimeout < 0 {
		cfg.ConnReadTimeout = 0
	}

	if cfg.ConnWriteTimeout < 0 {
		cfg.ConnWriteTimeout = 0
	}

	handler := NewHandler(e, NewSessionManager(sb, tracer, cfg.Address))
	a := cfg.Auth.Mysql()
	l, err := mysql.NewListener(cfg.Protocol, cfg.Address, a, handler, cfg.ConnReadTimeout, cfg.ConnWriteTimeout)
	if err != nil {
		return nil, err
	}

	return &Server{Listener: l}, nil
}

// Start starts accepting connections on the server.
func (s *Server) Start() error {
	s.Listener.Accept()
	return nil
}

// Close closes the server connection.
func (s *Server) Close() error {
	s.Listener.Close()
	return nil
}
