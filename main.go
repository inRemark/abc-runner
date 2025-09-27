package main

import "abc-runner/app/bootstrap"

func main() {
	app := bootstrap.NewApplication()
	if err := app.Run(); err != nil {
		panic(err)
	}
}
