package main

import (
	"netdisk_in_go/router"
	"netdisk_in_go/utils"
)

func main() {
	utils.InitMySQL()
	r := router.Router()
	err := r.Run(":8080")
	if err != nil {
		panic("system run on 8080 failed")
	}
}
