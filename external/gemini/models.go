package gemini

import (
	"synaply/internal/repository/modelsDB"
)

type WordInfoRequest struct {
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
	SourceWord string `json:"source_word"`
	TargetWord string `json:"target_word"`
}

type WordInfoResponse struct {
	Response string `json:"response"`
}

func ReqToGeminiReq(req WordInfoRequest) *modelsDB.GeminiReq {
	return &modelsDB.GeminiReq{
		SourceLang: req.SourceLang,
		TargetLang: req.TargetLang,
		SourceWord: req.SourceWord,
		TargetWord: req.TargetWord,
	}
}

func GeminiRespToResp(resp *modelsDB.GeminiResp) *WordInfoResponse {
	return &WordInfoResponse{
		Response: resp.Response,
	}
}

type PracticeWithGemini struct {
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
	Topic      string `json:"topic"`
}

type StartPracticeWithGeminiResponse struct {
	Level     string   `json:"level"`
	Sentences []string `json:"sentences"`
}

type CheckAnswerPracticeWithGeminiResponse struct {
	GeneralComment string                     `json:"general_comment"`
	Results        []PracticeWithGeminiResult `json:"results"`
}

type PracticeWithGeminiResult struct {
	SentenceNumber   int    `json:"sentence_number"`
	YourVersion      string `json:"your_version"`
	Status           string `json:"status"`
	StatusLocalized  string `json:"status_localized"`
	TeacherComment   string `json:"teacher_comment"`
	IdealTranslation string `json:"ideal_translation"`
}

type WordListResp struct {
	SourceWord string
	TargetWord string
	Comment    string
}

type WordListReq struct {
	SourceLang string
	TargetLang string
	Level      string
	Topic      string
	UserTopic  string
}
type GemTranslationResp struct {
	SourceWord string
	TargetWord string
}

type GemTranslationReq struct {
	SourceWord string
	TargetWord string
	SourceLang string
	TargetLang string
}
