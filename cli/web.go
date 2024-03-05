// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/alcomist/go-portfolio/cli/controller"
	"github.com/alcomist/go-portfolio/internal/glog"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func main() {

	defer glog.Set(os.Args[0])()

	port := flag.Int("port", 6290, "Port number")
	flag.Parse()

	if *port < 0 || *port > 65535 {
		log.Fatalf("invalid port number : %d", *port)
	}

	router := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	router.StaticFile("/", "www/index.html")

	router.GET("/hello/world", controller.HelloWorld)

	addr := fmt.Sprintf(":%d", *port)
	if err := router.Run(addr); err != nil {
		log.Panic(err)
	}
}
