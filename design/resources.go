package design

import (
	d "github.com/goadesign/goa/design"
	a "github.com/goadesign/goa/design/apidsl"
)

var _ = a.Resource("status", func() {

	a.DefaultMedia(ALMStatus)
	a.BasePath("/status")

	a.Action("show", func() {
		a.Routing(
			a.GET(""),
		)
		a.Description("Show the status of the current running instance")
		a.Response(d.OK)
		a.Response(d.ServiceUnavailable, ALMStatus)
	})
})

var nameValidationFunction = func() {
	a.MaxLength(62) // maximum name length is 62 characters
	a.MinLength(1)  // minimum name length is 1 characters
	a.Pattern("^[^_|-].*")
	a.Example("name for the object")
}
