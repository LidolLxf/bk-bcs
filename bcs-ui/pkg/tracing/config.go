/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tracing

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentationName = "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Config represents the configuration options available for the http.Handler
// and http.Transport types.
type Config struct {
	Tracer            trace.Tracer
	Meter             metric.Meter
	Propagators       propagation.TextMapPropagator
	SpanStartOptions  []trace.SpanStartOption
	ReadEvent         bool
	WriteEvent        bool
	Filters           []Filter
	SpanNameFormatter func(string, *http.Request) string

	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
}

// Option interface used for setting optional config properties.
type Option interface {
	apply(*Config)
}

type optionFunc func(*Config)

func (o optionFunc) apply(c *Config) {
	o(c)
}

// NewConfig creates a new config struct and applies opts to it.
func NewConfig(opts ...Option) *Config {
	c := &Config{
		Propagators:    otel.GetTextMapPropagator(),
		MeterProvider:  otel.GetMeterProvider(),
		Filters:        []Filter{},
		TracerProvider: otel.GetTracerProvider(),
	}
	for _, opt := range opts {
		opt.apply(c)
	}

	c.Meter = c.MeterProvider.Meter(
		instrumentationName,
	)
	return c
}

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return optionFunc(func(cfg *Config) {
		if provider != nil {
			cfg.TracerProvider = provider
		}
	})
}

// WithMeterProvider specifies a meter provider to use for creating a meter.
// If none is specified, the global provider is used.
func WithMeterProvider(provider metric.MeterProvider) Option {
	return optionFunc(func(cfg *Config) {
		if provider != nil {
			cfg.MeterProvider = provider
		}
	})
}

// WithPublicEndpoint configures the Handler to link the span with an incoming
// span context. If this option is not provided, then the association is a child
// association instead of a link.
func WithPublicEndpoint() Option {
	return optionFunc(func(c *Config) {
		c.SpanStartOptions = append(c.SpanStartOptions, trace.WithNewRoot())
	})
}

// WithPropagators configures specific propagators. If this
// option isn't specified, then the global TextMapPropagator is used.
func WithPropagators(ps propagation.TextMapPropagator) Option {
	return optionFunc(func(c *Config) {
		if ps != nil {
			c.Propagators = ps
		}
	})
}

// WithSpanOptions configures an additional set of
// trace.SpanOptions, which are applied to each new span.
func WithSpanOptions(opts ...trace.SpanStartOption) Option {
	return optionFunc(func(c *Config) {
		c.SpanStartOptions = append(c.SpanStartOptions, opts...)
	})
}

// WithFilter adds a filter to the list of filters used by the handler.
// If any filter indicates to exclude a request then the request will not be
// traced. All filters must allow a request to be traced for a Span to be created.
// If no filters are provided then all requests are traced.
// Filters will be invoked for each processed request, it is advised to make them
// simple and fast.
func WithFilter(f Filter) Option {
	return optionFunc(func(c *Config) {
		c.Filters = append(c.Filters, f)
	})
}

type event int

// Different types of events that can be recorded, see WithMessageEvents
const (
	ReadEvents event = iota
	WriteEvents
)

// WithMessageEvents configures the Handler to record the specified events
// (span.AddEvent) on spans. By default only summary attributes are added at the
// end of the request.
//
// Valid events are:
//   - ReadEvents: Record the number of bytes read after every http.Request.Body.Read
//     using the ReadBytesKey
//   - WriteEvents: Record the number of bytes written after every http.ResponeWriter.Write
//     using the WriteBytesKey
func WithMessageEvents(events ...event) Option {
	return optionFunc(func(c *Config) {
		for _, e := range events {
			switch e {
			case ReadEvents:
				c.ReadEvent = true
			case WriteEvents:
				c.WriteEvent = true
			}
		}
	})
}

// WithSpanNameFormatter takes a function that will be called on every
// request and the returned string will become the Span Name
func WithSpanNameFormatter(f func(operation string, r *http.Request) string) Option {
	return optionFunc(func(c *Config) {
		c.SpanNameFormatter = f
	})
}
