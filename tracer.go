package goredis

import (
	"golang.org/x/net/context"
	"strings"

	"github.com/go-redis/redis"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func WrapRedisClient(ctx context.Context, client *redis.Client) *redis.Client {
	if ctx == nil {
		return client
	}

	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan == nil {
		return client
	}

	ctxClient := client.WithContext(ctx)
	opts := ctxClient.Options()
	ctxClient.WrapProcess(process(parentSpan, opts))
	ctxClient.WrapProcessPipeline(processPipeline(parentSpan, opts))

	return ctxClient
}

func process(parentSpan opentracing.Span, opts *redis.Options) func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	return func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {

			dbMethod := formatCommandAsDbMethod(cmd)
			span := getSpan(parentSpan, opts, "redis", dbMethod)
			defer span.Finish()

			return oldProcess(cmd)
		}
	}
}

func processPipeline(parentSpan opentracing.Span, opts *redis.Options) func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
	return func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
		return func(cmds []redis.Cmder) error {
			dbMethod := formatCommandsAsDbMethods(cmds)
			span := getSpan(parentSpan, opts, "redis-pipeline", dbMethod)
			defer span.Finish()

			return oldProcess(cmds)
		}
	}
}

func formatCommandAsDbMethod(cmd redis.Cmder) string {
	return cmd.Name()
}

func formatCommandsAsDbMethods(cmds []redis.Cmder) string {
	cmdsAsDbMethods := make([]string, len(cmds))
	for i, cmd := range cmds {
		dbMethod := formatCommandAsDbMethod(cmd)
		cmdsAsDbMethods[i] = dbMethod
	}

	return strings.Join(cmdsAsDbMethods, " -> ")
}

func getSpan(parentSpan opentracing.Span, opts *redis.Options, operationName, dbMethod string) opentracing.Span {
	tracerClient := parentSpan.Tracer()
	span := tracerClient.StartSpan(operationName, opentracing.ChildOf(parentSpan.Context()))

	ext.DBType.Set(span, "redis")
	ext.PeerAddress.Set(span, opts.Addr)
	ext.SpanKind.Set(span, "client")

	span.SetTag("db.method", dbMethod)
	return span
}
