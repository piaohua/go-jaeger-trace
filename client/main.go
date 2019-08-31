package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptrace"

	"github.com/opentracing-contrib/go-stdlib/nethttp"

	"github.com/alextanhongpin/go-jaeger-trace/tracer"
	"github.com/opentracing/opentracing-go"
	tlog "github.com/opentracing/opentracing-go/log"
)

var t opentracing.Tracer
var closer io.Closer

// const port = ":8081"
// Traefik endpoint
const port = ":80"

func main() {
	t, closer = tracer.New("client", "localhost:5775")
	defer closer.Close()
	opentracing.SetGlobalTracer(t)

	// ctx := context.Background()
	// askGoogle(ctx)
	//runClient(t)
	runClient2(t)
}

type clientTrace struct {
	span opentracing.Span
}

func (t *clientTrace) dnsStart(info httptrace.DNSStartInfo) {
	// t.span.LogKV(
	// 	tlog.String("event", "DNS start"),
	// 	tlog.Object("host", info.Host),
	// )
	t.span.LogEvent("DNS start")
	t.span.LogFields(tlog.String("host", info.Host))

}

func (t *clientTrace) dnsDone(httptrace.DNSDoneInfo) {
	t.span.LogFields(tlog.String("event", "DNS done"))
}

func NewClientTrace(span opentracing.Span) *httptrace.ClientTrace {
	trace := &clientTrace{span: span}
	return &httptrace.ClientTrace{
		DNSStart: trace.dnsStart,
		DNSDone:  trace.dnsDone,
	}
}

func runClient(tracer opentracing.Tracer) {
	// nethttp.Transport from go-stdlib will do the tracing
	c := &http.Client{Transport: &nethttp.Transport{}}

	// Create a top-level span to represent full work of the client
	span := tracer.StartSpan("client")
	span.SetTag("hello", "client")
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)

	endpoint := fmt.Sprintf("http://localhost%s/redirect", port)
	log.Println("endpoint", endpoint)
	req, err := http.NewRequest("GET", endpoint, nil)
	// The hostname for traefik service discovery
	req.Host = "foo"
	if err != nil {
		log.Fatal(err)
	}

	req = req.WithContext(ctx)
	req, ht := nethttp.TraceRequest(tracer, req)
	defer ht.Finish()

	res, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(body))
}

func askGoogle(ctx context.Context) {
	var parentCtx opentracing.SpanContext
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan != nil {
		parentCtx = parentSpan.Context()
	}

	// Start a new span to wrap HTTP request
	span := t.StartSpan("ask google", opentracing.ChildOf(parentCtx))
	defer span.Finish()

	// Make the Span current in the context
	ctx = opentracing.ContextWithSpan(ctx, span)
	req, err := http.NewRequest("GET", "http://google.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	// Attach ClientTrace to the Context, and Context to request
	trace := NewClientTrace(span)
	ctx = httptrace.WithClientTrace(ctx, trace)
	req = req.WithContext(ctx)

	// Execute the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
}

func runClient2(tracer opentracing.Tracer) {
	// nethttp.Transport from go-stdlib will do the tracing
	c := &http.Client{Transport: &nethttp.Transport{}}

	// Create a top-level span to represent full work of the client
	span := tracer.StartSpan("client2")
	span.SetTag("client2_tag", "client2")
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)

	endpoint := fmt.Sprintf("http://localhost%s/home", port)
	log.Println("endpoint", endpoint)
	req, err := http.NewRequest("GET", endpoint, nil)
	// The hostname for traefik service discovery
	req.Host = "foo"
	if err != nil {
		log.Fatal(err)
	}

	// Inject the trace information into the HTTP Headers.
	err = span.Tracer().Inject(span.Context(),
		opentracing.TextMap,
		opentracing.HTTPHeadersCarrier(req.Header))
	if err != nil {
		log.Fatalf("%s: Couldn't inject headers (%v)", req.URL.Path, err)
	}

	req = req.WithContext(ctx)
	req, ht := nethttp.TraceRequest(tracer, req)
	defer ht.Finish()

	res, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(body))
}
