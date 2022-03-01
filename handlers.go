package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"regexp"
	"time"
)

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
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
		maxIdleConnDuration, _ := time.ParseDuration("1h")
		cfg.Client = &fasthttp.Client{
			ReadTimeout:                   readTimeout,
			WriteTimeout:                  writeTimeout,
			MaxIdleConnDuration:           maxIdleConnDuration,
			NoDefaultUserAgentHeader:      true,
			DisableHeaderNamesNormalizing: false,
			DisablePathNormalizing:        true,
			// increase DNS cache time to an hour instead of default minute
			//Dial: (&fasthttp.TCPDialer{
			//	Concurrency:      4096,
			//	DNSCacheDuration: time.Hour,
			//}).Dial,
		}
		log.Println("Starting new client for: ", cfg.URL)
	}

	var re = regexp.MustCompile(`^/`)
	uri := re.ReplaceAllString(c.Request.RequestURI, ``)
	url := fmt.Sprintf("%s/%s", cfg.URL, uri)

	log.Println(cfg.URL)

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.SetBodyRaw(buf)

	// Add Headers
	for k, v := range c.Request.Header {
		req.Header.Set(k, v[0])
	}

	resp := fasthttp.AcquireResponse()
	err := cfg.Client.Do(req, resp)

	fasthttp.ReleaseRequest(req)

	if err == nil {
		log.Debugf("DEBUG Response: %s\n", resp.Body())
		return resp, nil
	} else {
		log.Errorf(err.Error())
		return nil, err
	}

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
			defer fasthttp.ReleaseResponse(resp)
			if err != nil {
				log.Warnln(re.URL, err.Error())
			}
		}(replica)
	}

	//
	// Main and return << to original response
	resp, err := replicate(Config.Main, c, bodyBytes)
	defer fasthttp.ReleaseResponse(resp)

	if err != nil {
		log.Errorf("%s", err.Error())
	}

	responseBody := resp.Body()

	// Replace Request with response
	resp.Header.VisitAll(func(key, value []byte) {
		c.Writer.Header().Set(string(key), string(value))
	})

	if responseBody != nil {
		c.Data(resp.StatusCode(), string(resp.Header.ContentType()), responseBody)
	} else {
		c.Data(resp.StatusCode(), string(resp.Header.ContentType()), responseBody)
	}

}
