package main

import (
	"fmt"
	"os"
	"strings"

	jira "github.com/andygrunwald/go-jira"
)

func main() {
	tp := jira.BasicAuthTransport{
		Username: strings.TrimSpace(os.Getenv("JIRA_USERNAME")),
		Password: strings.TrimSpace(os.Getenv("JIRA_PASSWORD")),
	}

	client, err := jira.NewClient(tp.Client(), strings.TrimSpace(os.Getenv("JIRA_URL")))
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	//jira.InitIssueWithMetaAndFields()
	meta, _, err := client.Issue.GetCreateMeta("PC")

	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	issueType, err := meta.GetIssueTypeWithName("Help")
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	issueTypeMeta, _, err := client.Issue.GetIssueTypeMeta("PC", issueType)
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	fmt.Printf("\nissueTypeMeta: %#v\n", issueTypeMeta)
	for _, m := range meta.IssueTypes {
		fmt.Printf("\n%#v\n", m)
	}
}
