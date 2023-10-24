package watchchanges

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/KevinWang15/k/pkg/consts"
	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/tidwall/gjson"
)

var (
	// Store old values keyed by uid
	oldValues = make(map[string]string)

	printFormattedJson = os.Getenv(consts.K_PRINT_FORMATTED_JSON) == "true"
	noEllipsis         = os.Getenv(consts.K_NO_ELLIPSIS) == "true"
	printBodyOfAdded   = os.Getenv(consts.K_PRINT_BODY_OF_ADDED) == "true"
)

func Run() {

	scanner := bufio.NewScanner(os.Stdin)
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
	eventType := gjson.Get(line, "type").String()
	object := gjson.Get(line, "object")
	kind := object.Get("kind").String()
	metadata := object.Get("metadata")
	namespace := metadata.Get("namespace").String()
	name := metadata.Get("name").String()
	uid := metadata.Get("uid").String()

	// Remove managedFields and resourceVersion
	objectMap := object.Map()
	metadataMap := objectMap["metadata"].Map()
	delete(metadataMap, "managedFields")
	delete(metadataMap, "resourceVersion")
	objectMap["metadata"] = gjson.Parse(gjsonMapToJson(metadataMap))
	object = gjson.Parse(gjsonMapToJson(objectMap))
	currentTime := faintWhite.Sprintf(time.Now().Format(time.StampMilli) + " ")

	modified := func() {
		oldValue, ok := oldValues[uid]
		if !ok {
			fmt.Printf("Error: No old value for uid %s\n", uid)
			return
		}
		newValue := mustMarshalJson(object.Value())
		oldValues[uid] = newValue

		diffText := renderDiff(oldValue, newValue)
		if diffText != "" {
			fmt.Printf(currentTime+boldYellow.Sprintf("MODIFIED")+": %s %s/%s - %s\n", kind, namespace, name, diffText)
		}
	}

	switch eventType {
	case "ADDED":
		if oldValues[uid] != "" {
			modified()
		} else {
			oldValues[uid] = mustMarshalJson(object.Value())
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
	dmp := diffmatchpatch.New()
	if printFormattedJson {
		oldValue = mustFormatJSON(oldValue)
		newValue = mustFormatJSON(newValue)
	}
	diffs := dmp.DiffMain(oldValue, newValue, false)
	if noMeaningfulDiffs(diffs) {
		return ""
	}
	var diffText strings.Builder
	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			diffText.WriteString(color.GreenString(diff.Text))
		case diffmatchpatch.DiffDelete:
			diffText.WriteString("\033[9m" + color.RedString(diff.Text) + "\033[0m")
		case diffmatchpatch.DiffEqual:
			text := diff.Text
			if !noEllipsis && len(text) > 30 {
				text = text[:27] + "..."
			}
			diffText.WriteString(text)
		}
	}
	return diffText.String()
}

func mustMarshalJson(value interface{}) string {
	result, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(result)
}

func mustFormatJSON(inputJSON string) string {
	var data interface{}

	// Parse the input JSON into an interface{}
	if err := json.Unmarshal([]byte(inputJSON), &data); err != nil {
		panic(err)
	}

	// Marshal the interface{} back into an indented JSON string
	formattedJSON, err := json.MarshalIndent(data, "", "    ") // You can customize the indentation here
	if err != nil {
		panic(err)
	}

	return string(formattedJSON)
}

func noMeaningfulDiffs(diffs []diffmatchpatch.Diff) bool {
	for _, diff := range diffs {
		if diff.Type != diffmatchpatch.DiffEqual {
			return false
		}
	}
	return true
}

func gjsonMapToJson(objectMap map[string]gjson.Result) string {
	var builder strings.Builder
	builder.WriteString("{")

	// Sort the keys
	keys := make([]string, 0, len(objectMap))
	for key := range objectMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Iterate over the sorted keys
	for _, key := range keys {
		value := objectMap[key]
		builder.WriteString(fmt.Sprintf("\"%s\": %s,", key, value.String()))
	}
	builder.WriteString("}")
	return builder.String()
}
