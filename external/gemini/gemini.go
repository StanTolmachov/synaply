package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"synaply/slogger"
)

type Service interface {
	WordInfo(ctx context.Context, req WordInfoRequest) (string, error)
	StartPracticeWithGemini(ctx context.Context, req *PracticeWithGemini, wordList string) (*StartPracticeWithGeminiResponse, error)
	CheckAnswerPracticeWithGemini(ctx context.Context, req *PracticeWithGemini, translate string) (*CheckAnswerPracticeWithGeminiResponse, error)
	WordList(ctx context.Context, req WordListReq) ([]WordListResp, error)
	WordTranslate(ctx context.Context, req *GemTranslationReq) (*GemTranslationResp, error)
}

type service struct {
	client *genai.Client
	Model  string
}

func NewService(key string, model string) (Service, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{APIKey: key})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}
	return &service{
		client: client,
		Model:  model,
	}, nil
}

const WordInfoPromptTemplate = `You are an expert linguistic analysis and teaching algorithm.

Source Language Code: "%[1]s"
Target Language Code: "%[2]s"

Input: The user wants to learn the target word "%[4]s" (Target Language) which translates to "%[3]s" in their native language (Source Language).
First, identify the actual Source Language based on the code "%[1]s" and the word "%[3]s".

YOUR TASK: Provide a comprehensive and detailed educational explanation of the word "%[4]s".

CRITICAL RULES:
1. ALL explanations MUST be written entirely in the Source Language.
2. DO NOT use English for explanations unless the Source Language is English.
3. Pronunciation must be written using ONLY the letters of the Source Language alphabet. Capitalize the stressed vowel.
4. Keep explanations concise, highly structured, and easy to read.

Format your response exactly like this in Markdown (translate the bold headers to the Source Language):

**[Pronunciation Header]**: [pronunciation]
**[Grammar & Level Header]**: [CEFR level, part of speech, gender, irregular forms]
**[Meaning & Usage Header]**: [Explain the precise meaning, context of usage (e.g., formal/informal), and any subtle nuances compared to its direct translation]
**[Synonyms & Antonyms Header]**: [Provide 2-3 synonyms and 1-2 antonyms, if applicable]
**[Collocations Header]**: [Provide 2 or 3 most common word combinations or phrases. E.g., "heavy rain", "make a decision"]
**[Examples Header]**:
1. [Example in Target Lang] - [Translation in Source Lang]
2. [Example in Target Lang] - [Translation in Source Lang]
3. [Example in Target Lang] - [Translation in Source Lang]`

var ErrLimitExceeded = errors.New("gemini limit exceeded")

func (s *service) WordInfo(ctx context.Context, req WordInfoRequest) (string, error) {

	promt := BuildGeminiPrompt(req)

	result, err := s.client.Models.GenerateContent(
		ctx,
		s.Model,
		genai.Text(promt),
		nil,
	)

	if err != nil {
		slogger.Log.ErrorContext(ctx, "Genai client response error", "error", err)
		if strings.Contains(err.Error(), "429") || strings.Contains(strings.ToLower(err.Error()), "quota") {
			return "", ErrLimitExceeded
		}
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if result == nil || result.Text() == "" {
		return "", fmt.Errorf("no result")
	}

	slogger.Log.DebugContext(ctx, "Genai promt and  result", "promt", promt, "result", result)

	return result.Text(), nil
}
func BuildGeminiPrompt(req WordInfoRequest) string {
	// I use explicit argument indexing %[1]s, %[2]s, etc.
	// This is a cool Go feature! It allows you to pass an argument once,
	// but use it in the template multiple times without duplicating it in the code.
	return fmt.Sprintf(
		WordInfoPromptTemplate,
		req.SourceLang, // %[1]s
		req.TargetLang, // %[2]s
		req.SourceWord, // %[3]s
		req.TargetWord, // %[4]s
	)
}

const StartPracticeWithGeminiPromptTemplate = `You are a test-generating algorithm.

Source Language Code: "%[1]s"
Target Language Code: "%[2]s"
Topic: "%[3]s"

Input: A list of words in "SourceWord - TargetWord" format.
First, identify the actual languages based on the codes and the provided words (e.g., if "%[1]s" is "UK" or "UKR" and the words are Ukrainian, the language is Ukrainian. If "%[1]s" is "RU", it is Russian).

YOUR TASK: Generate exactly 5 sentences ENTIRELY in the Source Language. 
The student will translate these sentences into the Target Language ("%[2]s") later. Therefore, you must provide the base sentences to translate.

CRITICAL RULES:
1. ALL sentences inside the JSON MUST be written in the Source Language.
2. DO NOT use the Target Language ("%[2]s") for the sentences.
3. DO NOT use English (unless the Source Language is actually English).
4. Construct the sentences primarily using the Source words from the list.

FORMATTING RULE: Respond STRICTLY in valid JSON. Do not use markdown code blocks like ` + "```json" + `.

{
  "level": "<CEFR level, e.g., A2>",
  "sentences": [
    "<Sentence 1 strictly in the Source Language>",
    "<Sentence 2 strictly in the Source Language>",
    "<Sentence 3 strictly in the Source Language>",
    "<Sentence 4 strictly in the Source Language>",
    "<Sentence 5 strictly in the Source Language>"
  ]
}
`

func (s *service) StartPracticeWithGemini(ctx context.Context, req *PracticeWithGemini, wordList string) (*StartPracticeWithGeminiResponse, error) {
	SystemPrompt := fmt.Sprintf(
		StartPracticeWithGeminiPromptTemplate,
		req.SourceLang,
		req.TargetLang,
		req.Topic,
	)
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: SystemPrompt},
			},
		},
		Temperature:      genai.Ptr[float32](0.3),
		ResponseMIMEType: "application/json",
	}

	result, err := s.client.Models.GenerateContent(
		ctx,
		s.Model,
		genai.Text(wordList),
		config,
	)

	if err != nil {
		slogger.Log.ErrorContext(ctx, "Genai client response error", "error", err)
		if strings.Contains(err.Error(), "429") || strings.Contains(strings.ToLower(err.Error()), "quota") {
			return nil, ErrLimitExceeded
		}
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if result == nil || result.Text() == "" {
		return nil, fmt.Errorf("no result")
	}

	slogger.Log.DebugContext(ctx, "Genai prompt and result", "result", result.Text())

	cleanJSON := extractJSON(result.Text())
	if cleanJSON == "" {
		slogger.Log.ErrorContext(ctx, "failed to extract JSON from gemini response", "text", result.Text())
		return nil, fmt.Errorf("failed to extract JSON from gemini response")
	}

	var resp StartPracticeWithGeminiResponse
	if err := json.Unmarshal([]byte(cleanJSON), &resp); err != nil {
		slogger.Log.ErrorContext(ctx, "failed to unmarshal gemini response", "error", err, "text", result.Text())
		return nil, fmt.Errorf("failed to unmarshal gemini response: %w", err)
	}

	return &resp, nil
}

const CheckAnswerPracticeWithGeminiPromptTemplate = `You are an automated language-evaluation algorithm.

Source Language Code: "%[1]s" (Language for explanations and localized UI text)
Target Language Code: "%[2]s" (Language of the user's translations)
Topic: "%[3]s"

First, identify the actual Source Language based on the code "%[1]s" (e.g., if "%[1]s" is "UK" or "UKR", the language is strictly Ukrainian).

Input: You will receive the original 5 sentences in the Source Language and the user's translation attempts in the Target Language.

YOUR TASK: Evaluate the user's translations for grammatical correctness and accuracy.

CRITICAL RULES:
1. All general feedback, teacher comments, and localized statuses ("status_localized") MUST be written STRICTLY in the Source Language.
2. DO NOT use English for feedback (unless the Source Language is English).
3. If a sentence is translated perfectly, generate a short praise in the Source Language.
4. If there are mistakes, explain them clearly in the Source Language and provide the ideal translation in the Target Language.
5. IF THE USER SKIPPED A SENTENCE, set the internal status to "skipped", provide the ideal translation, and write "Skipped" (translated to the Source Language) in the comments.

FORMATTING RULE: You MUST respond STRICTLY in valid JSON format. Do not include markdown code blocks (like ` + "```json" + `).

Use exactly this JSON schema:
{
  "general_comment": "<Overall feedback in the Source Language>",
  "results": [
    {
      "sentence_number": 1,
      "your_version": "<The student's text, or '-' if skipped>",
      "status": "<MUST BE ONE OF: 'correct', 'mistake', 'skipped'>",
      "status_localized": "<Translate the status to the Source Language. E.g., 'Correct', 'Has errors', 'Skipped'>",
      "teacher_comment": "<Your explanation in the Source Language>",
      "ideal_translation": "<The correct translation in the Target Language>"
    }
  ]
}`

func (s *service) CheckAnswerPracticeWithGemini(ctx context.Context, req *PracticeWithGemini, translate string) (*CheckAnswerPracticeWithGeminiResponse, error) {
	SystemPrompt := fmt.Sprintf(
		CheckAnswerPracticeWithGeminiPromptTemplate,
		req.SourceLang,
		req.TargetLang,
		req.Topic,
	)
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: SystemPrompt},
			},
		},
		Temperature:      genai.Ptr[float32](0.1),
		ResponseMIMEType: "application/json",
	}

	result, err := s.client.Models.GenerateContent(
		ctx,
		s.Model,
		genai.Text(translate),
		config,
	)

	if err != nil {
		slogger.Log.ErrorContext(ctx, "Genai client response error", "error", err)
		if strings.Contains(err.Error(), "429") || strings.Contains(strings.ToLower(err.Error()), "quota") {
			return nil, ErrLimitExceeded
		}
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if result == nil || result.Text() == "" {
		return nil, fmt.Errorf("no result")
	}

	slogger.Log.DebugContext(ctx, "Genai prompt and result", "result", result.Text())

	cleanJSON := extractJSON(result.Text())
	if cleanJSON == "" {
		slogger.Log.ErrorContext(ctx, "failed to extract JSON from gemini response", "text", result.Text())
		return nil, fmt.Errorf("failed to extract JSON from gemini response")
	}

	var resp CheckAnswerPracticeWithGeminiResponse
	if err := json.Unmarshal([]byte(cleanJSON), &resp); err != nil {
		slogger.Log.ErrorContext(ctx, "failed to unmarshal gemini response", "error", err, "text", result.Text())
		return nil, fmt.Errorf("failed to unmarshal gemini response: %w", err)
	}

	return &resp, nil
}

const WordListRespPromptTemplate = `You are a vocabulary-generating algorithm.

Source Language Code: "%[1]s"
Target Language Code: "%[2]s"
Target Proficiency Level: "%[3]s"
Predefined Topic: "%[4]q"
User's Custom Request: "%[5]q"

First, identify the actual Source and Target languages based on their codes (e.g., if "%[1]s" is "UK" or "UKR", the Source Language is strictly Ukrainian).

YOUR TASK: Generate a highly relevant vocabulary list of around 100 words based on the Topic or Custom Request.

CRITICAL RULES:
1. The complexity of the words MUST strictly match the requested Proficiency Level (%[3]s).
2. The "SourceWord" MUST be written in the Source Language.
3. The "TargetWord" MUST be written in the Target Language.
4. DO NOT use English for the SourceWord or Comment unless the Source Language is English.
5. For the "Comment" field, you MUST write the phonetic pronunciation of the TargetWord. CRITICAL: You MUST transliterate how the TargetWord sounds using ONLY the alphabet of the Source Language ("%[1]s"). DO NOT use letters from the Target Language. DO NOT use International Phonetic Alphabet (IPA) symbols. Capitalize the stressed vowel.

Topic Selection Logic:
- If a "Predefined Topic" is provided, select words relevant to it.
- If a "User's Custom Request" is provided, incorporate this specific context.
- If the "Predefined Topic" is empty, rely ENTIRELY on the "User's Custom Request".

FORMATTING RULE: You MUST respond STRICTLY in valid JSON format. Do not include markdown code blocks (like ` + "```json" + `). Do not include any other text.

Use exactly this JSON schema (array of objects):
[
  {
    "SourceWord": "<Word or phrase in the Source Language>",
    "TargetWord": "<Translation in the Target Language>",
    "Comment": "<Phonetic pronunciation of TargetWord written STRICTLY using the %[1]s alphabet>"
  }
]
`

func (s *service) WordList(ctx context.Context, req WordListReq) ([]WordListResp, error) {
	if req.Topic == "" && req.UserTopic == "" {
		return nil, fmt.Errorf("either topic or user_topic must be provided")
	}

	SystemPrompt := fmt.Sprintf(
		WordListRespPromptTemplate,
		req.SourceLang,
		req.TargetLang,
		req.Level,
		req.Topic,
		req.UserTopic,
	)
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: SystemPrompt},
			},
		},
		Temperature:      genai.Ptr[float32](0.1),
		ResponseMIMEType: "application/json",
	}

	result, err := s.client.Models.GenerateContent(
		ctx,
		s.Model,
		genai.Text("Generate the word list according to the system instructions."),
		config,
	)

	if err != nil {
		slogger.Log.ErrorContext(ctx, "Genai client response error", "error", err)
		if strings.Contains(err.Error(), "429") || strings.Contains(strings.ToLower(err.Error()), "quota") {
			return nil, ErrLimitExceeded
		}
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if result == nil || result.Text() == "" {
		slogger.Log.ErrorContext(ctx, "Gemini returned empty result for WordList", "request", req)
		return nil, fmt.Errorf("no result from Gemini")
	}

	slogger.Log.DebugContext(ctx, "Gemini WordList raw response", "text", result.Text())

	cleanJSON := extractJSON(result.Text())
	if cleanJSON == "" {
		slogger.Log.ErrorContext(ctx, "failed to extract JSON from gemini response", "text", result.Text(), "request", req)
		return nil, fmt.Errorf("failed to extract JSON from gemini response")
	}

	var resp []WordListResp
	if err := json.Unmarshal([]byte(cleanJSON), &resp); err != nil {
		slogger.Log.ErrorContext(ctx, "failed to unmarshal gemini response", "error", err, "text", result.Text(), "request", req)
		return nil, fmt.Errorf("failed to unmarshal gemini response: %w", err)
	}

	return resp, nil
}

func extractJSON(s string) string {
	firstOpenBrace := strings.Index(s, "{")
	firstOpenBracket := strings.Index(s, "[")

	start := -1
	if firstOpenBrace != -1 && (firstOpenBracket == -1 || firstOpenBrace < firstOpenBracket) {
		start = firstOpenBrace
	} else {
		start = firstOpenBracket
	}

	if start == -1 {
		return ""
	}

	lastCloseBrace := strings.LastIndex(s, "}")
	lastCloseBracket := strings.LastIndex(s, "]")

	end := -1
	if lastCloseBrace != -1 && (lastCloseBracket == -1 || lastCloseBrace > lastCloseBracket) {
		end = lastCloseBrace
	} else {
		end = lastCloseBracket
	}

	if end == -1 || end < start {
		return ""
	}

	return s[start : end+1]
}

const WordTranslatePromptTemplate = `You are a precise bilingual translation algorithm.

Source Language Code: "%[1]s"
Target Language Code: "%[2]s"
Source Word (optional): "%[3]q"
Target Word (optional): "%[4]q"

First, identify the actual Source and Target languages based on the codes and provided words.

YOUR TASK:
- If "Source Word" is provided, translate it into the Target Language.
- If "Target Word" is provided, translate it into the Source Language.
- If both are provided, prioritize "Source Word" -> "Target Language".
- Return one clear dictionary-form translation (base form, no explanations).

CRITICAL RULES:
1. Output must be STRICTLY valid JSON.
2. Do not use markdown code blocks.
3. Do not add extra text outside JSON.
4. Preserve original script/case where appropriate.
5. For multi-word inputs, return the natural equivalent phrase.

Use exactly this JSON schema:
{
  "SourceWord": "<word in Source Language>",
  "TargetWord": "<word in Target Language>"
}`

func (s *service) WordTranslate(ctx context.Context, req *GemTranslationReq) (*GemTranslationResp, error) {
	var systemPrompt string
	if req.SourceWord != "" {
		systemPrompt = fmt.Sprintf(
			WordTranslatePromptTemplate,
			req.SourceLang,
			req.TargetLang,
			req.SourceWord,
			"",
		)
	} else {
		systemPrompt = fmt.Sprintf(
			WordTranslatePromptTemplate,
			req.SourceLang,
			req.TargetLang,
			"",
			req.TargetWord,
		)
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: systemPrompt},
			},
		},
		Temperature:      genai.Ptr[float32](0.1),
		ResponseMIMEType: "application/json",
	}

	result, err := s.client.Models.GenerateContent(
		ctx,
		s.Model,
		genai.Text("Generate the translation according to the system instructions."),
		config)

	if err != nil {
		slogger.Log.ErrorContext(ctx, "Genai client response error", "error", err)
		if strings.Contains(err.Error(), "429") || strings.Contains(strings.ToLower(err.Error()), "quota") {
			return nil, ErrLimitExceeded
		}
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if result == nil || result.Text() == "" {
		slogger.Log.ErrorContext(ctx, "Gemini returned empty result for WordList", "request", req)
		return nil, fmt.Errorf("no result from Gemini")
	}

	slogger.Log.DebugContext(ctx, "Gemini translation raw response", "text", result.Text())

	cleanJSON := extractJSON(result.Text())
	if cleanJSON == "" {
		slogger.Log.ErrorContext(ctx, "failed to extract JSON from gemini response", "text", result.Text(), "request", req)
		return nil, fmt.Errorf("failed to extract JSON from gemini response")
	}

	var resp GemTranslationResp
	if err := json.Unmarshal([]byte(cleanJSON), &resp); err != nil {
		slogger.Log.ErrorContext(ctx, "failed to unmarshal gemini response", "error", err, "text", result.Text(), "request", req)
		return nil, fmt.Errorf("failed to unmarshal gemini response: %w", err)
	}

	return &resp, nil
}
