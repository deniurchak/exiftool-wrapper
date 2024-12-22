package server

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os/exec"
)

type Description struct {
	Lang string `xml:"lang,attr"`
	Text string `xml:",chardata"`
}

type Tag struct {
	ID          string        `xml:"id,attr"`
	Name        string        `xml:"name,attr"`
	Type        string        `xml:"type,attr"`
	Writable    string        `xml:"writable,attr"`
	Group2      string        `xml:"g2,attr"`
	Description []Description `xml:"desc"`
}

type Table struct {
	Name string `xml:"name,attr"`
	Tags []Tag  `xml:"tag"`
}

type TagInfo struct {
	Tables []Table `xml:"table"`
}

type TagResponse struct {
	Path        string            `json:"path"`
	Writable    bool              `json:"writable"`
	Type        string            `json:"type"`
	Group       string            `json:"group"`
	Description map[string]string `json:"description"`
}

func HandleTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	stdout, cmd, err := executeExifTool(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Listen for context cancellation and kill the process if canceled
	go func() {
		<-ctx.Done()
		fmt.Println("Context canceled, killing process")
		cmd.Process.Kill()
	}()

	xmlDecoder := xml.NewDecoder(stdout)
	jsonEncoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")

	// Start the JSON array
	w.Write([]byte(`{"tags":[`))

	// All table apart from the first one should be comma separated
	isFirstTable := true

	for {
		xmlToken, err := xmlDecoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			cmd.Process.Kill()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if startElement, ok := xmlToken.(xml.StartElement); ok {
			// we are processing the table tag
			if startElement.Name.Local == "table" {
				var table Table
				if err := xmlDecoder.DecodeElement(&table, &startElement); err != nil {
					cmd.Process.Kill()
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				isFirstTable = writeTable(table, jsonEncoder, w, isFirstTable)
			}
		}
	}

	// End the JSON array
	w.Write([]byte("]}"))

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != context.Canceled {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func executeExifTool(ctx context.Context) (io.ReadCloser, *exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, "exiftool", "-listx")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	return stdout, cmd, nil
}

func writeTable(table Table, encoder *json.Encoder, w http.ResponseWriter, isFirstTable bool) bool {
	for _, tag := range table.Tags {
		tagResp := TagResponse{
			Path:        table.Name + ":" + tag.Name,
			Type:        tag.Type,
			Group:       table.Name,
			Writable:    tag.Writable == "true",
			Description: make(map[string]string),
		}

		for _, desc := range tag.Description {
			tagResp.Description[desc.Lang] = desc.Text
		}

		if !isFirstTable {
			w.Write([]byte(","))
		}
		isFirstTable = false
		encoder.Encode(tagResp)
	}
	return isFirstTable
}
