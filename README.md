# go-azure-service-bus-sender

Command notes

```bash
go mod init go-azure-service-bus-sender
go get github.com/Azure/azure-sdk-for-go/sdk/azidentity
go get github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus
export AZURE_SERVICEBUS_CONNECTION_STRING="Endpoint=sb://<YOUR_NAMESPACE>.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=<YOUR_SHARED_ACCESS_KEY>"
export AZURE_SERVICEBUS_QUEUE_NAME="myqueue"
export BATCH_SIZE=10
go run main.go -batchSize=10
```
