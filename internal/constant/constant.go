// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package constant

const (
	EnvKeyTOMLFile      = "TOML"
	EnvKeyConfigFileDir = "CONFIG_DIR"

	// CKDBMain Database config keys
	CKDBMain = "main_db"

	// CKECMain ElasticCluster config keys
	CKECMain = "es"

	// SlackChannelDefault slack channels
	SlackChannelDefault = "default"

	ElasticSortOrderAsc  = "asc"
	ElasticSortOrderDesc = "desc"

	ElasticKeyExists = "_exists_"

	NgramMaxTokenSize    = 100
	NgramMaxStringLength = 10000

	TimeFormat     = "20060102"
	TimeFormatDash = "2006-01-02"

	LocaleEnglish = "en_us"
	LocaleKorean  = "ko_kr"

	DefaultJson = "[]"

	QueryTypeInsert = "insert"
	QueryTypeDelete = "delete"
	QueryTypeUpdate = "update"
	QueryTypeCreate = "create"
	QueryTypeSelect = "select"

	EQ  = "="
	NEQ = "<>"
	IN  = "IN"

	DBOrderAsc  = "ASC"
	DBOrderDesc = "DESC"

	TopMostNodePID = 9999999999999

	TargetPrimary     = "tp"
	TargetSecondary   = "ts"
	TargetPrimaryOnly = "tpo"
	TargetMainIf      = "tmi"

	ResultOK = "OK"
)
