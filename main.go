package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func GetClient() *azservicebus.Client {
	connectionString, ok := os.LookupEnv("AZURE_SERVICEBUS_CONNECTION_STRING") //ex: Endpoint=sb://<YOUR_NAMESPACE>.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=<YOUR_SHARED_ACCESS_KEY>
	if !ok {
		panic("AZURE_SERVICEBUS_CONNECTION_STRING environment variable not found")
	}

	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		panic(err)
	}
	return client
}

func SendMessageBatch(messages []string, client *azservicebus.Client) {
	queue, ok := os.LookupEnv("AZURE_SERVICEBUS_QUEUE_NAME") //ex: myqueue
	if !ok {
		panic("AZURE_SERVICEBUS_QUEUE_NAME environment variable not found")
	}
	sender, err := client.NewSender(queue, nil)
	if err != nil {
		panic(err)
	}
	defer sender.Close(context.TODO())

	batch, err := sender.NewMessageBatch(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	for _, message := range messages {
		err := batch.AddMessage(&azservicebus.Message{Body: []byte(message)}, nil)
		if errors.Is(err, azservicebus.ErrMessageTooLarge) {
			fmt.Printf("Message batch is full. We should send it and create a new one.\n")
		}
	}

	if err := sender.SendMessageBatch(context.TODO(), batch, nil); err != nil {
		panic(err)
	}
}

func main() {
	batchNbr := 0

	for {
		batchNbr++
		messageNbr := 0
		// starting with a batch size of 1, send messages to service bus and double the batch size each time until we reach 256
		for i := 1; i <= 256; i *= 2 {
			messages := make([]string, i)
			for j := 0; j < i; j++ {
				messageNbr++
				messages[j] = "batch " + strconv.Itoa(batchNbr) + " message " + strconv.Itoa(messageNbr)
			}

			// log the batch size and the number of messages in the batch
			fmt.Printf("Sending batch %d with %d messages\n", batchNbr, i)

			client := GetClient()
			SendMessageBatch(messages, client)
		}

		// cool down
		time.Sleep(2 * time.Minute)
	}
}
