package handler

import "io"

type HTMLRenderer interface {
	ExecuteTemplate(w io.Writer, name string, data interface{}) error
}
