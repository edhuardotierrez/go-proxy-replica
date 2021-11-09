package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

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
	// Master and return << to original response
	resp, err := replicate(Config.Master, c, bodyBytes)

	if err != nil {
		log.Errorln(err.Error())
	}

	if resp.StatusCode > 302 {
		c.AbortWithStatus(resp.StatusCode)
		return
	}
	if resp.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(resp.Body)
		c.Data(resp.StatusCode, resp.Header.Get("Content-type"), bodyBytes)
	} else {
		c.Data(resp.StatusCode, resp.Header.Get("Content-type"), bodyBytes)
	}

}
