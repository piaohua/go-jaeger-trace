package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alextanhongpin/go-jaeger-trace/tracer"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
)

const port = ":8081"

func index(w http.ResponseWriter, r *http.Request) {
	endpoint := fmt.Sprintf("http://localhost%s/redirect", port)
	http.Redirect(w, r, endpoint, 301)
}

func redirect(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `{"message": "hello!"}`)
}

func main() {
	t, closer := tracer.New("server", "localhost:5775")
	defer closer.Close()
	opentracing.SetGlobalTracer(t)

	http.HandleFunc("/", index)
	http.HandleFunc("/redirect", redirect)
	http.HandleFunc("/home", serviceHandler)
	fmt.Printf("listening to port *%s. press ctrl + c to cancel", port)
	http.ListenAndServe(port, nethttp.Middleware(opentracing.GlobalTracer(), http.DefaultServeMux))
}

func serviceHandler(w http.ResponseWriter, r *http.Request) {
	opName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	var sp opentracing.Span
	spCtx, err := opentracing.GlobalTracer().Extract(
		opentracing.TextMap,
		opentracing.HTTPHeadersCarrier(r.Header))
	log.Printf("%s: extract headers (%v)", r.URL.Path, r.Header)
	if err == nil {
		sp = opentracing.StartSpan(opName, opentracing.ChildOf(spCtx))
	} else {
		sp = opentracing.StartSpan(opName)
		log.Printf("%s: Couldn't extract headers (%v)", r.URL.Path, err)
	}
	defer sp.Finish()

	sp.LogEventWithPayload("debug", fmt.Sprint("home request LogEventWithPayload opName ", opName))
	sp.LogEvent(fmt.Sprintf("home request LogEvent opName %s", opName))
	sp.Log(opentracing.LogData{
		Timestamp: time.Now(),
		Event:     fmt.Sprintf("home request Log opName %s", opName),
		Payload:   opName,
	})

	//w.Write([]byte("... done!"))
	toGoogle(w, r)
}

func toGoogle(w http.ResponseWriter, r *http.Request) {
	endpoint := fmt.Sprintf("https://www.google.com")
	http.Redirect(w, r, endpoint, 301)
}
