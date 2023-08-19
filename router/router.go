package router

import (
	"database/sql"
	"embed"
)

// NOTE: all:template は template ディレクトリ以下の全てのファイルを埋め込む
// all なしだと _ から始まるファイルは埋め込まれない

//go:embed all:template
var templates embed.FS

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}
