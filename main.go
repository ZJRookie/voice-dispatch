package main

import (
	"voice-dispatch/config"
	"voice-dispatch/infra"
	"voice-dispatch/route"
)

func init() {
	config.Init()
	infra.Init()
	route.Init()

}

func main() {

}
