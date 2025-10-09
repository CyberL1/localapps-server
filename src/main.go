package main

import (
	"localapps-server/cmd"
	"localapps-server/utils"
)

func main() {
	utils.UpdateCliConfigCache()
	cmd.Execute()
}
