package main

import "github.com/sonaak/vokun/app"

func main() {
	// create a new App
	app, err := app.Setup()
	if err != nil {
		panic(err)
	}

	// run it forever
	panic(app.Run())
}