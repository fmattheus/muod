package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fmattheus/muod/pkg/config"
	"github.com/fmattheus/muod/pkg/ping"
)

// Constants for output formatting
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	// Minimum timeout to prevent too frequent pings
	minTimeout = 100 * time.Millisecond
)

// Flag variables
var (
	debugFlag   bool
	timeoutFlag string
	plainFlag   bool
	countFlag   int
	configFlag  string
	timeout     time.Duration
)

// parseTimeout converts a string timeout value to time.Duration
func parseTimeout(t string) (time.Duration, error) {
	seconds, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid timeout value: %v", err)
	}
	if seconds <= 0 {
		return 0, fmt.Errorf("timeout must be greater than 0")
	}
	duration := time.Duration(seconds * float64(time.Second))
	if duration < minTimeout {
		return 0, fmt.Errorf("timeout must be at least %v", minTimeout)
	}
	return duration, nil
}

// debugPrint prints debug messages if debug mode is enabled
func debugPrint(format string, args ...interface{}) {
	if debugFlag {
		fmt.Fprintf(os.Stderr, "Debug: "+format+"\n", args...)
	}
}

func init() {
	// First define the config file flag so we can load the right config
	flag.StringVar(&configFlag, "config", "", "Path to config file (default: $XDG_CONFIG_HOME/muod/muod.yaml)")
	flag.StringVar(&configFlag, "f", "", "Path to config file (shorthand)")

	// Define debug flags first so we can use them for config loading
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug output")
	flag.BoolVar(&debugFlag, "d", false, "Enable debug output (shorthand)")

	// Pre-parse just the config and debug flags
	flag.Parse()
	// Set debug mode in config package
	config.Debug = debugFlag
	// Reset flag.Parsed() so we can parse again after setting up all flags
	flag.CommandLine.Init(flag.CommandLine.Name(), flag.ContinueOnError)

	// Load configuration
	cfg, err := config.LoadConfig(configFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Define remaining flags with values from config
	flag.StringVar(&timeoutFlag, "timeout", fmt.Sprintf("%.1f", cfg.DefaultTimeout.Seconds()), "Timeout in seconds (e.g., 5, 0.5)")
	flag.StringVar(&timeoutFlag, "t", fmt.Sprintf("%.1f", cfg.DefaultTimeout.Seconds()), "Timeout in seconds (shorthand)")
	
	flag.BoolVar(&plainFlag, "plain", !cfg.ShowTimestamps, "Plain output without timestamps")
	flag.BoolVar(&plainFlag, "p", !cfg.ShowTimestamps, "Plain output without timestamps (shorthand)")

	flag.IntVar(&countFlag, "count", cfg.DefaultCount, "Number of ping rounds to send (-1 for infinite, 0 to exit after DNS resolution)")
	flag.IntVar(&countFlag, "c", cfg.DefaultCount, "Number of ping rounds to send (shorthand)")
}

func monitorHosts(resolvedHosts []ping.HostInfo) {
	// If count is 0, return immediately after DNS resolution
	if countFlag == 0 {
		return
	}

	pinger, err := ping.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pinger: %v\n", err)
		os.Exit(1)
	}
	defer pinger.Close()

	start := time.Now()
	count := 0

	for {
		nextPingTime := start.Add(time.Duration(count) * timeout)
		if wait := time.Until(nextPingTime); wait > 0 {
			debugPrint("Waiting %v until next ping round", wait)
			time.Sleep(wait)
		}

		var parts []string

		// Add timestamp unless plain output is requested
		if !plainFlag {
			timestamp := time.Now().Format("15:04:05")
			parts = append(parts, timestamp)
		}

		// Ping each host
		for _, host := range resolvedHosts {
			rtt, err := pinger.Ping(host.IPAddr, timeout)
			if err != nil {
				debugPrint("[%s] Ping failed: %v", host.Hostname, err)
				parts = append(parts, fmt.Sprintf("%s%s%s", colorRed, host.Hostname, colorReset))
			} else {
				debugPrint("[%s] Ping successful, RTT: %v", host.Hostname, rtt)
				parts = append(parts, fmt.Sprintf("%s%s%s", colorGreen, host.Hostname, colorReset))
			}
		}

		// Print all hosts on one line with a newline at the end
		fmt.Printf("%s\n", strings.Join(parts, " "))

		count++
		if countFlag > 0 && count >= countFlag {
			break
		}
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] hostname1 [hostname2 ...]\n\n", "muod")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nConfiguration:\n")
		fmt.Fprintf(os.Stderr, "  MUOD can be configured using a YAML file ($XDG_CONFIG_HOME/muod/muod.yaml)\n")
		fmt.Fprintf(os.Stderr, "  Example configuration:\n")
		fmt.Fprintf(os.Stderr, "    default_timeout: 5s\n")
		fmt.Fprintf(os.Stderr, "    show_timestamps: true\n")
		fmt.Fprintf(os.Stderr, "    default_count: -1\n")
	}
	
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(configFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if cfg.DefaultTimeout < minTimeout {
		fmt.Fprintf(os.Stderr, "Warning: Config default_timeout is too low, using %v\n", minTimeout)
		cfg.DefaultTimeout = minTimeout
	}

	timeout, err = parseTimeout(timeoutFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	hosts := flag.Args()
	if len(hosts) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	debugPrint("Resolving hosts...")
	resolvedHosts, err := ping.ResolveHosts(hosts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if countFlag == 0 {
		fmt.Println("DNS resolution complete. Exiting as requested (count=0).")
		os.Exit(0)
	}

	// Build a concise status line
	status := fmt.Sprintf("Monitoring %d host", len(hosts))
	if len(hosts) > 1 {
		status += "s"
	}
	if countFlag > 0 {
		status += fmt.Sprintf(" for %d round", countFlag)
		if countFlag > 1 {
			status += "s"
		}
	}
	status += fmt.Sprintf(" (timeout: %.1fs) - Press Ctrl+C to stop", timeout.Seconds())
	fmt.Println(status)

	if debugFlag {
		fmt.Println("Debug mode enabled")
	}
	
	monitorHosts(resolvedHosts)
} 