// Package main demonstrates how to attach external links to issues.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one team
//
// This example shows:
// 1. Creating a test issue
// 2. Creating a custom attachment with metadata
// 3. Attaching a documentation URL
// 4. Attaching a GitHub PR URL
// 5. Attaching a Slack thread URL
// 6. Deleting attachments
// 7. Cleaning up test data
//
// Attachment types:
// - AttachmentCreate: Custom attachments with metadata
// - AttachmentLinkURL: Any external URL
// - AttachmentLinkGitHubPR: GitHub Pull Requests (auto-extracts title)
// - AttachmentLinkSlack: Slack threads
// - AttachmentDelete: Remove attachments
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/attach_links
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatal("LINEAR_API_KEY environment variable not set")
	}

	// Create client
	client, err := linear.NewClient(apiKey)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Step 1: Get team and create test issue
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		log.Fatalf("Failed to get teams: %v", err)
	}
	if len(teams.Nodes) == 0 {
		log.Fatal("No teams found")
	}
	teamID := teams.Nodes[0].ID

	fmt.Println("Creating test issue...")
	title := "Test Issue - Attachment Demo"
	desc := "This issue demonstrates attaching external resources"
	issue, err := client.IssueCreate(ctx, intgraphql.IssueCreateInput{
		TeamID:      teamID,
		Title:       &title,
		Description: &desc,
	})
	if err != nil {
		log.Fatalf("Failed to create issue: %v", err)
	}
	fmt.Printf("Created issue: [%.0f] %s\n", issue.Number, issue.Title)
	fmt.Println()

	attachments := []string{}

	// Step 2: Create a custom attachment with metadata
	fmt.Println("Creating custom attachment...")
	customTitle := "Build Status"
	customSubtitle := "CI Build #456 - Passed"
	customURL := "https://ci.example.com/builds/456"

	customAttachment, err := client.AttachmentCreate(ctx, intgraphql.AttachmentCreateInput{
		IssueID:  issue.ID,
		Title:    customTitle,
		Subtitle: &customSubtitle,
		URL:      customURL,
	})
	if err != nil {
		log.Printf("Note: Custom attachment creation may require OAuth app mode. Error: %v\n", err)
	} else {
		fmt.Printf("✓ Custom attachment created!\n")
		fmt.Printf("  Title: %s\n", customAttachment.Title)
		if customAttachment.Subtitle != nil {
			fmt.Printf("  Subtitle: %s\n", *customAttachment.Subtitle)
		}
		if customAttachment.URL != "" {
			fmt.Printf("  URL: %s\n", customAttachment.URL)
		}
		fmt.Printf("  ID: %s\n", customAttachment.ID)
		attachments = append(attachments, customAttachment.ID)
		fmt.Println()
	}

	// Step 3: Attach a documentation URL
	docsURL := "https://docs.example.com/api/troubleshooting"
	docsTitle := "API Troubleshooting Guide"

	fmt.Printf("Attaching documentation URL...\n")
	attachment1, err := client.AttachmentLinkURL(ctx, issue.ID, docsURL, &docsTitle)
	if err != nil {
		log.Fatalf("Failed to attach URL: %v", err)
	}

	fmt.Printf("✓ Documentation attached!\n")
	fmt.Printf("  Title: %s\n", attachment1.Title)
	fmt.Printf("  URL: %s\n", attachment1.URL)
	fmt.Printf("  ID: %s\n", attachment1.ID)
	attachments = append(attachments, attachment1.ID)
	fmt.Println()

	// Step 4: Attach a GitHub PR (with auto-title extraction)
	prURL := "https://github.com/chainguard-sandbox/go-linear/pull/1"

	fmt.Printf("Attaching GitHub PR...\n")
	attachment2, err := client.AttachmentLinkGitHubPR(ctx, issue.ID, prURL)
	if err != nil {
		// GitHub integration may not be configured - this is expected
		fmt.Printf("Note: GitHub PR link requires GitHub integration in Linear workspace\n")
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Continuing with documentation URL example only...")
	} else {
		fmt.Printf("✓ GitHub PR attached!\n")
		fmt.Printf("  Title: %s (auto-extracted)\n", attachment2.Title)
		fmt.Printf("  URL: %s\n", attachment2.URL)
		fmt.Printf("  ID: %s\n", attachment2.ID)
		attachments = append(attachments, attachment2.ID)
		fmt.Println()
	}

	// Step 5: Attach a Slack thread
	slackURL := "https://myworkspace.slack.com/archives/C123ABC/p1234567890"

	fmt.Printf("Attaching Slack thread...\n")
	slackAttachment, err := client.AttachmentLinkSlack(ctx, issue.ID, slackURL)
	if err != nil {
		// Slack integration may not be configured
		fmt.Printf("Note: Slack link requires Slack integration in Linear workspace\n")
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Continuing with other examples...")
	} else {
		fmt.Printf("✓ Slack thread attached!\n")
		fmt.Printf("  Title: %s\n", slackAttachment.Title)
		fmt.Printf("  URL: %s\n", slackAttachment.URL)
		fmt.Printf("  ID: %s\n", slackAttachment.ID)
		attachments = append(attachments, slackAttachment.ID)
		fmt.Println()
	}

	// Step 6: Delete attachments (cleanup)
	if len(attachments) > 0 {
		fmt.Printf("Deleting %d attachments...\n", len(attachments))
		for _, attID := range attachments {
			if err := client.AttachmentDelete(ctx, attID); err != nil {
				log.Printf("Warning: Failed to delete attachment: %v", err)
			}
		}
		fmt.Println("✓ Attachments deleted")
		fmt.Println()
	}

	// Step 7: Clean up test issue
	fmt.Println("Cleaning up test issue...")
	if err := client.IssueDelete(ctx, issue.ID); err != nil {
		log.Printf("Warning: Failed to delete issue: %v", err)
	}
	fmt.Println("✓ Test issue deleted")
	fmt.Println()

	fmt.Println("Attachment Features:")
	fmt.Println("  - AttachmentCreate: Custom attachments with metadata (CI builds, etc.)")
	fmt.Println("  - AttachmentLinkURL: Link any external resource")
	fmt.Println("  - AttachmentLinkGitHubPR: Link PRs (auto-extracts metadata)")
	fmt.Println("  - AttachmentLinkSlack: Link Slack discussions")
	fmt.Println("  - AttachmentDelete: Remove attachments")
	fmt.Println()
	fmt.Println("Also available (see schema):")
	fmt.Println("  - AttachmentLinkJiraIssue: Migration from Jira")
	fmt.Println("  - AttachmentLinkIntercom, Salesforce, Zendesk, etc.")
	fmt.Println()
	fmt.Println("Use cases:")
	fmt.Println("  - Link PRs to issues for code traceability")
	fmt.Println("  - Attach design docs, API specs, runbooks")
	fmt.Println("  - Connect Linear to Slack discussions")
	fmt.Println("  - Custom integrations with CI/CD, monitoring tools")
	fmt.Println("  - Migrate from other tools with preserved links")
}
