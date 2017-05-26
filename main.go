package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bscott/golangflow/actions"
	"github.com/gobuffalo/envy"
	"github.com/as27/setenv"
)

func init() {
	setenv.File(".setenv")
}


func main() {
	port := envy.Get("PORT", "3000")
	log.Printf("Starting golangflow on port %s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), actions.App()))
}
