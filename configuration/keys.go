package configuration

const (
	EtcdAddress = "etcd.address"
	EtcdPrefix  = "etcd.prefix"

	// Database
	DatabaseType        = "database.type"
	DatabaseAddress     = "database.host"
	DatabasePort        = "database.port"
	DatabaseUsername    = "database.username"
	DatabasePassword    = "database.password"
	DatabaseName        = "database.database"
	DatabaseReplicaSet  = "database.replicaSet"
	DatabaseTLSEnabled  = "database.tls.enabled"
	DatabaseTLSKeyPath  = "database.tls.keyPath"
	DatabaseTLSCertPath = "database.tls.certPath"

	// HTTP server
	ServerAddress     = "http.address"
	ServerPathPrefix  = "http.pathPrefix"
	ServerTLSEnabled  = "http.tls.enabled"
	ServerTLSKeyPath  = "http.tls.keyPath"
	ServerTLSCertPath = "http.tls.certPath"

	// Observability

	// Metrics
	MetricsServerAddress = "observability.metrics.address"
	MetricsEnable        = "observability.metrics.enable"
	MetricsEndpoint      = "observability.metrics.endpoint"

	// Tracing
	TracingEnable      = "observability.tracing.enable"
	TracingAddress     = "observability.tracing.address"
	TracingAuth        = "observability.tracing.auth"
	TracingTLSEnabled  = "observability.tracing.tls.enabled"
	TracingTLSKeyPath  = "observability.tracing.tls.keyPath"
	TracingTLSCertPath = "observability.tracing.tls.certPath"
)
