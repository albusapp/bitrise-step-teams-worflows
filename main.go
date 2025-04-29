package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/log"
)

type Request struct {
	Type 		string 						`json:"type"`
	Attachments	[]Attachment 				`json:"attachments"`
}

type Attachment struct {
	ContentType	string						`json:"contentType,omitempty"`
	Content		Content 					`json:"content"`
}

type Content struct {
	Type 			string 					`json:"type"`
	Body			[]map[string]string 	`json:"body"`
}

func main() {
	log.SetEnableDebugLog(true)
	webhookUrl := os.Getenv("webhook_url")
	fields := parsesFields(os.Getenv("fields"))

	if err := post(webhookUrl, fields); err != nil {
		log.Errorf("Error: %s", err)
		os.Exit(1)
	}

	log.Donef("Workflow successfully triggered! 🚀")
}

func post(url string, fields map[string]string) error {
	b, err := json.Marshal(Request{
		Type:			"message",
		Attachments: 	[]Attachment{makeAttachment(fields)},
	})
	if err != nil {
		return err
	}

	log.Debugf("Url: %s\n", url)
	log.Debugf("Post Json Data: %s\n", b)

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send the request: %s", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("server error: %s, failed to read response: %s", resp.Status, err)
		}
		return fmt.Errorf("server error: %s, response: %s", resp.Status, body)
	}

	return nil
}

func makeAttachment(fields map[string]string) Attachment {
	return Attachment{
		ContentType: "application/json",
		Content:	Content{
			Type:		"application/json",
			Body:		[]map[string]string{fields},
		},
	}
}

func parsesFields(s string) map[string]string {
	result := make(map[string]string)
	result["type"] = "application/json"

	for _, line := range strings.Split(s, "\n") {
		split := strings.SplitN(line, "|", 2)
		if len(split) == 2 && split[0] != "" && split[1] != "" {
			result[split[0]] = split[1]
		}
	}

	return result
}