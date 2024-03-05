// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/alcomist/go-portfolio/internal/constant"
	"github.com/gin-gonic/gin"
	"net/http"
)

func HelloWorld(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"status": constant.ResultOK, "json": constant.DefaultJson})
}
