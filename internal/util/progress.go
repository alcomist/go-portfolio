// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"bytes"
	"fmt"
)

type ProgressInfo struct {
	nodeCount          int
	nodeProcessedCount int
	dataCount          int
	dataProcessedCount int
}

func NewProgressInfo() *ProgressInfo {
	return &ProgressInfo{}
}

func (p *ProgressInfo) SetNodeCount(c int) {
	p.nodeCount = c
}

func (p *ProgressInfo) SetDataCount(c int) {
	p.dataCount = c
}

func (p *ProgressInfo) AddNodeProcessedCount(c int) {
	p.nodeProcessedCount += c
}

func (p *ProgressInfo) AddDataProcessedCount(c int) {
	p.dataProcessedCount += c
}

func (p *ProgressInfo) String() string {

	var b bytes.Buffer

	// node % (c/t), data % (c/t)
	nodeRate := Progress(p.nodeProcessedCount, p.nodeCount)
	dataRate := Progress(p.dataProcessedCount, p.dataCount)

	fmt.Fprintf(&b, "NODE : %v (%d / %d) / DATA : %v (%d / %d)",
		nodeRate, p.nodeProcessedCount, p.nodeCount,
		dataRate, p.dataProcessedCount, p.dataCount)

	return b.String()
}
