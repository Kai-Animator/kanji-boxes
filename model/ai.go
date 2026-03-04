package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const defaultOpenAIModel = "gpt-4o-mini"

type aiFillResultMsg struct {
	values map[int]string
	err    error
}

var addFieldNames = map[int]string{
	addKanjiField:    "kanji",
	addEnglishField:  "english",
	addHiraganaField: "hiragana",
	addUsageField:    "usage",
	addTagsField:     "tags",
}

var addFieldIndexByName = map[string]int{
	"kanji":    addKanjiField,
	"english":  addEnglishField,
	"hiragana": addHiraganaField,
	"usage":    addUsageField,
	"tags":     addTagsField,
}

func (m Model) requestAIFillMissing() (Model, tea.Cmd) {
	targets := make([]int, 0, addTagsField+1)
	for field := addKanjiField; field <= addTagsField; field++ {
		if strings.TrimSpace(m.addInputs[field].Value()) == "" {
			targets = append(targets, field)
		}
	}
	if len(targets) == 0 {
		m.addMessage = "No missing fields. Use ctrl+r to regenerate one field."
		return m, nil
	}
	return m.requestAIFill(targets)
}

func (m Model) requestAIFillFocused() (Model, tea.Cmd) {
	if m.addFocus < addKanjiField || m.addFocus > addTagsField {
		m.addMessage = "Focus a card field to regenerate (Kanji/English/Hiragana/Usage/Tags)."
		return m, nil
	}
	return m.requestAIFill([]int{m.addFocus})
}

func (m Model) requestAIFill(targets []int) (Model, tea.Cmd) {
	if m.aiBusy {
		m.addMessage = "AI generation already in progress."
		return m, nil
	}

	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		m.addMessage = "OPENAI_API_KEY is required. Set it in your environment first."
		return m, nil
	}

	known := make(map[string]string)
	for field := addKanjiField; field <= addTagsField; field++ {
		value := strings.TrimSpace(m.addInputs[field].Value())
		if value == "" {
			continue
		}
		known[addFieldNames[field]] = value
	}
	if len(known) == 0 {
		m.addMessage = "Fill at least one card field before requesting AI."
		return m, nil
	}

	m.aiBusy = true
	m.addMessage = "Generating suggestions..."
	return m, generateAIFillCmd(apiKey, known, targets)
}

func generateAIFillCmd(apiKey string, known map[string]string, targets []int) tea.Cmd {
	return func() tea.Msg {
		values, err := generateAIFill(apiKey, known, targets)
		return aiFillResultMsg{values: values, err: err}
	}
}

func generateAIFill(apiKey string, known map[string]string, targets []int) (map[int]string, error) {
	targetNames := make([]string, 0, len(targets))
	for _, target := range targets {
		name, ok := addFieldNames[target]
		if !ok {
			continue
		}
		targetNames = append(targetNames, name)
	}
	if len(targetNames) == 0 {
		return nil, errors.New("no valid target fields")
	}

	knownJSON, err := json.Marshal(known)
	if err != nil {
		return nil, err
	}

	userPrompt := fmt.Sprintf(
		"Known fields JSON: %s\nTarget fields: %s\nReturn a strict JSON object with only the target fields. If uncertain, make the best plausible guess. Rules: kanji=single Japanese expression, english=short meaning, hiragana=reading only in hiragana, usage=one natural Japanese sentence, tags=pipe-separated tags.",
		string(knownJSON),
		strings.Join(targetNames, ", "),
	)

	requestBody := map[string]any{
		"model": defaultOpenAIModel,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You complete Japanese flashcard fields. Return valid JSON only.",
			},
			{
				"role":    "user",
				"content": userPrompt,
			},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.7,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("openai error (%d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var completion struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &completion); err != nil {
		return nil, err
	}
	if len(completion.Choices) == 0 {
		return nil, errors.New("openai returned no choices")
	}

	content := strings.TrimSpace(stripCodeFences(completion.Choices[0].Message.Content))
	if content == "" {
		return nil, errors.New("openai returned empty content")
	}

	var generated map[string]any
	if err := json.Unmarshal([]byte(content), &generated); err != nil {
		return nil, fmt.Errorf("invalid json response: %w", err)
	}

	values := make(map[int]string)
	for fieldName, raw := range generated {
		index, ok := addFieldIndexByName[fieldName]
		if !ok {
			continue
		}
		value := strings.TrimSpace(fmt.Sprintf("%v", raw))
		if fieldName == "tags" {
			value = normalizeTagsOutput(value)
		}
		values[index] = value
	}

	return values, nil
}

func stripCodeFences(value string) string {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}
	trimmed = strings.TrimPrefix(trimmed, "```")
	if idx := strings.Index(trimmed, "\n"); idx >= 0 {
		trimmed = trimmed[idx+1:]
	}
	if idx := strings.LastIndex(trimmed, "```"); idx >= 0 {
		trimmed = trimmed[:idx]
	}
	return strings.TrimSpace(trimmed)
}

func normalizeTagsOutput(value string) string {
	if value == "" {
		return ""
	}
	replacer := strings.NewReplacer("\n", "|", ",", "|", ";", "|", "、", "|", " ", "|")
	normalized := replacer.Replace(value)
	parts := strings.Split(normalized, "|")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		cleaned = append(cleaned, part)
	}
	return strings.Join(cleaned, "|")
}
