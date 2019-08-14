package main

import (
	"log"

	"github.com/bscott/golangflow/actions"
	"github.com/gobuffalo/envy"
)

func main() {
	port := envy.Get("PORT", "3000")
	log.Printf("Starting golangflow on port %s\n", port)
	app := actions.App()
	log.Fatal(app.Serve())
	//log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), actions.App()))
}
