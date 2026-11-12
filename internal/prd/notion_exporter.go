package prd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// NotionRichText represents rich text elements in the Notion API.
type NotionRichText struct {
	Type string `json:"type"`
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
}

// NotionBlock represents a structural block in the Notion API.
type NotionBlock struct {
	Object    string                 `json:"object"`
	Type      string                 `json:"type"`
	Heading1  *NotionHeadingBlock    `json:"heading_1,omitempty"`
	Heading2  *NotionHeadingBlock    `json:"heading_2,omitempty"`
	Heading3  *NotionHeadingBlock    `json:"heading_3,omitempty"`
	Paragraph *NotionParagraphBlock  `json:"paragraph,omitempty"`
	Bullet    *NotionParagraphBlock  `json:"bulleted_list_item,omitempty"`
	Quote     *NotionParagraphBlock  `json:"quote,omitempty"`
	Code      *NotionCodeBlock       `json:"code,omitempty"`
	Callout   *NotionCalloutBlock    `json:"callout,omitempty"`
}

type NotionHeadingBlock struct {
	RichText []NotionRichText `json:"rich_text"`
}

type NotionParagraphBlock struct {
	RichText []NotionRichText `json:"rich_text"`
}

type NotionCodeBlock struct {
	RichText []NotionRichText `json:"rich_text"`
	Language string           `json:"language"`
}

type NotionCalloutBlock struct {
	RichText []NotionRichText `json:"rich_text"`
	Icon     struct {
		Type  string `json:"type"`
		Emoji string `json:"emoji"`
	} `json:"icon"`
}

func makeRichText(content string) []NotionRichText {
	return []NotionRichText{
		{
			Type: "text",
			Text: struct {
				Content string `json:"content"`
			}{Content: content},
		},
	}
}

// ConvertMarkdownToNotionBlocks transforms a Markdown string into a Notion block tree JSON payload.
func ConvertMarkdownToNotionBlocks(title, md string) string {
	lines := strings.Split(md, "\n")
	var blocks []NotionBlock

	inCodeBlock := false
	var codeBuffer strings.Builder
	codeLang := "plain text"

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle code block fences
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				// Close code block
				blocks = append(blocks, NotionBlock{
					Object: "block",
					Type:   "code",
					Code: &NotionCodeBlock{
						RichText: makeRichText(codeBuffer.String()),
						Language: codeLang,
					},
				})
				codeBuffer.Reset()
				inCodeBlock = false
			} else {
				inCodeBlock = true
				lang := strings.TrimPrefix(trimmed, "```")
				if lang != "" {
					codeLang = lang
				} else {
					codeLang = "plain text"
				}
			}
			continue
		}

		if inCodeBlock {
			codeBuffer.WriteString(line + "\n")
			continue
		}

		if trimmed == "" {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "# "):
			blocks = append(blocks, NotionBlock{
				Object:   "block",
				Type:     "heading_1",
				Heading1: &NotionHeadingBlock{RichText: makeRichText(strings.TrimPrefix(trimmed, "# "))},
			})
		case strings.HasPrefix(trimmed, "## "):
			blocks = append(blocks, NotionBlock{
				Object:   "block",
				Type:     "heading_2",
				Heading2: &NotionHeadingBlock{RichText: makeRichText(strings.TrimPrefix(trimmed, "## "))},
			})
		case strings.HasPrefix(trimmed, "### "):
			blocks = append(blocks, NotionBlock{
				Object:   "block",
				Type:     "heading_3",
				Heading3: &NotionHeadingBlock{RichText: makeRichText(strings.TrimPrefix(trimmed, "### "))},
			})
		case strings.HasPrefix(trimmed, "> "):
			quoteText := strings.TrimPrefix(trimmed, "> ")
			quoteText = strings.Trim(quoteText, "\"")
			blocks = append(blocks, NotionBlock{
				Object: "block",
				Type:   "quote",
				Quote:  &NotionParagraphBlock{RichText: makeRichText(quoteText)},
			})
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			bulletText := strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
			blocks = append(blocks, NotionBlock{
				Object: "block",
				Type:   "bulleted_list_item",
				Bullet: &NotionParagraphBlock{RichText: makeRichText(bulletText)},
			})
		default:
			blocks = append(blocks, NotionBlock{
				Object:    "block",
				Type:      "paragraph",
				Paragraph: &NotionParagraphBlock{RichText: makeRichText(trimmed)},
			})
		}
	}

	payload := map[string]interface{}{
		"title":    title,
		"children": blocks,
	}

	data, _ := json.MarshalIndent(payload, "", "  ")
	return string(data)
}

// ExportToNotionAPI publishes the generated PRD to Notion via their official REST API.
func ExportToNotionAPI(ctx context.Context, pageTitle, blocksJSON, notionToken, parentID string) (string, error) {
	if notionToken == "" || notionToken == "mock" || notionToken == "demo" {
		// Mock export returning a simulated Notion page URL
		return fmt.Sprintf("https://notion.so/feedbackloom/prd-%d", time.Now().Unix()), nil
	}

	var parsed struct {
		Children []NotionBlock `json:"children"`
	}
	if err := json.Unmarshal([]byte(blocksJSON), &parsed); err != nil {
		return "", fmt.Errorf("unmarshal blocks: %w", err)
	}

	requestPayload := map[string]interface{}{
		"parent": map[string]string{
			"page_id": parentID,
		},
		"properties": map[string]interface{}{
			"title": []map[string]interface{}{
				{
					"text": map[string]string{
						"content": pageTitle,
					},
				},
			},
		},
		"children": parsed.Children,
	}

	bodyBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.notion.com/v1/pages", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+notionToken)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("notion api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("notion api error status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	return result.URL, nil
}
