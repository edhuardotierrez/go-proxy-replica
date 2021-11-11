package main

import (
	"flag"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	"os"
	"strings"
)

// RunWithManager support custom autocert manager
func RunWithManager(r http.Handler, m *autocert.Manager) error {
	s := &http.Server{
		Addr:      ":https",
		TLSConfig: m.TLSConfig(),
		Handler:   r,
	}

	go func() {
		err := http.ListenAndServe(Config.Server.BindAddress, m.HTTPHandler(http.HandlerFunc(redirect)))
		if err != nil {
			log.Errorf(err.Error())
		}
	}()

	return s.ListenAndServeTLS("", "")
}

func redirect(w http.ResponseWriter, req *http.Request) {
	if Config.Server.AutoTLS.Redirect {
		target := "https://" + req.Host + req.RequestURI
		http.Redirect(w, req, target, http.StatusMovedPermanently)
	}
}

// -------------------------------------------------------------------------------- Main Functions

func main() {

	//
	configFlag := flag.String("config", "", "config path (ex: -config ./proxies.yaml")
	versionFlag := flag.Bool("version", false, "get version")
	logLevel := flag.String("level", "", "set log level")

	// auto config
	domainsFlag := flag.String("domains", "", "(optional) lets encrypt domains")
	emailFlag := flag.String("email", "", "(optional) lets encrypt email")
	masterFlag := flag.String("master", "", "(optional) master url")

	flag.Parse()

	log.Infoln("go-proxy-replica " + Version)

	if *versionFlag {
		os.Exit(0)
	}

	if len(strings.TrimSpace(*logLevel)) > 2 {
		LogLevel = strings.TrimSpace(*logLevel)
	}

	if len(strings.TrimSpace(*configFlag)) > 4 {
		ConfigFilePath = *configFlag
	}

	if len(strings.TrimSpace(*logLevel)) > 2 {
		LogLevel = strings.TrimSpace(*logLevel)
	}

	// Config Init
	LoadConfig()

	// Replace Vars
	if len(strings.TrimSpace(*domainsFlag)) > 2 {
		log.Warningf(">> Runtime changing variable: domains (%s)", *domainsFlag)
		var domains []string
		for _, v := range strings.Split(*domainsFlag, ",") {
			domain := strings.TrimSpace(v)
			if strings.Contains(domain, ".") {
				domains = append(domains, domain)
			}
		}

		if len(domains) > 0 {
			Config.Server.AutoTLS.Enabled = true
			Config.Server.AutoTLS.Domains = domains
		}

	}

	if len(strings.TrimSpace(*emailFlag)) > 2 {
		log.Warningf(">> Runtime changing variable: email (%s)", *emailFlag)
		Config.Server.AutoTLS.Email = *emailFlag
	}

	if len(strings.TrimSpace(*masterFlag)) > 2 {
		log.Warningf(">> Runtime changing variable: master (%s)", *masterFlag)
		Config.Master.URL = *masterFlag
	}

	// Route
	route := gin.Default()

	// Allow all origins (from public domains)
	route.Use(cors.Default())

	// Add all routers
	route.NoRoute(ReplyProxyHandler)

	if Config.Server.AutoTLS.Enabled {
		//
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(Config.Server.AutoTLS.Domains...),
			Cache:      autocert.DirCache("/tmp/.cache"),
			Email:      Config.Server.AutoTLS.Email,
		}
		if err := RunWithManager(route, &m); err != nil {
			log.Errorf("Failed to run server: %v", err)
		}
	} else {

		if err := route.Run(Config.Server.BindAddress); err != nil {
			log.Errorf("Failed to run server: %v", err)
		}

	}

}
