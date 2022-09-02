package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

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
	batchSize, ok := os.LookupEnv("BATCH_SIZE") //ex: 10
	if !ok {
		panic("BATCH_SIZE environment variable not found")
	}

	batchSizeInt, err := strconv.Atoi(batchSize)
	if err != nil {
		panic(err)
	}

	messages := []string{}
	for i := 1; i <= batchSizeInt; i++ {
		messages = append(messages, "message "+strconv.Itoa(i))
	}

	client := GetClient()
	fmt.Println("send messages as a batch...")
	SendMessageBatch(messages[:], client)
}