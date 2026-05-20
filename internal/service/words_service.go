package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/open-spaced-repetition/go-fsrs/v4"

	"synaply/external/deepl"
	"synaply/external/gemini"
	"synaply/internal/cache"
	"synaply/internal/models"
	"synaply/internal/repository"
	"synaply/internal/repository/modelsDB"
	"synaply/slogger"
)

var (
	fsrsParams    = fsrs.DefaultParam()
	fsrsScheduler = fsrs.NewFSRS(fsrsParams)
)

type WordsService interface {
	Create(ctx context.Context, req models.CreateReq, userID uuid.UUID) (*models.Response, error)
	LessonStart(ctx context.Context, userID uuid.UUID) (*models.Response, error)
	CheckAnswer(ctx context.Context, req models.AnswerReq, userID uuid.UUID) (bool, *models.Response, error)
	Finish(ctx context.Context, userID uuid.UUID) error
	Translate(ctx context.Context, req models.TranslateReq) (*models.TranslateResp, error)
	WordInfo(ctx context.Context, req gemini.WordInfoRequest) (*gemini.WordInfoResponse, error)
	StartPracticeWithGemini(ctx context.Context, req *gemini.PracticeWithGemini, userID uuid.UUID) (*gemini.StartPracticeWithGeminiResponse, error)
	CheckAnswerPracticeWithGemini(ctx context.Context, gemReq *gemini.PracticeWithGemini, userID uuid.UUID, translate string) (*gemini.CheckAnswerPracticeWithGeminiResponse, error)
	FinishPracticeWithGemini(ctx context.Context, userID uuid.UUID) error
	GetWordsList(ctx context.Context, req modelsDB.GetWordsListReq) ([]modelsDB.GetWordsListResp, int, error)
	DeleteWord(ctx context.Context, wordID string, userID uuid.UUID) error
	DeleteAllWords(ctx context.Context, userID uuid.UUID) error
	UpdateWordFields(ctx context.Context, req modelsDB.UpdateWordReq, userID uuid.UUID) error
	WordList(ctx context.Context, user *models.UserResponse, req models.WordListReq) ([]models.WordListResp, error)
	CreateBatch(ctx context.Context, req models.CreateBatchReq, userID uuid.UUID) error
	ImportWords(ctx context.Context, file []byte, fileName string, userID uuid.UUID, sourceLang, targetLang string) error
	GetProgressStats(ctx context.Context, userID uuid.UUID) (*models.ProgressStats, error)

	CreatePublicWordList(ctx context.Context, req models.CreatePublicWordListRequest, userID uuid.UUID) (uuid.UUID, error)
	GetPublicWordLists(ctx context.Context, sourceLang, targetLang, level string) ([]modelsDB.PublicWordList, error)
	GetPublicWordListByID(ctx context.Context, listID uuid.UUID) (*modelsDB.PublicWordListDetail, error)
	UpdatePublicWordList(ctx context.Context, listID uuid.UUID, req models.CreatePublicWordListRequest, userID uuid.UUID) error
	AddPublicListToUser(ctx context.Context, listID uuid.UUID, userID uuid.UUID) error

	CreatePlaylist(ctx context.Context, req models.CreatePlaylistRequest, userID uuid.UUID) (uuid.UUID, error)
	GetPlaylists(ctx context.Context) ([]modelsDB.Playlist, error)
	GetPlaylistByID(ctx context.Context, playlistID uuid.UUID) (*modelsDB.PlaylistDetail, error)
	UpdatePlaylist(ctx context.Context, playlistID uuid.UUID, req models.CreatePlaylistRequest, userID uuid.UUID) error
	DeletePlaylist(ctx context.Context, playlistID uuid.UUID, userID uuid.UUID) error
}

type wordsService struct {
	repo  repository.WordsRepository
	cache cache.CacheRepositoryI
	deepl deepl.ServiceI
	wg    *sync.WaitGroup
	gem   gemini.Service
}

func NewWordsService(repo repository.WordsRepository, cache cache.CacheRepositoryI, deepl deepl.ServiceI, wg *sync.WaitGroup, gem gemini.Service) *wordsService {
	return &wordsService{repo: repo, cache: cache, deepl: deepl, wg: wg, gem: gem}
}

func (s *wordsService) Translate(ctx context.Context, req models.TranslateReq) (*models.TranslateResp, error) {
	sourceWord := strings.TrimSpace(req.SourceWord)
	targetWord := strings.TrimSpace(req.TargetWord)
	if sourceWord == "" && targetWord == "" {
		return nil, errors.New("source or target word is required")
	}

	gemReq := &gemini.GemTranslationReq{
		SourceWord: sourceWord,
		TargetWord: targetWord,
		SourceLang: req.SourceLang,
		TargetLang: req.TargetLang,
	}
	resp, err := s.gem.WordTranslate(ctx, gemReq)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("translation response is empty")
	}

	resp.SourceWord = strings.TrimSpace(resp.SourceWord)
	resp.TargetWord = strings.TrimSpace(resp.TargetWord)
	if resp.SourceWord == "" {
		resp.SourceWord = sourceWord
	}
	if resp.TargetWord == "" {
		resp.TargetWord = targetWord
	}
	if sourceWord != "" && resp.TargetWord == "" {
		return nil, errors.New("target word translation is empty")
	}
	if sourceWord == "" && targetWord != "" && resp.SourceWord == "" {
		return nil, errors.New("source word translation is empty")
	}

	return &models.TranslateResp{
		ID:         req.ID,
		SourceWord: resp.SourceWord,
		TargetWord: resp.TargetWord,
	}, nil

}

func (s *wordsService) Create(ctx context.Context, req models.CreateReq, userID uuid.UUID) (*models.Response, error) {
	sourceWord := strings.ToLower(strings.TrimSpace(req.SourceWord))
	targetWord := strings.ToLower(strings.TrimSpace(req.TargetWord))

	req.SourceWord = sourceWord
	req.TargetWord = targetWord
	resp, err := s.repo.Create(ctx, models.CreateReqToDB(req, userID))
	if err != nil {
		return nil, err
	}
	return models.DBtoResponse(resp), nil
}

func (s *wordsService) LessonStart(ctx context.Context, userID uuid.UUID) (*models.Response, error) {
	key := models.CacheKey(userID)
	slogger.Log.DebugContext(ctx, "LessonStart is started ")

	nextWord, err := s.GetNextWordFromCache(ctx, userID)
	if err == nil {
		nextWordLog := models.LessonWordToResponse(nextWord)
		slogger.Log.DebugContext(ctx, "answer from GetNextWordFromCache", "next_word", nextWordLog)
		return models.LessonWordToResponse(nextWord), nil
	}
	slogger.Log.ErrorContext(ctx, "failed to get next word from cache", "error", err)

	wordsDB, err := s.repo.GetLessonWords(ctx, userID)
	if err != nil {
		return nil, err
	}

	LessonWords := models.LessonWordsDBtoLessonWords(wordsDB)
	val, err := json.Marshal(LessonWords)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal lesson: %w", err)
	}

	err = s.cache.Set(ctx, key, val, 15*time.Minute)
	if err != nil {
		slogger.Log.ErrorContext(ctx, "failed to cache lesson", "key", key, "error", err)
		return nil, fmt.Errorf("failed to set lesson: %w", err)
	}

	id, err := FindMinIdx(LessonWords)
	if err != nil {
		return nil, fmt.Errorf("lesson is empty or all words locked: %w", err)
	}

	nextWordDB := LessonWords[id]
	resp := models.LessonWordToResponse(&nextWordDB)
	slogger.Log.DebugContext(ctx, "successfully created lesson", "next word:", nextWordDB)

	return resp, nil
}

func (s *wordsService) CheckAnswer(ctx context.Context, req models.AnswerReq, userID uuid.UUID) (bool, *models.Response, error) {
	key := models.CacheKey(userID)
	var isCorrect bool
	lesson := make(map[string]models.Lesson)
	var word models.Lesson

	// 1. Достаем урок из кэша или БД
	data, err := s.cache.Get(ctx, key)
	if err == nil {
		if err := json.Unmarshal([]byte(data), &lesson); err != nil {
			return isCorrect, nil, fmt.Errorf("failed to unmarshal lesson: %w", err)
		}
		word = lesson[req.ID]
		if word.ID == uuid.Nil {
			wordDB, err := s.repo.GetWordByID(ctx, req.ID)
			if err != nil {
				return isCorrect, nil, err
			}
			word = models.LessonDBToLesson(wordDB)
		}
	} else {
		slogger.Log.DebugContext(ctx, "failed to get lesson from cache for checking", "key", key)
		wordDB, err := s.repo.GetWordByID(ctx, req.ID)
		if err != nil {
			return isCorrect, nil, err
		}
		word = models.LessonDBToLesson(wordDB)
	}

	card := fsrs.Card{
		Due:           word.Due,
		Stability:     word.Stability,
		Difficulty:    word.Difficulty,
		ElapsedDays:   word.ElapsedDays,
		ScheduledDays: word.ScheduledDays,
		Reps:          word.Reps,
		Lapses:        word.Lapses,
		State:         word.State,
		LastReview:    word.LastReview,
	}

	isMistake := !strings.EqualFold(req.TargetWord, word.TargetWord)
	isCorrect = !isMistake

	var newCard fsrs.Card

	if word.FSRSLocked {
		if isMistake {
			schedulingInfo := fsrsScheduler.Next(card, time.Now().UTC(), fsrs.Again)
			newCard = schedulingInfo.Card

			word.FSRSLocked = false
			word.MistakesInSession++
		} else {
			newCard = card
		}
	} else {
		if isMistake {
			if word.MistakesInSession == 0 {
				schedulingInfo := fsrsScheduler.Next(card, time.Now().UTC(), fsrs.Again)
				newCard = schedulingInfo.Card
			} else {
				newCard = card
			}
			word.MistakesInSession++

		} else {
			var rating fsrs.Rating
			if word.MistakesInSession == 0 {
				if word.State == 0 {
					rating = fsrs.Easy
				} else {
					rating = fsrs.Good
				}
			} else {
				rating = fsrs.Hard
			}

			schedulingInfo := fsrsScheduler.Next(card, time.Now().UTC(), rating)
			newCard = schedulingInfo.Card

			word.FSRSLocked = true
		}
	}

	lesson[req.ID] = models.Lesson{
		ID:         word.ID,
		SourceWord: word.SourceWord,
		TargetWord: word.TargetWord,
		Comment:    word.Comment,
		SourceLang: word.SourceLang,
		TargetLang: word.TargetLang,

		Due:           newCard.Due,
		Stability:     newCard.Stability,
		Difficulty:    newCard.Difficulty,
		ElapsedDays:   newCard.ElapsedDays,
		ScheduledDays: newCard.ScheduledDays,
		Reps:          newCard.Reps,
		Lapses:        newCard.Lapses,
		State:         newCard.State,
		LastReview:    newCard.LastReview,

		Index:             word.Index + 1,
		MistakesInSession: word.MistakesInSession,
		FSRSLocked:        word.FSRSLocked,
	}

	val, err := json.Marshal(lesson)
	if err != nil {
		return isCorrect, nil, fmt.Errorf("failed to marshal lesson: %w", err)
	}
	err = s.cache.Set(ctx, key, val, 15*time.Minute)
	if err != nil {
		slogger.Log.ErrorContext(ctx, "failed to cache lesson", "key", key, "error", err)
	}

	bgCtx := context.WithoutCancel(ctx)
	s.wg.Add(1)
	updatedWord := lesson[req.ID]
	go func(wordToSave models.Lesson) {
		defer s.wg.Done()
		timeoutCtx, cancel := context.WithTimeout(bgCtx, 5*time.Second)
		defer cancel()

		wordDB := models.LessonToLessonDB(&wordToSave)
		err := s.repo.UpdateWord(timeoutCtx, wordDB)
		if err != nil {
			slogger.Log.ErrorContext(timeoutCtx, "failed to update lesson", "key", key, "error", err)
		}
	}(updatedWord)

	if isCorrect {
		nextWord, err := s.GetNextWordFromCache(ctx, userID)
		if err != nil {
			if errors.Is(err, models.ErrNoWordsForLesson) {
				return isCorrect, nil, nil
			}
			return isCorrect, nil, err
		}
		resp := models.LessonWordToResponse(nextWord)
		slogger.Log.DebugContext(ctx, "successfully created next word", "next word:", nextWord)
		return isCorrect, resp, nil
	}

	return isCorrect, models.LessonWordToResponse(&word), nil
}

func (s *wordsService) GetNextWordFromCache(ctx context.Context, userID uuid.UUID) (*models.Lesson, error) {
	key := models.CacheKey(userID)
	data, err := s.cache.Get(ctx, key)
	if errors.Is(err, cache.ErrCacheMiss) {
		slogger.Log.DebugContext(ctx, "GetNextWordFromCache is cache miss")
	}
	if err != nil {
		return nil, err
	}

	lesson := make(map[string]models.Lesson)
	if err := json.Unmarshal([]byte(data), &lesson); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lesson: %w", err)
	}
	slogger.Log.DebugContext(ctx, "GetNextWordFromCache lesson from cache", "lesson", lesson)

	id, err := FindMinIdx(lesson)
	if err != nil {
		slogger.Log.DebugContext(ctx, "GetNextWordFromCache: lesson finished (all words answered or locked)")
		return nil, err
	}
	word, exists := lesson[id]
	if !exists {
		errNotFound := errors.New("word not found in cache payload")
		slogger.Log.DebugContext(ctx, "GetNextWordFromCache is not exists", "error", errNotFound)
		return nil, errNotFound
	}

	slogger.Log.DebugContext(ctx, "GetNextWordFromCache is finished ", "word:", word)

	return &word, nil
}

func FindMinIdx(lesson map[string]models.Lesson) (string, error) {
	minIdx := math.MaxInt32
	var nextWordID string
	found := false
	for key, word := range lesson {
		if word.FSRSLocked {
			continue
		}
		if word.Index < minIdx {
			minIdx = word.Index
			nextWordID = key
			found = true
		}
	}

	if !found {
		return "", errors.New("no unlocked words ready for review")
	}

	return nextWordID, nil
}

func (s *wordsService) Finish(ctx context.Context, userID uuid.UUID) error {

	err := s.cache.Del(ctx, models.CacheKey(userID))
	if err != nil {
		slogger.Log.ErrorContext(ctx, "failed to delete cache lesson", "error", err)
	}

	slogger.Log.DebugContext(ctx, "lesson finished and cache cleared", "userId", userID)
	return nil
}

func (s *wordsService) WordInfo(ctx context.Context, req gemini.WordInfoRequest) (*gemini.WordInfoResponse, error) {

	searchReq := &modelsDB.GeminiReq{
		SourceLang: req.SourceLang,
		TargetLang: req.TargetLang,
		SourceWord: req.SourceWord,
		TargetWord: req.TargetWord,
	}

	dbResp, err := s.repo.GetWordInfo(ctx, searchReq)
	if err == nil {
		slogger.Log.DebugContext(ctx, "success find in db", "resp", dbResp)
		return &gemini.WordInfoResponse{Response: dbResp.Response}, nil
	}

	respString, err := s.gem.WordInfo(ctx, req)
	if err != nil {
		slogger.Log.ErrorContext(ctx, "Gemini failed", "error", err)
		return nil, err
	}
	slogger.Log.DebugContext(ctx, "3. WordInfo from gemini", "word", respString)
	detachedCtx := context.WithoutCancel(ctx)
	s.wg.Add(1)
	go func(geminiAnswer string) {
		defer s.wg.Done()
		timeoutCtx, cancel := context.WithTimeout(detachedCtx, 5*time.Second)
		defer cancel()

		saveReq := modelsDB.GeminiReq{
			SourceLang: req.SourceLang,
			TargetLang: req.TargetLang,
			SourceWord: req.SourceWord,
			TargetWord: req.TargetWord,
			Response:   geminiAnswer,
		}

		if err := s.repo.SetWordInfo(timeoutCtx, saveReq); err != nil {
			slogger.Log.ErrorContext(timeoutCtx, "failed save to db:", "error", err)
		} else {
			slogger.Log.DebugContext(timeoutCtx, "success save to db", "resp", saveReq)
		}
	}(respString)

	return &gemini.WordInfoResponse{Response: respString}, nil
}

func (s *wordsService) StartPracticeWithGemini(ctx context.Context, req *gemini.PracticeWithGemini, userID uuid.UUID) (*gemini.StartPracticeWithGeminiResponse, error) {
	temp := &modelsDB.WordsForGeminiReq{
		UserID:     userID,
		SourceLang: req.SourceLang,
		TargetLang: req.TargetLang,
	}

	wordList, err := s.repo.GetWordsForGemini(ctx, temp)
	if err != nil {
		return nil, err
	}
	var wordsListString strings.Builder
	for _, word := range wordList {
		line := fmt.Sprintf("%s - %s\n", word.SourceWord, word.TargetWord)
		wordsListString.WriteString(line)
	}

	resp, err := s.gem.StartPracticeWithGemini(ctx, req, wordsListString.String())
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("PracticeWithGemini:%s", userID)

	// For the cache, we save sentences as a single string so that CheckAnswer can retrieve them
	var sentencesBuilder strings.Builder
	for i, s := range resp.Sentences {
		sentencesBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, s))
	}

	data := &models.PracticeWithGeminiCache{
		TaskTranslate: sentencesBuilder.String(),
		Topic:         req.Topic,
	}
	val, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = s.cache.Set(ctx, key, val, 15*time.Minute)
	if errors.Is(err, cache.ErrCacheMiss) {
		slogger.Log.DebugContext(ctx, "GetNextWordFromCache is cache miss")
	}
	if err != nil {
		return nil, err
	}

	return resp, nil
}
func (s *wordsService) CheckAnswerPracticeWithGemini(ctx context.Context, gemReq *gemini.PracticeWithGemini, userID uuid.UUID, userTranslate string) (*gemini.CheckAnswerPracticeWithGeminiResponse, error) {
	key := fmt.Sprintf("PracticeWithGemini:%s", userID)
	data, err := s.cache.Get(ctx, key)
	if errors.Is(err, cache.ErrCacheMiss) {
		slogger.Log.DebugContext(ctx, "GetNextWordFromCache is cache miss")
	}
	if err != nil {
		return nil, err
	}

	var practiceWithGeminiCache models.PracticeWithGeminiCache
	if err := json.Unmarshal([]byte(data), &practiceWithGeminiCache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lesson: %w", err)
	}

	var taskWithTranslate strings.Builder
	taskWithTranslate.WriteString("5 sentence for userTranslate:\n")
	taskWithTranslate.WriteString(practiceWithGeminiCache.TaskTranslate)
	taskWithTranslate.WriteString("answer from user:\n")
	taskWithTranslate.WriteString(userTranslate)

	gemReq.Topic = practiceWithGeminiCache.Topic
	resp, err := s.gem.CheckAnswerPracticeWithGemini(ctx, gemReq, taskWithTranslate.String())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *wordsService) FinishPracticeWithGemini(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf("PracticeWithGemini:%s", userID)
	err := s.cache.Del(ctx, key)
	if err != nil {
		return err
	}
	return nil
}

func (s *wordsService) GetWordsList(ctx context.Context, req modelsDB.GetWordsListReq) ([]modelsDB.GetWordsListResp, int, error) {
	return s.repo.GetWordsList(ctx, req)
}

func (s *wordsService) DeleteWord(ctx context.Context, wordID string, userID uuid.UUID) error {
	return s.repo.DeleteWord(ctx, wordID, userID)
}

func (s *wordsService) DeleteAllWords(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteAllWords(ctx, userID)
}

func (s *wordsService) UpdateWordFields(ctx context.Context, req modelsDB.UpdateWordReq, userID uuid.UUID) error {
	return s.repo.UpdateWordFields(ctx, req, userID)
}

func (s *wordsService) WordList(ctx context.Context, user *models.UserResponse, req models.WordListReq) ([]models.WordListResp, error) {
	topicForCache := req.Topic
	if topicForCache == "" {
		topicForCache = req.UserTopic
	}

	// 1. Trying to get from DB
	cached, err := s.repo.GetGeminiWordList(ctx, user.SourceLang, user.TargetLang, req.Level, topicForCache)
	if err == nil && cached != nil {
		var wordListResp []models.WordListResp
		if err := json.Unmarshal(cached.Response, &wordListResp); err == nil {
			slogger.Log.DebugContext(ctx, "WordList found in DB cache", "topic", topicForCache)
			return wordListResp, nil
		}
		slogger.Log.ErrorContext(ctx, "Failed to unmarshal cached word list", "error", err)
	}

	// 2. If not in DB, go to Gemini
	wordListRespGem, err := s.gem.WordList(ctx, models.WordListReqToGemWordListReq(req))
	if err != nil {
		return nil, err
	}

	wordListResp := make([]models.WordListResp, len(wordListRespGem))
	for i, word := range wordListRespGem {
		wordListResp[i] = models.WordListRespGemToWordListResp(word)
	}

	// 3. Save to DB for future use
	respJSON, err := json.Marshal(wordListResp)
	if err == nil {
		errSave := s.repo.SaveGeminiWordList(ctx, modelsDB.GeminiWordList{
			SourceLang: user.SourceLang,
			TargetLang: user.TargetLang,
			Level:      req.Level,
			Topic:      topicForCache,
			Response:   respJSON,
		})
		if errSave != nil {
			slogger.Log.ErrorContext(ctx, "Failed to save word list to DB cache", "error", errSave)
		}
	} else {
		slogger.Log.ErrorContext(ctx, "Failed to marshal word list for cache", "error", err)
	}

	return wordListResp, nil
}

// Add to WordsService interface:
// CreateBatch(ctx context.Context, reqs []models.CreateReq, userID uuid.UUID) error

func (s *wordsService) CreateBatch(ctx context.Context, req models.CreateBatchReq, userID uuid.UUID) error {
	var dbReqs []modelsDB.CreateReq

	for _, word := range req.Words {
		sourceWord := strings.ToLower(strings.TrimSpace(word.SourceWord))
		targetWord := strings.ToLower(strings.TrimSpace(word.TargetWord))

		dbReqs = append(dbReqs, modelsDB.CreateReq{
			UserID:     userID,
			SourceLang: req.SourceLang,
			TargetLang: req.TargetLang,
			SourceWord: sourceWord,
			TargetWord: targetWord,
			Comment:    word.Comment,
		})
	}

	if len(dbReqs) == 0 {
		return nil // Nothing to save
	}

	return s.repo.CreateBatch(ctx, dbReqs)
}

func (s *wordsService) ImportWords(ctx context.Context, file []byte, fileName string, userID uuid.UUID, sourceLang, targetLang string) error {
	var words []models.WordListResp

	if strings.HasSuffix(fileName, ".json") {
		var req models.CreateBatchReq
		if err := json.Unmarshal(file, &req); err == nil && len(req.Words) > 0 {
			words = req.Words
			// Update languages from request if provided
			if req.SourceLang != "" {
				sourceLang = req.SourceLang
			}
			if req.TargetLang != "" {
				targetLang = req.TargetLang
			}
		} else {
			// Try to unmarshal as simple array of words
			if err := json.Unmarshal(file, &words); err != nil {
				return fmt.Errorf("failed to parse JSON: %w", err)
			}
		}
	} else if strings.HasSuffix(fileName, ".csv") {
		reader := csv.NewReader(bytes.NewReader(file))
		// Handle potential different delimiters or BOM could be added here
		for {
			record, err := reader.Read()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read CSV: %w", err)
			}

			if len(record) < 2 {
				continue // Skip invalid lines
			}

			word := models.WordListResp{
				SourceWord: record[0],
				TargetWord: record[1],
			}
			if len(record) >= 3 {
				word.Comment = record[2]
			}
			words = append(words, word)
		}
	} else {
		return errors.New("unsupported file format")
	}

	if len(words) == 0 {
		return errors.New("no words found in file")
	}

	return s.CreateBatch(ctx, models.CreateBatchReq{
		SourceLang: sourceLang,
		TargetLang: targetLang,
		Words:      words,
	}, userID)
}

func (s *wordsService) GetProgressStats(ctx context.Context, userID uuid.UUID) (*models.ProgressStats, error) {
	return s.repo.GetProgressStats(ctx, userID)
}

func (s *wordsService) CreatePublicWordList(ctx context.Context, req models.CreatePublicWordListRequest, userID uuid.UUID) (uuid.UUID, error) {
	if req.Level == "" {
		return uuid.Nil, errors.New("level is required")
	}
	list := modelsDB.PublicWordList{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		SourceLang:  req.SourceLang,
		TargetLang:  req.TargetLang,
		Level:       req.Level,
	}

	items := make([]modelsDB.PublicWordListItem, 0, len(req.Words))
	for _, w := range req.Words {
		items = append(items, modelsDB.PublicWordListItem{
			SourceWord: w.SourceWord,
			TargetWord: w.TargetWord,
			Comment:    w.Comment,
		})
	}

	return s.repo.CreatePublicWordList(ctx, list, items)
}

func (s *wordsService) GetPublicWordLists(ctx context.Context, sourceLang, targetLang, level string) ([]modelsDB.PublicWordList, error) {
	return s.repo.GetPublicWordLists(ctx, sourceLang, targetLang, level)
}

func (s *wordsService) GetPublicWordListByID(ctx context.Context, listID uuid.UUID) (*modelsDB.PublicWordListDetail, error) {
	return s.repo.GetPublicWordListByID(ctx, listID)
}

func (s *wordsService) UpdatePublicWordList(ctx context.Context, listID uuid.UUID, req models.CreatePublicWordListRequest, userID uuid.UUID) error {
	// Check if list exists and user is the owner
	detail, err := s.repo.GetPublicWordListByID(ctx, listID)
	if err != nil {
		return err
	}

	if detail.UserID != userID {
		return errors.New("unauthorized: you can only edit your own lists")
	}

	list := modelsDB.PublicWordList{
		ID:          listID,
		Title:       req.Title,
		Description: req.Description,
		Level:       req.Level,
	}

	items := make([]modelsDB.PublicWordListItem, 0, len(req.Words))
	for _, w := range req.Words {
		items = append(items, modelsDB.PublicWordListItem{
			SourceWord: w.SourceWord,
			TargetWord: w.TargetWord,
			Comment:    w.Comment,
		})
	}

	return s.repo.UpdatePublicWordList(ctx, list, items)
}

func (s *wordsService) AddPublicListToUser(ctx context.Context, listID uuid.UUID, userID uuid.UUID) error {
	detail, err := s.repo.GetPublicWordListByID(ctx, listID)
	if err != nil {
		return err
	}

	reqs := make([]modelsDB.CreateReq, 0, len(detail.Items))
	for _, item := range detail.Items {
		sourceWord := strings.ToLower(strings.TrimSpace(item.SourceWord))
		targetWord := strings.ToLower(strings.TrimSpace(item.TargetWord))

		reqs = append(reqs, modelsDB.CreateReq{
			UserID:     userID,
			SourceLang: detail.SourceLang,
			TargetLang: detail.TargetLang,
			SourceWord: sourceWord,
			TargetWord: targetWord,
			Comment:    item.Comment,
		})
	}

	return s.repo.CreateBatch(ctx, reqs)
}

func (s *wordsService) CreatePlaylist(ctx context.Context, req models.CreatePlaylistRequest, userID uuid.UUID) (uuid.UUID, error) {
	playlist := modelsDB.Playlist{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
	}
	return s.repo.CreatePlaylist(ctx, playlist, req.ListIDs)
}

func (s *wordsService) GetPlaylists(ctx context.Context) ([]modelsDB.Playlist, error) {
	return s.repo.GetPlaylists(ctx)
}

func (s *wordsService) GetPlaylistByID(ctx context.Context, playlistID uuid.UUID) (*modelsDB.PlaylistDetail, error) {
	return s.repo.GetPlaylistByID(ctx, playlistID)
}

func (s *wordsService) UpdatePlaylist(ctx context.Context, playlistID uuid.UUID, req models.CreatePlaylistRequest, userID uuid.UUID) error {
	playlist := modelsDB.Playlist{
		ID:          playlistID,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
	}
	return s.repo.UpdatePlaylist(ctx, playlist, req.ListIDs)
}

func (s *wordsService) DeletePlaylist(ctx context.Context, playlistID uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeletePlaylist(ctx, playlistID, userID)
}
