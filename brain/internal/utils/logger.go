package utils

import (
	"log"
	"os"
)

var Logger = log.New(os.Stdout, "[orchestrator] ", log.LstdFlags|log.Lshortfile)
