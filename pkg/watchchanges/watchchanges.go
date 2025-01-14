package watchchanges

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/KevinWang15/k/pkg/consts"
	"github.com/fatih/color"
	"github.com/pmezard/go-difflib/difflib"
)

var (
	// Store old values keyed by uid
	oldValues = make(map[string]string)

	printBodyOfAdded = os.Getenv(consts.K_PRINT_BODY_OF_ADDED) == "true"

	contextLines = (func() int {
		contextLinesEnv := os.Getenv(consts.K_DIFF_CONTEXT_LINES)
		if contextLinesEnv != "" {
			v, err := strconv.Atoi(contextLinesEnv)
			if err != nil {
				panic(fmt.Errorf("failed to parse %s: %w", consts.K_DIFF_CONTEXT_LINES, err))
			}
			return v
		} else {
			return 3
		}
	})()
)

func Run() {

	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 10*64*1024)
	scanner.Buffer(buf, 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		processLine(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

var (
	faintWhite = color.New(color.FgWhite).Add(color.Faint)
	boldGreen  = color.New(color.FgGreen).Add(color.Bold)
	boldYellow = color.New(color.FgYellow).Add(color.Bold)
	boldRed    = color.New(color.FgRed).Add(color.Bold)
)

func processLine(line string) {
	parsedLine := map[string]interface{}{}
	err := json.Unmarshal([]byte(line), &parsedLine)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal line: %w", err))
	}

	eventType := parsedLine["type"].(string)
	object := parsedLine["object"].(map[string]interface{})

	kind := object["kind"].(string)
	if strings.HasSuffix(kind, "List") && object["items"] != nil {
		for _, item := range object["items"].([]interface{}) {
			processObject(item.(map[string]interface{}), eventType)
		}
	} else {
		processObject(object, eventType)
	}
}

func processObject(object map[string]interface{}, eventType string) {
	kind := object["kind"].(string)
	metadata := object["metadata"].(map[string]interface{})

	// Get values with nil checks
	var namespace, name, uid string

	// Handle namespace - it might be nil for cluster-scoped resources
	if ns, exists := metadata["namespace"]; exists && ns != nil {
		namespace = ns.(string)
	} else {
		namespace = "<no-namespace>"
	}

	// Handle name - required field in Kubernetes
	if n, exists := metadata["name"]; exists && n != nil {
		name = n.(string)
	} else {
		name = "<no-name>"
	}

	// Handle UID - required field in Kubernetes
	if u, exists := metadata["uid"]; exists && u != nil {
		uid = u.(string)
	} else {
		uid = fmt.Sprintf("%s/%s/%s", kind, namespace, name)
	}

	// Remove managedFields and resourceVersion
	delete(metadata, "managedFields")
	delete(metadata, "resourceVersion")
	currentTime := faintWhite.Sprintf(time.Now().Format(time.StampMilli) + " ")

	modified := func() {
		oldValue, ok := oldValues[uid]
		if !ok {
			fmt.Printf("Error: No old value for uid %s\n", uid)
			return
		}
		newValue := mustMarshalJson(object)
		oldValues[uid] = newValue

		diffText := renderDiff(oldValue, newValue)
		if diffText != "" {
			fmt.Printf(currentTime+boldYellow.Sprintf("MODIFIED")+": %s %s/%s\n%s\n", kind, namespace, name, diffText)
		}
	}

	switch eventType {
	case "ADDED":
		if oldValues[uid] != "" {
			modified()
		} else {
			oldValues[uid] = mustMarshalJson(object)
			if printBodyOfAdded {
				fmt.Printf(currentTime+boldGreen.Sprintf("ADDED")+": %s %s/%s - %s\n", kind, namespace, name, color.GreenString(oldValues[uid]))
			} else {
				fmt.Printf(currentTime+boldGreen.Sprintf("ADDED")+": %s %s/%s\n", kind, namespace, name)
			}
		}
	case "MODIFIED":
		modified()
	case "DELETED":
		fmt.Printf(currentTime+boldRed.Sprintf("DELETED")+": %s %s/%s\n", kind, namespace, name)
	default:
		fmt.Printf(currentTime+"Unknown event type: %s\n", eventType)
	}
}

func renderDiff(oldValue string, newValue string) string {
	oldLines := difflib.SplitLines(oldValue)
	newLines := difflib.SplitLines(newValue)

	context := contextLines
	if context == -1 {
		if len(oldLines) > len(newLines) {
			context = len(oldLines)
		} else {
			context = len(newLines)
		}
	}
	diffString, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:       oldLines,
		B:       newLines,
		Context: context,
	})
	if err != nil {
		panic(err)
	}

	return colorizeDiff(diffString)
}

func mustMarshalJson(value interface{}) string {
	result, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(result)
}

func colorizeDiff(diffString string) string {
	var colorizedDiff strings.Builder
	for _, line := range strings.Split(diffString, "\n") {
		switch {
		case strings.HasPrefix(line, "+"):
			colorizedDiff.WriteString("\033[32m" + line + "\033[0m\n") // Green for additions
		case strings.HasPrefix(line, "-"):
			colorizedDiff.WriteString("\033[31m" + line + "\033[0m\n") // Red for deletions
		default:
			colorizedDiff.WriteString(line + "\n")
		}
	}
	return colorizedDiff.String()
}
