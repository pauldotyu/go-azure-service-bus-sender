package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func GetClient() *azservicebus.Client {
	// Use the connection string if it is available, otherwise use Azure Identity
	connectionString, ok := os.LookupEnv("AZURE_SERVICEBUS_CONNECTIONSTRING") //ex: Endpoint=sb://<YOUR_NAMESPACE>.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=<YOUR_SHARED_ACCESS_KEY>

	if ok {
		client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
		if err != nil {
			log.Fatalf("failed to create a Service Bus client: %v", err)
		} else {
			fmt.Println("Service Bus client created from connection string")
			return client
		}
	} else {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			log.Fatalf("failed to obtain a credential: %v", err)
		}
		sbHostname, ok := os.LookupEnv("AZURE_SERVICEBUS_FULLYQUALIFIEDNAMESPACE") //ex: <YOUR_NAMESPACE>.servicebus.windows.net
		if !ok {
			panic("AZURE_SERVICEBUS_FULLYQUALIFIEDNAMESPACE environment variable not found")
		}
		client, err := azservicebus.NewClient(sbHostname, cred, nil)
		if err != nil {
			log.Fatalf("failed to create a Service Bus client: %v", err)
		} else {
			fmt.Println("Service Bus client created with Azure Identity")
			return client
		}
	}

	return nil
}

func SendMessageBatch(messages []string, client *azservicebus.Client) {
	queue, ok := os.LookupEnv("AZURE_SERVICEBUS_QUEUENAME") //ex: myqueue
	if !ok {
		panic("AZURE_SERVICEBUS_QUEUENAME environment variable not found")
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

		client := GetClient()
		// starting with a batch size of 1, send messages to service bus and double the batch size each time until we reach 256
		for i := 1; i <= 256; i *= 2 {
			messages := make([]string, i)
			for j := 0; j < i; j++ {
				messageNbr++
				messages[j] = "batch " + strconv.Itoa(batchNbr) + " message " + strconv.Itoa(messageNbr)
			}

			// log the batch size and the number of messages in the batch
			fmt.Printf("Sending batch %d with %d messages\n", batchNbr, i)

			SendMessageBatch(messages, client)
		}

		fmt.Printf("Backing off for 2 minutes\n")
		time.Sleep(2 * time.Minute)
	}
}
