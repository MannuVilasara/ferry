package cli

import (
	"encoding/json"
	"ferry/internal/daemon"
	"ferry/internal/helper"
	"ferry/internal/logger"
	"flag"
	"fmt"
	"net"
	"os"
	"os/user"
	"slices"
	"syscall"
	"text/tabwriter"
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
func HandleFlags() (bool, string) {
	checkCfg := flag.Bool("t", false, "Test if the Config is valid.")
	checkReload := flag.Bool("r", false, "Reload the config on the running server.")
	checkInfo := flag.Bool("l", false, "Show Info about the Servers Registered")
	configPath := flag.String("c", "config.yaml", "Path to the config file")
	flag.Parse()

	if *checkCfg {
		_, err := helper.LoadConfig(*configPath)
		if err != nil {
			fmt.Printf("Config is invalid: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config is valid  \n")
		return true, *configPath
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
		return true, "" 
	}

	if *checkInfo {
		sock,err := net.Dial("unix", "/tmp/ferry.sock")
		if err != nil {
			logger.Fatal("Failed to dial unix socket: %v", err)
		}
		defer sock.Close()
		
		var data struct {
			ActiveBackends int                    `json:"active_backends"`
			TotalBackends  int                    `json:"total_backends"`
			Backends       []daemon.BackendStatus `json:"backends"`
		}
		
		if err := json.NewDecoder(sock).Decode(&data); err != nil {
			logger.Fatal("Failed to decode response: %v", err)
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 1, 2, ' ', 0)
		
		fmt.Fprintln(writer, "NAME\tURL\tSTATUS\t")
		fmt.Fprintln(writer, "----\t---\t------\t")

		for _, b := range data.Backends {
			status := logger.Red + "DEAD" + logger.Reset
			if b.IsAlive {
				status = logger.Green + "ALIVE" + logger.Reset
			}
			fmt.Fprintf(writer, "%s\t%s\t%s\t\n", b.Name, b.URL, status)
		}

		fmt.Fprintf(writer, "\nTotal: %d | Active: %d\n", data.TotalBackends, data.ActiveBackends)
		writer.Flush()
		
		return true, *configPath
	}

	return false, *configPath
}
