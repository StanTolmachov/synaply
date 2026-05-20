package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/open-spaced-repetition/go-fsrs/v4"

	"synaply/external/gemini"
	"synaply/internal/repository/modelsDB"
)

type Word struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	SourceLang string
	TargetLang string
	SourceWord string
	TargetWord string
	Comment    string
	// FSRS fields
	Due           time.Time
	Stability     float64
	Difficulty    float64
	ElapsedDays   uint64
	ScheduledDays uint64
	Reps          uint64
	Lapses        uint64
	State         int8
	LastReview    time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type TranslateReq struct {
	ID         uuid.UUID `json:"-"` // ID is taken from context, so we forbid taking it from JSON
	SourceLang string    `json:"source_lang"`
	TargetLang string    `json:"target_lang"`
	SourceWord string    `json:"source_word"`
	TargetWord string    `json:"target_word"`
}
type TranslateResp struct {
	ID         uuid.UUID `json:"id"`
	SourceWord string    `json:"source_word"`
	TargetWord string    `json:"target_word"`
}
type CreateReq struct {
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
	SourceWord string `json:"source_word"`
	TargetWord string `json:"target_word"`
	Comment    string `json:"comment"`
}

func CreateReqToDB(req CreateReq, userID uuid.UUID) modelsDB.CreateReq {
	return modelsDB.CreateReq{
		UserID:     userID,
		SourceLang: req.SourceLang,
		TargetLang: req.TargetLang,
		SourceWord: req.SourceWord,
		TargetWord: req.TargetWord,
		Comment:    req.Comment,
	}
}

func DBtoResponse(word *modelsDB.Word) *Response {
	return &Response{
		ID:         word.UserID.String(),
		SourceWord: word.SourceWord,
		TargetWord: word.TargetWord,
		Comment:    word.Comment,
		SourceLang: word.SourceLang,
		TargetLang: word.TargetLang,
	}
}

type Response struct {
	ID         string `json:"id"`
	SourceWord string `json:"source_word"`
	TargetWord string `json:"target_word"`
	Comment    string `json:"comment"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

type Lesson struct {
	ID         uuid.UUID
	SourceWord string
	TargetWord string
	Comment    string
	SourceLang string
	TargetLang string
	// FSRS fields
	Due           time.Time
	Stability     float64
	Difficulty    float64
	ElapsedDays   uint64
	ScheduledDays uint64
	Reps          uint64
	Lapses        uint64
	State         fsrs.State
	LastReview    time.Time
	//index for order
	Index             int
	MistakesInSession int
	FSRSLocked        bool
}

func LessonToLessonDB(lesson *Lesson) modelsDB.LessonDB {
	var lastReviewDB *time.Time
	if !lesson.LastReview.IsZero() {
		lastReviewDB = &lesson.LastReview
	}
	return modelsDB.LessonDB{
		ID:            lesson.ID,
		SourceWord:    lesson.SourceWord,
		TargetWord:    lesson.TargetWord,
		Comment:       lesson.Comment,
		SourceLang:    lesson.SourceLang,
		TargetLang:    lesson.TargetLang,
		Due:           lesson.Due,
		Stability:     lesson.Stability,
		Difficulty:    lesson.Difficulty,
		ElapsedDays:   lesson.ElapsedDays,
		ScheduledDays: lesson.ScheduledDays,
		Reps:          lesson.Reps,
		Lapses:        lesson.Lapses,
		State:         int(lesson.State),
		LastReview:    lastReviewDB,
	}
}

func LessonDBToLesson(lessonDB *modelsDB.LessonDB) Lesson {
	var lastReview time.Time
	if lessonDB.LastReview != nil {
		lastReview = *lessonDB.LastReview
	}
	return Lesson{
		ID:            lessonDB.ID,
		SourceWord:    lessonDB.SourceWord,
		TargetWord:    lessonDB.TargetWord,
		Comment:       lessonDB.Comment,
		SourceLang:    lessonDB.SourceLang,
		TargetLang:    lessonDB.TargetLang,
		Due:           lessonDB.Due,
		Stability:     lessonDB.Stability,
		Difficulty:    lessonDB.Difficulty,
		ElapsedDays:   lessonDB.ElapsedDays,
		ScheduledDays: lessonDB.ScheduledDays,
		Reps:          lessonDB.Reps,
		Lapses:        lessonDB.Lapses,
		State:         fsrs.State(lessonDB.State),
		LastReview:    lastReview,
		Index:         0,
	}
}

func LessonWordToResponse(word *Lesson) *Response {
	resp := &Response{
		ID:         word.ID.String(),
		SourceWord: word.SourceWord,
		TargetWord: word.TargetWord,
		Comment:    word.Comment,
		SourceLang: word.SourceLang,
		TargetLang: word.TargetLang,
	}
	return resp
}

type AnswerReq struct {
	ID         string `json:"id"`
	TargetWord string `json:"target_word"`
}

func CacheKey(userID uuid.UUID) string {
	return fmt.Sprintf("lesson_for_%s", userID)
}

func LessonWordsDBtoLessonWords(wordsDB []modelsDB.LessonDB) map[string]Lesson {
	var lessonWords = make(map[string]Lesson)
	for _, word := range wordsDB {
		lessonWords[word.ID.String()] = LessonDBToLesson(&word)
	}
	return lessonWords
}

var (
	ErrWordAlreadyExists = errors.New("word already exists")
	ErrDBTimeout         = errors.New("db timeout")
	ErrNoWordsForLesson  = errors.New("no words found for lesson")
)

type StartLessonResponse struct {
	TotalCorrect int64     `json:"total_correct"`
	Word         *Response `json:"word"`
}

type CheckAnswerResponse struct {
	IsCorrect    bool      `json:"is_correct"`
	NextWord     *Response `json:"next_word"`
	TotalCorrect int64     `json:"total_correct"`
}

type UserTranslateResponse struct {
	Translation string `json:"translation"`
}

type PracticeWithGeminiCache struct {
	TaskTranslate string `json:"task_translate"`
	Topic         string `json:"topic"`
}

type Topic struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type TopicResponse struct {
	Total int     `json:"total"`
	Topic []Topic `json:"topic"`
}

var topicList = []Topic{
	{Slug: "identity", Title: "Identity & Personal Info"},
	{Slug: "family", Title: "Family & Relationships"},
	{Slug: "home", Title: "Home & Daily Life"},
	{Slug: "food", Title: "Food & Drinks"},
	{Slug: "shopping", Title: "Shopping & Money"},
	{Slug: "health", Title: "Health & Body"},
	{Slug: "education", Title: "Education & Learning"},
	{Slug: "work", Title: "Work & Career"},
	{Slug: "leisure", Title: "Leisure & Sports"},
	{Slug: "travel", Title: "Travel & Transport"},
	{Slug: "city", Title: "City & Infrastructure"},
	{Slug: "nature", Title: "Nature & Environment"},
	{Slug: "technology", Title: "Technology & Media"},
	{Slug: "art", Title: "Art, Culture & Ideas"},
}

func GetTopicList() []Topic {
	return topicList
}

type ProficiencyLevel string

const (
	LevelA1 ProficiencyLevel = "A1"
	LevelA2 ProficiencyLevel = "A2"
	LevelB1 ProficiencyLevel = "B1"
	LevelB2 ProficiencyLevel = "B2"
	LevelC1 ProficiencyLevel = "C1"
	LevelC2 ProficiencyLevel = "C2"
)

type WordListReq struct {
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
	Level      string `json:"level"`
	Topic      string `json:"topic"`
	UserTopic  string `json:"user_topic"`
}

func WordListReqToGemWordListReq(req WordListReq) gemini.WordListReq {
	return gemini.WordListReq{
		SourceLang: req.SourceLang,
		TargetLang: req.TargetLang,
		Level:      req.Level,
		Topic:      req.Topic,
		UserTopic:  req.UserTopic,
	}
}

type WordListResp struct {
	SourceWord string `json:"source_word"`
	TargetWord string `json:"target_word"`
	Comment    string `json:"comment"`
}

type CreateBatchReq struct {
	SourceLang string         `json:"source_lang"`
	TargetLang string         `json:"target_lang"`
	Words      []WordListResp `json:"words"`
}

func WordListRespGemToWordListResp(req gemini.WordListResp) WordListResp {
	return WordListResp{
		SourceWord: req.SourceWord,
		TargetWord: req.TargetWord,
		Comment:    req.Comment,
	}
}

type ProgressStats struct {
	New        int64 `json:"new"`
	Learning   int64 `json:"learning"`
	Review     int64 `json:"review"`
	Relearning int64 `json:"relearning"`
}

type CreatePublicWordListRequest struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	SourceLang  string         `json:"source_lang"`
	TargetLang  string         `json:"target_lang"`
	Level       string         `json:"level"`
	Words       []WordListResp `json:"words"`
}

type CreatePlaylistRequest struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	ListIDs     []uuid.UUID `json:"list_ids"`
}
