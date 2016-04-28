package server

import "html/template"

const (
	info = `
	INFO
`
	list = `
	LIST
`
)

var (
	infoTemplate = template.Must(template.New("info").Parse(info))
	listTemplate = template.Must(template.New("list").Parse(list))
)
