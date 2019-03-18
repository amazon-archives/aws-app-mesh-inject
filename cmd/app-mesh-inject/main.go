/*
  Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.

  Licensed under the Apache License, Version 2.0 (the "License").
  You may not use this file except in compliance with the License.
  A copy of the License is located at

      http://www.apache.org/licenses/LICENSE-2.0

  or in the "license" file accompanying this file. This file is distributed
  on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
  express or implied. See the License for the specific language governing
  permissions and limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"github.com/awslabs/aws-app-mesh-inject/pkg/config"
	"github.com/awslabs/aws-app-mesh-inject/pkg/server"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
)

var (
	dev bool
	cfg config.Config
)

func init() {
	flag.StringVar(&cfg.Name, "name", os.Getenv("APPMESH_NAME"), "AWS App Mesh name")
	flag.StringVar(&cfg.Region, "region", os.Getenv("APPMESH_REGION"), "AWS App Mesh region")
	flag.StringVar(&cfg.LogLevel, "log-level", os.Getenv("APPMESH_LOG_LEVEL"), "AWS App Mesh envoy log level")
	flag.BoolVar(&cfg.EcrSecret, "ecr-secret", false, "Inject AWS app mesh pull secrets")
	flag.StringVar(&cfg.TlsCert, "tlscert", "/etc/webhook/certs/cert.pem", "Location of TLS Cert file.")
	flag.StringVar(&cfg.TlsKey, "tlskey", "/etc/webhook/certs/key.pem", "Location of TLS key file.")
	flag.BoolVar(&dev, "dev", false, "Run in dev mode no tls.")
}

func main() {
	flag.Parse()
	log.Info(cfg)
	var s *http.Server
	var err error
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := s.Shutdown(context.Background()); err != nil {
			log.Printf("Appmeshinject server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()
	if dev {
		log.Info("Serving Appmeshinject without TLS")
		s = server.NewServerNoSSL(cfg)
		log.Fatal(s.ListenAndServe())
	} else {
		log.Info("Starting new Appmeshinject Server")
		s, err = server.NewServer(cfg)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal(s.ListenAndServeTLS("", ""))
	}
	<-idleConnsClosed
}
