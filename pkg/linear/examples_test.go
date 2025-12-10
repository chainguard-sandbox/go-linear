package linear_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func ExampleNewClient() {
	client, err := linear.NewClient("lin_api_xxx")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	viewer, err := client.Viewer(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Authenticated as: %s\n", viewer.Name)
}

func ExampleNewClient_withOptions() {
	client, err := linear.NewClient("lin_api_xxx",
		linear.WithTimeout(60), // Custom timeout
		linear.WithUserAgent("my-app/1.0"),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = client // Use client
}

func ExampleClient_Issues() {
	client, _ := linear.NewClient("lin_api_xxx")

	// List first 10 issues
	first := int64(10)
	issues, err := client.Issues(context.Background(), &first, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, issue := range issues.Nodes {
		fmt.Printf("- %s\n", issue.Title)
	}
}

func ExampleNewIssueIterator() {
	client, _ := linear.NewClient("lin_api_xxx")

	// Automatically iterate through all issues
	iter := linear.NewIssueIterator(client, 50)
	for {
		issue, err := iter.Next(context.Background())
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Issue: %s\n", issue.Title)
	}
}
