# Tracing

## Span naming convention

`(type).package.FunctionName`

Example:

`service.chargePoint.GetChargePoint`

## Adding span to a function

Spans should be added to each layer of the application and any internal component, that calls external services, APIs,
etc.

### Span only

```go
package example

func Span(ctx context.Context, strParam string) {
	ctx, end := observability.Span(
		// Pass context from params and provide span name
		ctx, "example.Span",

		// Add params to span
		zap.String("strParam", strParam),
	)
	// Defer the call of end span function
	defer end()

	/*
	   Function content
	*/
}
```

### Span with logger

```go
package example

func SpanWithLogger(ctx context.Context, strParam string) {
	ctx, end, logger := observability.LogSpan(
		// Pass context from params and provide span name
		ctx, "example.SpanWithLogger",

		// Add params to span
		zap.String("strParam", strParam),
	)
	// Defer the call of end span function
	defer end()

	// Params passed in LogSpan function are automatically logged
	logger.Debug("Calling function")

	logger.Error("An error occurred", zap.Error(err))

	/*
	   Function content
	*/
}
```
