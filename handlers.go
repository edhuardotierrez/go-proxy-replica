package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

var hopHeaders = []string{
	"Connection",          // Connection
	"Proxy-Connection",    // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",          // Keep-Alive
	"Proxy-Authenticate",  // Proxy-Authenticate
	"Proxy-Authorization", // Proxy-Authorization
	"Te",                  // canonicalized version of "TE"
	"Trailer",             // not Trailers per URL above; https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",   // Transfer-Encoding
	"Upgrade",             // Upgrade
}

func replicate(cfg *configEndpoint, c *gin.Context, buf []byte) (*fasthttp.Response, error) {

	if cfg.Client == nil {
		readTimeout, _ := time.ParseDuration(cfg.Timeout)
		writeTimeout, _ := time.ParseDuration(cfg.Timeout)
		if readTimeout == 0 {
			readTimeout = ReplyTimeout
		}
		if writeTimeout == 0 {
			writeTimeout = ReplyTimeout
		}
		maxIdleConnDuration, _ := time.ParseDuration("30m")
		cfg.Client = &fasthttp.Client{
			ReadTimeout:                   readTimeout,
			WriteTimeout:                  writeTimeout,
			MaxIdleConnDuration:           maxIdleConnDuration,
			NoDefaultUserAgentHeader:      true,
			DisableHeaderNamesNormalizing: false,
			DisablePathNormalizing:        true,
			// increase DNS cache time to an hour instead of default minute
			Dial: (&fasthttp.TCPDialer{
				Concurrency:      4096,
				DNSCacheDuration: time.Hour,
			}).Dial,
		}
		log.Println("Starting new client for: ", cfg.URL)
	}

	var re = regexp.MustCompile(`^/`)
	uri := re.ReplaceAllString(c.Request.RequestURI, ``)
	url := fmt.Sprintf("%s/%s", cfg.URL, uri)

	log.Println(cfg.URL)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethodBytes([]byte(c.Request.Method))
	req.SetRequestURI(url)
	req.SetBodyRaw(buf)

	// Add Headers
	for k, v := range c.Request.Header {
		req.Header.Set(k, v[0])
	}

	// Remove headers before sending request
	for _, h := range hopHeaders {
		req.Header.Del(h)
	}

	resp := fasthttp.AcquireResponse()
	err := cfg.Client.Do(req, resp)

	if err != nil {
		log.Errorf(err.Error())
		return resp, err
	}

	// Delete headers (after response)
	for _, h := range hopHeaders {
		resp.Header.Del(h)
	}

	log.Debugf("DEBUG Response Length: %d\n", resp.Header.ContentLength())
	return resp, nil

}

func ReplyProxyHandler(c *gin.Context) {

	var bodyBytes []byte

	if c.Request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
	}

	// Routines for replicas
	for _, replica := range Config.Replicas {
		go func(re *configEndpoint) {
			resp, err := replicate(re, c, bodyBytes)
			if resp != nil {
				defer fasthttp.ReleaseResponse(resp)
			}
			if err != nil {
				log.Warnln(re.URL, err.Error())
			}
		}(replica)
	}

	//
	// Main and return << to original response
	resp, err := replicate(Config.Main, c, bodyBytes)

	if resp != nil {
		defer fasthttp.ReleaseResponse(resp)
	}

	if err != nil {
		c.Data(http.StatusBadGateway, "text/plain", []byte(err.Error()))
		return
	}

	if err != nil {
		log.Errorf("%s", err.Error())
	}

	responseBody := resp.Body()

	// Replace Request with response
	resp.Header.VisitAll(func(key, value []byte) {
		c.Request.Header.Set(string(key), string(value))
		c.Writer.Header().Set(string(key), string(value))
	})

	if responseBody != nil {
		c.Data(resp.StatusCode(), string(resp.Header.ContentType()), responseBody)
	} else {
		c.Data(resp.StatusCode(), string(resp.Header.ContentType()), responseBody)
	}

}
