package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	httpclient "github.com/tmc/langchaingo/llms/tongyi/internal/tongyiclient/httpclient"

	"github.com/google/uuid"
)

func ConnRecognitionClient() (*httpclient.WsClient, error) {
	// Initialize the client with the necessary parameters.
	header := http.Header{}
	header.Add("Authorization", os.Getenv("DASHSCOPE_API_KEY"))

	headerPara := ReqHeader{
		Streaming: "duplex",
		TaskID:    generateTaskID(),
		Action:    "run-task",
	}
	payload := PayloadIn{
		Model: "paraformer-realtime-v1",
		Parameters: Parameters{
			// only support 16000 sample-rate now
			SampleRate: 16000,
			Format:     "pcm",
		},
		Input:     map[string]interface{}{},
		Task:      "asr",
		TaskGroup: "audio",
		Function:  "recognition",
	}

	req := &Request{
		Header:  headerPara,
		Payload: payload,
	}

	client := httpclient.NewWsClient(PARAFORMER_WS_URL, header)

	fmt.Println("conn client...")
	client.ConnClient(req)

	return client, nil
}

func CloseRecognitionClient(cli *httpclient.WsClient) {
	cli.CloseClient()
}

func SendRadioData(cli *httpclient.WsClient, bytesData []byte) {
	cli.SendBinaryDates(bytesData)
}

type ResultWriter interface {
	WriteResult(string) error
}

func HandleRecognitionResult(cli *httpclient.WsClient, writer ResultWriter) {
	outputChan, errChan := cli.ResultChans()

BREAK_FOR:
	for {
		select {
		case output, ok := <-outputChan:
			if !ok {
				fmt.Println("outputChan is closed")
				break BREAK_FOR
			}

			// named pipe out to rust
			writer.WriteResult(string(output.Data) + "\n")

		case err := <-errChan:
			if err != nil {
				fmt.Println("error: ", err)
				// TODO: raise error
				break BREAK_FOR
			}
		}
	}

	fmt.Println("get recognition result...over")
}

// task_id length 32.
func generateTaskID() string {
	u, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	uuid := strings.Replace(u.String(), "-", "", -1)

	return uuid
}
