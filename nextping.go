package main

import (
	"bufio"
	"fmt"
//	"net"
	"os"
	"os/exec"
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
		fmt.Println("Usage: go run main.go <target_url>")
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

	fmt.Println("Traceroute results:")

	// Initialize results slice
	results := make([]ProbeResult, numProbes)
	for i := 0; i < numProbes; i++ {
		results[i] = ProbeResult{
			ProbeNumber: i + 1,
			LossPercent: 0.0,
		}
	}

	// Configure table writer
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Test Run", "Probe", "Loss", "Path Information"})

	// Keep track of the test run
	testRun := 1

	for {
		for i := 0; i < numProbes; i++ {
			output, err := runTraceroute(targetHostname, maxHops, timeout)
			if err != nil {
				fmt.Printf("Error performing traceroute: %v\n", err)
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

func runTraceroute(targetHostname string, maxHops, timeout int) (string, error) {
	cmd := exec.Command("traceroute", "-I", "-m", strconv.Itoa(maxHops), "-w", strconv.Itoa(timeout), targetHostname)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("exit status 2: %s", output)
	}

	return string(output), nil
}

func calculateLossPercent(output string) float64 {
	lines := strings.Split(output, "\n")
	totalPackets := 0
	lostPackets := 0

	for _, line := range lines {
		if strings.Contains(line, "ms") {
			totalPackets++
			if strings.Contains(line, "*") {
				lostPackets++
			}
		}
	}

	if totalPackets > 0 {
		lossPercent := (float64(lostPackets) / float64(totalPackets)) * 100
		return lossPercent
	}

	return 0.0
}

func getPathInformation(output string) string {
	lines := strings.Split(output, "\n")
	pathInfo := ""

	for _, line := range lines {
		if strings.Contains(line, "ms") && strings.Contains(line, "*") {
			pathInfo += line + "\n"
		}
	}

	return pathInfo
}
