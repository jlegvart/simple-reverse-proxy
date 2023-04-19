package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"playground/simplereverseproxy/config"
	"time"
)

var removedHeaders = []string{"Keep-Alive"}
var removedHeadersMap = map[string]struct{}{}

func main() {
	fmt.Println("Starting http reverse-proxy")

	config, err := config.ParseConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	initRemovedHeaders()

	proxy := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Println(req)

		prepareRequest(&config, req)

		response, err := http.DefaultClient.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(rw, err)
			return
		}

		addHeadersFromResponse(&rw, response)

		done := make(chan bool)

		go func() {
			for {
				select {
				case <-time.Tick(10 * time.Millisecond):
					rw.(http.Flusher).Flush()
				case <-done:
					return
				}
			}
		}()

		rw.WriteHeader(response.StatusCode)
		io.Copy(rw, response.Body)
		close(done)
	})

	if config.HttpsConfig.Enabled {
		startTLS(&config, &proxy)
	} else {
		start(&config, &proxy)
	}
}

func start(config *config.Config, proxy *http.HandlerFunc) {
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.LocalPort), proxy)

	if err != nil {
		log.Fatal(err)
	}
}

func startTLS(config *config.Config, proxy *http.HandlerFunc) {
	err := http.ListenAndServeTLS(fmt.Sprintf(":%d", config.LocalPort), config.HttpsConfig.CertPath, config.HttpsConfig.KeyPath, proxy)

	if err != nil {
		log.Fatal(err)
	}
}

func prepareRequest(config *config.Config, req *http.Request) {
	req.Host = config.ProxyUrl.Host
	req.URL.Host = config.ProxyUrl.Host
	req.URL.Scheme = config.ProxyUrl.Scheme
	req.RequestURI = ""

	reqHost, _, _ := net.SplitHostPort(req.RemoteAddr)
	req.Header.Set("X-Forwarded-For", reqHost)
}

func addHeadersFromResponse(rw *http.ResponseWriter, response *http.Response) {
	for key, v := range response.Header {
		for _, headerVal := range v {
			if headerShouldBeRemoved(key) {
				continue
			}

			(*rw).Header().Set(key, headerVal)
		}
	}
}

func headerShouldBeRemoved(headerName string) bool {
	_, found := removedHeadersMap[headerName]
	return found
}

func initRemovedHeaders() {
	for _, v := range removedHeaders {
		removedHeadersMap[v] = struct{}{}
	}
}
