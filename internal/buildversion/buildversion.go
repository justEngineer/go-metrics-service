// Package buildversion предназначен для отображения текущей версии сборки.
package buildversion

import "fmt"

const (
	BuildVersion = "NA"
	BuildDate    = "N/A"
	BuildCommit  = "N/A"
)

func Print() {
	fmt.Printf("Build version: %s\n", BuildVersion)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Build commit: %s\n", BuildCommit)
}

func Init() {
	print()
}
