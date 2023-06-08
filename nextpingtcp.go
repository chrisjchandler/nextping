package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

type ProbeResult struct {
	ProbeNumber int
	LossPercent float64
	Output      string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <target_url> [port]")
		return
	}

	targetURL := os.Args[1]
	targetHostname := extractHostname(targetURL)

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter the number of probes to send: ")
	numProbesStr, _ := reader.ReadString('\n')
	numProbes, _ := strconv.Atoi(strings.TrimSpace(numProbesStr))

	fmt.Print("Enter the maximum number of hops (TTL): ")
	maxHopsStr, _ := reader.ReadString('\n')
	maxHops, _ := strconv.Atoi(strings.TrimSpace(maxHopsStr))

	fmt.Print("Enter the timeout in milliseconds: ")
	timeoutStr, _ := reader.ReadString('\n')
	timeout, _ := strconv.Atoi(strings.TrimSpace(timeoutStr))

	port := 80
	if len(os.Args) >= 3 {
		portStr := os.Args[2]
		portInt, err := strconv.Atoi(portStr)
		if err == nil {
			port = portInt
		}
	}

	fmt.Println("Traceroute results:")

	// Initialize results slice
	results := make([]ProbeResult, numProbes)
	for i := 0; i < numProbes; i++ {
		results[i] = ProbeResult{
			ProbeNumber: i + 1,
			LossPercent: 0.0,
			Output:      "",
		}
	}

	// Configure table writer
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Test Run", "Probe", "Loss", "Path Information"})

	// Keep track of the test run
	testRun := 1

	for {
		for i := 0; i < numProbes; i++ {
			output, err := runTCPProbe(targetHostname, maxHops, timeout, port)
			if err != nil {
				fmt.Printf("Error performing TCP probe: %v\n", err)
				continue
			}

			loss := calculateLossPercent(output)
			results[i].LossPercent = loss
			results[i].Output = output

			// Clear the screen before printing the table
			fmt.Print("\033[H\033[2J")

			// Update the table
			table.ClearRows()
			for _, result := range results {
				row := []string{
					strconv.Itoa(testRun),
					strconv.Itoa(result.ProbeNumber),
					fmt.Sprintf("%.2f%%", result.LossPercent),
					getPathInformation(result.Output),
				}
				table.Append(row)
			}
			table.Render()

			// Increment the test run
			testRun++
		}

		time.Sleep(1 * time.Second) // Sleep for 1 second before running the next test
	}
}

func extractHostname(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		parts := strings.SplitN(url, "/", 3)
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	return url
}

func runTCPProbe(targetHostname string, maxHops, timeout, port int) (string, error) {
	destAddr, err := net.ResolveIPAddr("ip", targetHostname)
	if err != nil {
		return "", err
	}

	for ttl := 1; ttl <= maxHops; ttl++ {
		deadline := time.Now().Add(time.Duration(timeout) * time.Millisecond)
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", destAddr.String(), port), time.Duration(timeout)*time.Millisecond)
		if err != nil {
			return "", err
		}
		defer conn.Close()

		err = conn.SetDeadline(deadline)
		if err != nil {
			return "", err
		}

		ip := conn.LocalAddr().(*net.TCPAddr).IP.String()
		remoteIP := conn.RemoteAddr().(*net.TCPAddr).IP.String()

		probeOutput := fmt.Sprintf("TTL: %d, Source: %s, Destination: %s\n", ttl, ip, remoteIP)
		return probeOutput, nil
	}

	return "", fmt.Errorf("maximum number of hops reached")
}

func calculateLossPercent(output string) float64 {
	// Custom logic to calculate loss percentage based on the output
	// Modify as per your requirements
	return 0.0
}

func getPathInformation(output string) string {
	// Custom logic to extract path information from the output
	// Modify as per your requirements
	return ""
}
