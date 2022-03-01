package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
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

func replicate(cfg configEndpoint, c *gin.Context, buf []byte) (*http.Response, error) {

	//
	body := ioutil.NopCloser(bytes.NewBuffer(buf))

	var re = regexp.MustCompile(`^/`)
	uri := re.ReplaceAllString(c.Request.RequestURI, ``)
	url := fmt.Sprintf("%s/%s", cfg.URL, uri)

	log.Println(cfg.URL)

	//
	// req.Header.Add("X-Forwarded-Host", host)
	var timeout = ReplyTimeout

	t, err := time.ParseDuration(cfg.Timeout)
	if err != nil || cfg.Timeout == "" {
		timeout = t
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !cfg.VerifySSL},
	}
	client := &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	req, err := http.NewRequest(c.Request.Method, url, body)
	if err != nil {
		log.Errorln(err)
	}

	// Copy headers to >>
	req.Header = c.Request.Header
	resp, err := client.Do(req)

	return resp, err

}

func ReplyProxyHandler(c *gin.Context) {
	var bodyBytes []byte

	if c.Request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
	}

	// Routines for replicas
	go func() {
		for _, replica := range Config.Replicas {
			_, err := replicate(replica, c, bodyBytes)
			if err != nil {
				log.Warnln(replica.URL, err.Error())
			}
		}
	}()

	//
	// Main and return << to original response
	resp, err := replicate(Config.Main, c, bodyBytes)

	if err != nil {
		log.Errorf("%s", err.Error())
	}

	// Replace Request with response
	c.Request.Header = resp.Header
	c.Request.Header.Add("X-Go-Proxy-Replica", Version)
	if resp.StatusCode > 302 {
		c.String(resp.StatusCode, "%s", resp.Status)
		return
	}

	// Add Headers to Response
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}

	if resp.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(resp.Body)
		c.Data(resp.StatusCode, resp.Header.Get("Content-type"), bodyBytes)
	} else {
		c.Data(resp.StatusCode, resp.Header.Get("Content-type"), bodyBytes)
	}

}
