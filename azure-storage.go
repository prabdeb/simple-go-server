package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

// Azure Storage Quickstart Sample - Demonstrate how to upload, list, download, and delete blobs.
//
// Documentation References:
// - What is a Storage Account - https://docs.microsoft.com/azure/storage/common/storage-create-storage-account
// - Blob Service Concepts - https://docs.microsoft.com/rest/api/storageservices/Blob-Service-Concepts
// - Blob Service Go SDK API - https://godoc.org/github.com/Azure/azure-storage-blob-go
// - Blob Service REST API - https://docs.microsoft.com/rest/api/storageservices/Blob-Service-REST-API
// - Scalability and performance targets - https://docs.microsoft.com/azure/storage/common/storage-scalability-targets
// - Azure Storage Performance and Scalability checklist https://docs.microsoft.com/azure/storage/common/storage-performance-checklist
// - Storage Emulator - https://docs.microsoft.com/azure/storage/common/storage-use-emulator

// KeyVaultConfiguration for property file json keys
type KeyVaultConfiguration struct {
	AzureStorageAccount   string `json:"azureStorageAccount"`
	AzureStorageAccessKey string `json:"azureStorageAccessKey"`
}

func randomString() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return strconv.Itoa(r.Int())
}

func handleErrors(err error) {
	if err != nil {
		if serr, ok := err.(azblob.StorageError); ok { // This error is a Service-specific
			switch serr.ServiceCode() { // Compare serviceCode to ServiceCodeXxx constants
			case azblob.ServiceCodeContainerAlreadyExists:
				fmt.Println("Received 409. Container already exists")
				return
			}
		}
		log.Fatal(err)
	}
}

func storageTest() (string, error) {
	fmt.Printf("Azure Blob storage quick start sample\n")
	testLog := "Azure Blob storage quick start sample\n"

	// From the Azure portal, get your storage account name and key and set environment variables.
	accountName, accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT"), os.Getenv("AZURE_STORAGE_ACCESS_KEY")
	if len(accountName) == 0 || len(accountKey) == 0 {
		azureKeyVaultSecretFile := os.Getenv("AZURE_KEYVAULT_SECRET_FILE")
		if len(azureKeyVaultSecretFile) != 0 {
			//filename is the path to the json config file
			file, err := ioutil.ReadFile(azureKeyVaultSecretFile)
			if err != nil {
				log.Println("Not able to read the File: " + err.Error())
				testLog += "Not able to read the File: " + err.Error()
				return testLog, err
			}
			var keyVaultConfiguration KeyVaultConfiguration
			err = json.Unmarshal([]byte(file), &keyVaultConfiguration)
			if err != nil {
				log.Println("Not able to parse the File: " + err.Error())
				testLog += "Not able to parse the File: " + err.Error()
				return testLog, err
			}
			accountName = keyVaultConfiguration.AzureStorageAccount
			accountKey = keyVaultConfiguration.AzureStorageAccessKey
		} else {
			log.Println("Either the AZURE_STORAGE_ACCOUNT or AZURE_STORAGE_ACCESS_KEY environment variable is not set")
			testLog += "Either the AZURE_STORAGE_ACCOUNT or AZURE_STORAGE_ACCESS_KEY environment variable is not set"
			return testLog, nil
		}
	}

	// Create a default request pipeline using your storage account name and account key.
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Println("Invalid credentials with error: " + err.Error())
		testLog += "Invalid credentials with error: " + err.Error()
		return testLog, err
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// Create a random string for the quick start container
	containerName := fmt.Sprintf("quickstart-%s", randomString())

	// From the Azure portal, get your storage account blob service URL endpoint.
	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))

	// Create a ContainerURL object that wraps the container URL and a request
	// pipeline to make requests.
	containerURL := azblob.NewContainerURL(*URL, p)

	// Create the container
	fmt.Printf("Creating a container named %s\n", containerName)
	testLog += "Creating a container named " + containerName + "\n"
	ctx := context.Background() // This example uses a never-expiring context
	_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	handleErrors(err)

	// Create a file to test the upload and download.
	fmt.Printf("Creating a dummy file to test the upload and download\n")
	testLog += "Creating a dummy file to test the upload and download\n"
	data := []byte("hello world this is a blob\n")
	fileName := randomString()
	err = ioutil.WriteFile(fileName, data, 0700)
	handleErrors(err)

	// Here's how to upload a blob.
	blobURL := containerURL.NewBlockBlobURL(fileName)
	file, err := os.Open(fileName)
	handleErrors(err)

	// You can use the low-level PutBlob API to upload files. Low-level APIs are simple wrappers for the Azure Storage REST APIs.
	// Note that PutBlob can upload up to 256MB data in one shot. Details: https://docs.microsoft.com/en-us/rest/api/storageservices/put-blob
	// Following is commented out intentionally because we will instead use UploadFileToBlockBlob API to upload the blob
	// _, err = blobURL.PutBlob(ctx, file, azblob.BlobHTTPHeaders{}, azblob.Metadata{}, azblob.BlobAccessConditions{})
	// handleErrors(err)

	// The high-level API UploadFileToBlockBlob function uploads blocks in parallel for optimal performance, and can handle large files as well.
	// This function calls PutBlock/PutBlockList for files larger 256 MBs, and calls PutBlob for any file smaller
	fmt.Printf("Uploading the file with blob name: %s\n", fileName)
	testLog += "Uploading the file with blob name: " + fileName + "\n"
	_, err = azblob.UploadFileToBlockBlob(ctx, file, blobURL, azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16})
	handleErrors(err)

	// List the container that we have created above
	fmt.Println("Listing the blobs in the container:")
	testLog += "Listing the blobs in the container:"
	for marker := (azblob.Marker{}); marker.NotDone(); {
		// Get a result segment starting with the blob indicated by the current Marker.
		listBlob, err := containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{})
		handleErrors(err)

		// ListBlobs returns the start of the next segment; you MUST use this to get
		// the next segment (after processing the current result segment).
		marker = listBlob.NextMarker

		// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
		for _, blobInfo := range listBlob.Segment.BlobItems {
			fmt.Print("	Blob name: " + blobInfo.Name + "\n")
			testLog += "	Blob name: " + blobInfo.Name + "\n"
		}
	}

	// Here's how to download the blob
	downloadResponse, err := blobURL.Download(ctx, 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false)

	// NOTE: automatically retries are performed if the connection fails
	bodyStream := downloadResponse.Body(azblob.RetryReaderOptions{MaxRetryRequests: 20})

	// read the body into a buffer
	downloadedData := bytes.Buffer{}
	_, err = downloadedData.ReadFrom(bodyStream)
	handleErrors(err)

	// The downloaded blob data is in downloadData's buffer. :Let's print it
	fmt.Printf("Downloaded the blob: " + downloadedData.String())
	testLog += "Downloaded the blob: " + downloadedData.String()

	// Cleaning up the quick start by deleting the container and the file created locally
	fmt.Printf("Cleaning up.\n")
	testLog += "Cleaning up.\n"
	containerURL.Delete(ctx, azblob.ContainerAccessConditions{})
	file.Close()
	os.Remove(fileName)

	return testLog, nil
}
