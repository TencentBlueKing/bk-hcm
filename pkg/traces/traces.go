/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

// Package traces 提供链路追踪相关的功能，用于跟踪和监控分布式系统中的请求流程
package traces

import (
	"context"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
)

// tracer is the global tracer.
var tracer = otel.Tracer(constant.TraceSystemName)

// TraceOption defines the trace option.
type TraceOption struct {
	Enabled     bool
	ServiceName string
	Token       string
	Endpoint    string
}

// InitTracer initializes the tracer provider.
func InitTracer(ctx context.Context, opt TraceOption) error {
	// create an HTTP client
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(opt.Endpoint),
		otlptracehttp.WithInsecure(),
	)

	// create an exporter
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		logs.Errorf("create trace exporter failed, err: %v", err)
		return err
	}

	// create a resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(opt.ServiceName),
			attribute.String("bk.data.token", opt.Token),
		),
	)
	if err != nil {
		logs.Errorf("create resource failed, err: %v", err)
		return err
	}

	tracerProviderOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	}
	if opt.Enabled {
		tracerProviderOpts = append(tracerProviderOpts, sdktrace.WithSampler(sdktrace.AlwaysSample()))
	} else {
		tracerProviderOpts = append(tracerProviderOpts, sdktrace.WithSampler(sdktrace.NeverSample()))
	}

	// register a tracer provider
	tracerProvider := sdktrace.NewTracerProvider(tracerProviderOpts...)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{},
		propagation.Baggage{}))

	// create a tracer
	tracer = otel.Tracer(opt.ServiceName)

	// shutdown the tracer provider when the application exits
	go func() {
		select {
		case <-ctx.Done():
			if err := tracerProvider.Shutdown(ctx); err != nil {
				logs.Errorf("tracer provider shutdown failed, err: %v", err)
			}
		}
	}()

	return nil
}

// StartReq starts a new span for a restful request.
func StartReq(req *restful.Request, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	// try to extract span context from request
	propagator := otel.GetTextMapPropagator()
	ctx := propagator.Extract(req.Request.Context(), propagation.HeaderCarrier(req.Request.Header))

	if spanContext := trace.SpanContextFromContext(ctx); spanContext.IsValid() && spanContext.IsRemote() {
		// if the request is already traced, use the existing span context
		opts = append(opts, trace.WithLinks(trace.Link{SpanContext: spanContext}))
	} else {
		// extract request ID from header
		if kt, err := kit.FromHeader(req.Request.Context(), req.Request.Header); err == nil {
			// convert request ID to trace ID
			// in order that we can search the trace by the BK request ID directly
			if requestID, err := uuid.Parse(kt.Rid); err == nil {
				traceID := trace.TraceID{}
				copy(traceID[:], requestID[:])

				// create context from specified trace ID
				ctx = trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
					TraceID:    traceID,
					SpanID:     trace.SpanID{},
					TraceFlags: trace.FlagsSampled,
				}))
			}
		}
	}

	// span kind server (callee)
	opts = append(opts, trace.WithSpanKind(trace.SpanKindServer))

	// set http route as span attribute
	if route := req.SelectedRoutePath(); route != "" {
		opts = append(opts, trace.WithAttributes(semconv.HTTPRoute(route)))
	}

	// start span
	ctx, span := tracer.Start(ctx, spanName, opts...)

	// set request context to request
	req.Request = req.Request.WithContext(ctx)

	return ctx, span
}

// StartCtx starts a new span with a context.
func StartCtx(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, name, opts...)
}

// MakeFilter returns a restful filter for tracing.
func MakeFilter(spanName string) func(*restful.Request, *restful.Response, *restful.FilterChain) {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		_, span := StartReq(req, spanName)

		chain.ProcessFilter(req, resp)

		// set span status
		if resp.Error() != nil {
			span.SetStatus(codes.Error, resp.Error().Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}
}
