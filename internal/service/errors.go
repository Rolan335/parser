package service

import "errors"

var ErrExtract = errors.New("extract metadata failed")

// Системные поля exiftool, которые не нужно отдавать клиенту
var systemFields = []string{
	"SourceFile",
	"Directory",
	"FilePermissions",
	"FileAccessDate",
	"FileModifyDate",
	"FileInodeChangeDate",
	"ExifToolVersion",
}
