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
	"net/http"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/awslabs/aws-app-mesh-inject/config"
	"github.com/awslabs/aws-app-mesh-inject/server"
)

var (
	dev bool
)

func init() {
	flag.BoolVar(&dev, "dev", false, "Run in dev mode no tls.")
}

func main() {
	flag.Parse()
	log.Info(config.DefaultConfig)
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
		s = server.NewServerNoSSL(config.DefaultConfig)
		log.Fatal(s.ListenAndServe())
	} else {
		log.Info("Starting new Appmeshinject Server")
		s, err = server.NewServer(config.DefaultConfig)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal(s.ListenAndServeTLS("", ""))
	}
	<-idleConnsClosed
}
