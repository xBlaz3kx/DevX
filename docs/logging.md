# Logging

We use [zap](https://github.com/uber-go/zap) for our logger. It is a fast, structured, leveled logging in Go.

The logger is configured in the `observability` package and is a global logger. The default log
level is `info` and can be configured via the `LOG_LEVEL` environment variable.

Logs should be structured and contain as much information as possible without being too verbose. Use the correct log
level for the message.

```go
logger := observability.Log().With(
zap.String("param1", param1),
// ...
)

logger.Debug("Debug logging")

logger.Error("An error occurred", zap.Error(err))
```

## Logging with context

```go
logger := observability.Log().Ctx(ctx)
```

## Logging with traces

```go
ctx, end, logger := observability.LogSpan(ctx, "span.newspan")
defer end()
```
