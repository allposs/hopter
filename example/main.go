package main

import (
	web "github.com/allposs/hopter"
)

func main() {
	config := web.NewConfig()
	web.New(config).Attach().Run()
}
