package common

import "github.com/unrolled/render"

var jRender = render.New(render.Options{
	Charset:       "UTF-8",                                          // Sets encoding for json and html content-types. Default is "UTF-8".
	IndentJSON:    true,                                             // Output human readable JSON.
	IndentXML:     true,                                             // Output human readable XML.
	PrefixJSON:    []byte(""),                                       // Prefixes JSON responses with the given bytes.
	PrefixXML:     []byte("<?xml version='1.0' encoding='UTF-8'?>"), // Prefixes XML responses with the given bytes.
	StreamingJSON: true,
})

type QueryResult struct {
	Total int64       `json:"total"`
	Data  interface{} `json:"data"`
}

func Render() *render.Render {
	return jRender
}
