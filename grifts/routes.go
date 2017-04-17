package grifts

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/bscott/golangflow/actions"
	. "github.com/markbates/grift/grift"
)

var _ = Add("routes", func(c *Context) error {
	a := actions.App()
	routes := a.Routes()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "METHOD\t PATH\t HANDLER")
	fmt.Fprintln(w, "------\t ----\t -------")
	for _, r := range routes {
		fmt.Fprintf(w, "%s\t %s\t %s\n", r.Method, r.Path, r.HandlerName)
	}
	w.Flush()
	return nil
})
