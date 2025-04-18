package http

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"sync"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/tavsec/gin-healthcheck/checks"
	"github.com/tavsec/gin-healthcheck/config"
	timeout "github.com/vearne/gin-timeout"
	"github.com/xBlaz3kx/DevX/observability"
	"github.com/xBlaz3kx/DevX/tls"
	"go.uber.org/zap"
)

var once sync.Once

// Http configuration for the API with TLS settings
type Configuration struct {
	// Address is the address of the HTTP server
	Address string `yaml:"address" json:"address" mapstructure:"address"`

	CORS *CORS `yaml:"cors" json:"cors" mapstructure:"cors"`

	// TLS is the TLS configuration for the HTTP server
	TLS tls.TLS `mapstructure:"tls" yaml:"tls" json:"tls"`
}

type Server struct {
	config Configuration
	router *gin.Engine
	server *http.Server
	obs    observability.Observability
}

func NewServer(config Configuration, obs observability.Observability, optionFuncs ...func(*Options)) *Server {
	options := newOptions()
	router := gin.New()
	logger := obs.Log().Logger

	// Apply options
	for _, optionFunc := range optionFuncs {
		optionFunc(options)
	}

	once.Do(func() {
		gin.DebugPrintFunc = func(format string, values ...interface{}) {
			// Remove newlines and tabs from the format string
			regex := regexp.MustCompile(`[\n\t]`)
			logger.Sugar().Debugf(regex.ReplaceAllString(format, ""), values...)
		}
	})

	obs.SetupGinMiddleware(router)

	router.NoRoute(func(context *gin.Context) {
		context.JSON(http.StatusNotFound, ErrorPayload{
			Error:       "Not Found",
			Description: "The requested resource was not found",
		})
	})

	router.NoMethod(func(context *gin.Context) {
		context.JSON(http.StatusMethodNotAllowed, ErrorPayload{
			Error:       "Not Allowed",
			Description: "Method not allowed",
		})
	})

	router.Use(
		ginzap.RecoveryWithZap(logger, true),
		ginzap.Ginzap(logger, time.RFC3339, true),
		timeout.Timeout(
			timeout.WithTimeout(options.timeout),
			timeout.WithErrorHttpCode(http.StatusServiceUnavailable),
			timeout.WithDefaultMsg(EmptyResponse{}),
		),
		errorHandler,
	)

	return &Server{
		config: config,
		router: router,
	}
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

// Run starts the HTTP server with the given healthchecks.
func (s *Server) Run(checks ...checks.Check) {
	// swagger:route GET /healthz healthCheck internal livelinessCheck
	// Perform healthcheck on the service.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//     Responses:
	//       default: emptyResponse
	//       200: emptyResponse
	//       503: errorResponse

	// Add a healthcheck endpoint
	err := healthcheck.New(s.router, config.DefaultConfig(), checks)
	if err != nil {
		s.obs.Log().With(zap.Error(err)).Panic("Cannot initialize healthcheck endpoint")
		return
	}

	s.server = &http.Server{Addr: s.config.Address, Handler: s.router}

	go func() {
		if s.config.TLS.IsEnabled {
			if err := s.server.ListenAndServeTLS(s.config.TLS.CertificatePath, s.config.TLS.PrivateKeyPath); err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.obs.Log().Panic("HTTP server failed to start", zap.Error(err))
			}
		}

		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.obs.Log().Panic("HTTP server failed to start", zap.Error(err))
		}
	}()
}

// Shutdown stops the HTTP server gracefully
func (s *Server) Shutdown() {
	s.obs.Log().Info("Shutting down the HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		s.obs.Log().Error("HTTP server shutdown failed", zap.Error(err))
	}
}
