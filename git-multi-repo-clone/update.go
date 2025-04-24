package main

import (
	"os/exec"
)

// This file contains functions to help with testing

// Make exec.Command mockable for testing
var execCommand = exec.Command