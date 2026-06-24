package cli

import (
	"slices"
	"ferry/internal/helper"
	"ferry/internal/logger"
	"flag"
	"fmt"
	"os"
	"os/user"
	"syscall"
)

// CheckPermissions ensures the current user is in the 'ferry' group.
// It exits the program if they are not.
func CheckPermissions() {
	currentUser, err := user.Current()
	if err != nil {
		logger.Fatal("Failed to get current user: %v", err)
	}

	ferryGroup, err := user.LookupGroup("ferry")
	if err != nil {
		logger.Fatal("Security Error: The 'ferry' group does not exist on this system.")
	}

	groupIDs, err := currentUser.GroupIds()
	if err != nil {
		logger.Fatal("Failed to get user groups: %v", err)
	}

	isAuthorized := slices.Contains(groupIDs, ferryGroup.Gid)

	if !isAuthorized {
		fmt.Println("Permission Denied: You must be in the 'ferry' group to run this command.")
		os.Exit(1)
	}
}

// HandleFlags parses CLI arguments. It returns true if a command was executed
// and the main program should exit, or false if the server should start normally.
func HandleFlags() bool {
	checkCfg := flag.Bool("c", false, "Check if the Config is valid.")
	checkReload := flag.Bool("r", false, "Reload the config on the running server.")
	flag.Parse()

	if *checkCfg {
		_, err := helper.LoadConfig("config.yaml")
		if err != nil {
			fmt.Printf("Config is invalid: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config is valid  \n")
		return true 
	}

	if *checkReload {
		pidBytes, err := os.ReadFile("/tmp/ferry.pid")
		if err != nil {
			fmt.Printf("Ferry doesn't Seem to be running. Couldn't find /tmp/ferry.pid\n")
			os.Exit(1)
		}

		var pid int
		fmt.Sscanf(string(pidBytes), "%d", &pid)

		ps, err := os.FindProcess(pid)
		if err != nil {
			logger.Fatal("Ferry doesn't seem to be running, send SIGHUP failed: %v", err)
		}

		err = ps.Signal(syscall.SIGHUP)
		if err != nil {
			logger.Fatal("Failed to send signal: %v", err)
		}

		fmt.Printf("Config reloaded successfully  \n")
		return true 
	}

	return false
}
