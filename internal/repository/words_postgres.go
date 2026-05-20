package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"synaply/internal/models"
	"synaply/internal/repository/modelsDB"
	"synaply/slogger"
)

type wordsPostgres struct {
	db *Postgres
}

func NewWordsPostgres(db *Postgres) WordsRepository {
	return &wordsPostgres{db: db}
}

func (p *wordsPostgres) Create(ctx context.Context, req modelsDB.CreateReq) (*modelsDB.Word, error) {
	slogger.Log.DebugContext(ctx, "create is started")

	query := `
	insert into words (user_id, source_lang, target_lang, source_word, target_word, comment) 
	values ($1, $2, $3, $4, $5, $6)
	returning id, user_id, source_lang, target_lang, source_word, target_word, comment
	`
	var resp modelsDB.Word
	err := p.db.db.QueryRowxContext(ctx, query, req.UserID, req.SourceLang, req.TargetLang, req.SourceWord, req.TargetWord, req.Comment).StructScan(&resp)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%w: %s", models.ErrWordAlreadyExists, err)
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("%w: %v", models.ErrDBTimeout, err)
		}
		return nil, fmt.Errorf("unexpected error creating word: %w", err)
	}
	slogger.Log.DebugContext(ctx, "Repo NewWord response", "resp", resp)
	return &resp, nil
}

func (p *wordsPostgres) GetLessonWords(ctx context.Context, userID uuid.UUID) ([]modelsDB.LessonDB, error) {
	query := `
SELECT 
    id, source_word, target_word, comment, source_lang, target_lang, 
    due, stability, difficulty, elapsed_days, scheduled_days, 
    reps, lapses, state, last_review
FROM words
WHERE user_id = $1 
  AND (
      -- Кандидат 1: Слова на повторение, чье время УЖЕ ПРИШЛО (due <= сейчас)
      (state != 0 AND due <= NOW()) 
      OR 
      -- Кандидат 2: Абсолютно новые слова
      (state = 0)
  )
ORDER BY 
    -- МАГИЯ ЗДЕСЬ: Присваиваем приоритеты.
    -- Сначала берем слова на повторение (приоритет 1), затем новые (приоритет 2)
    CASE WHEN state != 0 THEN 1 ELSE 2 END ASC,
    
    -- Внутри группы повторения сортируем так, чтобы самые "просроченные" шли первыми
    due ASC,
    
    -- Внутри группы новых слов сортируем по дате добавления (старые добавленные первыми)
    created_at ASC
LIMIT 10;
	`

	//todo
	/*
		func (p *wordsPostgres) GetLessonWords(ctx context.Context, userID uuid.UUID) ([]modelsDB.LessonDB, error) {
			query := `
			WITH ranked_words AS (
				-- Присваиваем ранг (порядковый номер) внутри каждой группы
				SELECT *,
				       ROW_NUMBER() OVER (
				           PARTITION BY state = 0
				           ORDER BY CASE WHEN state = 0 THEN created_at ELSE due END ASC
				       ) as rn
				FROM words
				WHERE user_id = $1
			),
			selection AS (
				-- Выбираем 3 новых (rn <= 3) и 7 сложных (rn <= 7)
				-- Но если в одной группе слов меньше, мы все равно ограничим общую выборку позже
				SELECT * FROM ranked_words WHERE (state = 0 AND rn <= 3)
				UNION ALL
				SELECT * FROM ranked_words WHERE (state != 0 AND rn <= 7)
			),
			filler AS (
				-- Добираем недостающие слова из общего списка тех, что не попали в основную выборку
				-- Сначала берем новые, потом сложные, или наоборот, в зависимости от приоритета
				SELECT * FROM ranked_words
				WHERE id NOT IN (SELECT id FROM selection)
				ORDER BY state = 0 DESC, created_at ASC
				LIMIT 10
			)
			-- Финальная выборка: объединяем и берем первые 10
			SELECT id, source_word, target_word, comment, source_lang, target_lang, due, stability, difficulty, elapsed_days, scheduled_days, reps, lapses, state, last_review
			FROM (
				SELECT * FROM selection
				UNION ALL
				SELECT * FROM filler
			) final
			LIMIT 10;
			`
			// ... остальной код метода
		}

	*/
	var resp []modelsDB.LessonDB
	err := p.db.db.SelectContext(ctx, &resp, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting lesson words: %w", err)
	}

	if len(resp) == 0 {
		return nil, models.ErrNoWordsForLesson
	}

	return resp, nil
}

func (p *wordsPostgres) GetWordByID(ctx context.Context, wordID string) (*modelsDB.LessonDB, error) {
	query := `
	select id, source_word, target_word, comment, source_lang, target_lang, due, stability, difficulty, elapsed_days, scheduled_days, reps, lapses, state, last_review
	from words
	where id = $1`

	var resp modelsDB.LessonDB
	err := p.db.db.QueryRowxContext(ctx, query, wordID).StructScan(&resp)
	if err != nil {
		return nil, fmt.Errorf("error getting word by id: %w", err)
	}
	return &resp, nil
}

func (p *wordsPostgres) Update(ctx context.Context, lesson map[string]modelsDB.LessonDB) error {
	query := `
update words
set comment = :comment,
    due = :due,
    stability = :stability,
    difficulty = :difficulty,
    elapsed_days = :elapsed_days,
    scheduled_days = :scheduled_days,
    reps = :reps,
    lapses = :lapses,
    state = :state,
    last_review = :last_review,
	updated_at = NOW()
where id = :id
`

	tx, err := p.db.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %s", err)
	}

	defer tx.Rollback()

	for _, word := range lesson {
		_, err = tx.NamedExecContext(ctx, query, word)
		if err != nil {
			return fmt.Errorf("error updating word: %s", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %s", err)
	}
	return nil

}

func (p *wordsPostgres) UpdateWord(ctx context.Context, word modelsDB.LessonDB) error {
	query := `
update words
set comment = :comment,
    due = :due,
    stability = :stability,
    difficulty = :difficulty,
    elapsed_days = :elapsed_days,
    scheduled_days = :scheduled_days,
    reps = :reps,
    lapses = :lapses,
    state = :state,
    last_review = :last_review,
	updated_at = NOW()
where id = :id
`

	_, err := p.db.db.NamedExecContext(ctx, query, word)
	if err != nil {
		return fmt.Errorf("error updating word: %s", err)
	}

	return nil

}

func (p *wordsPostgres) SetWordInfo(ctx context.Context, req modelsDB.GeminiReq) error {
	query := `
	INSERT INTO gemini (source_lang, target_lang, source_word, target_word, response)
	VALUES (:source_lang, :target_lang, :source_word, :target_word, :response)
	ON CONFLICT (source_lang, target_lang, source_word, target_word) 
	DO UPDATE SET response = EXCLUDED.response;
	`

	_, err := p.db.db.NamedExecContext(ctx, query, req)
	if err != nil {
		return fmt.Errorf("error saving word info: %w", err)
	}

	return nil
}

func (p *wordsPostgres) GetWordInfo(ctx context.Context, req *modelsDB.GeminiReq) (*modelsDB.GeminiResp, error) {
	query := `
        SELECT response
        FROM gemini
        WHERE LOWER(TRIM(source_lang)) = LOWER(TRIM($1)) 
          AND LOWER(TRIM(target_lang)) = LOWER(TRIM($2)) 
          AND LOWER(TRIM(source_word)) = LOWER(TRIM($3)) 
          AND LOWER(TRIM(target_word)) = LOWER(TRIM($4))
        LIMIT 1`
	var resp modelsDB.GeminiResp
	err := p.db.db.GetContext(ctx, &resp, query, req.SourceLang, req.TargetLang, req.SourceWord, req.TargetWord)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *wordsPostgres) GetWordsList(ctx context.Context, req modelsDB.GetWordsListReq) ([]modelsDB.GetWordsListResp, int, error) {
	var words []modelsDB.GetWordsListResp
	var total int

	search := "%" + req.Search + "%"

	countQuery := `SELECT count(*) FROM words WHERE user_id = $1 AND (source_word ILIKE $2 OR target_word ILIKE $2)`
	err := p.db.db.GetContext(ctx, &total, countQuery, req.UserID, search)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, source_word, target_word, comment 
              FROM words 
              WHERE user_id = $1 AND (source_word ILIKE $2 OR target_word ILIKE $2)
              ORDER BY created_at DESC 
              LIMIT $3 OFFSET $4`

	err = p.db.db.SelectContext(ctx, &words, query, req.UserID, search, req.Limit, req.Offset)
	if err != nil {
		return nil, 0, err
	}

	return words, total, nil
}

func (p *wordsPostgres) DeleteWord(ctx context.Context, wordID string, userID uuid.UUID) error {
	query := `DELETE FROM words WHERE id = $1 AND user_id = $2`
	_, err := p.db.db.ExecContext(ctx, query, wordID, userID)
	return err
}

func (p *wordsPostgres) DeleteAllWords(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM words WHERE user_id = $1`
	_, err := p.db.db.ExecContext(ctx, query, userID)
	return err
}

func (p *wordsPostgres) UpdateWordFields(ctx context.Context, req modelsDB.UpdateWordReq, userID uuid.UUID) error {
	query := `UPDATE words SET source_word = $1, target_word = $2, comment = $3, updated_at = now() 
              WHERE id = $4 AND user_id = $5`
	_, err := p.db.db.ExecContext(ctx, query, req.SourceWord, req.TargetWord, req.Comment, req.ID, userID)
	return err
}

func (p *wordsPostgres) GetWordsForGemini(ctx context.Context, req *modelsDB.WordsForGeminiReq) ([]modelsDB.WordsForGeminiResp, error) {
	query := `
	SELECT source_word, target_word
	FROM words
	WHERE user_id = $1
	  AND source_lang = $2
	  AND target_lang = $3
	  AND state = 2 
	ORDER BY 
	  stability DESC, 
	  lapses ASC     
	LIMIT 500         
	`
	var resp []modelsDB.WordsForGeminiResp
	err := p.db.db.SelectContext(ctx, &resp, query, req.UserID, req.SourceLang, req.TargetLang)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *wordsPostgres) CreateBatch(ctx context.Context, reqs []modelsDB.CreateReq) error {
	// We write the query just like for a single row
	query := `
	insert into words (user_id, source_lang, target_lang, source_word, target_word, comment) 
	values (:user_id, :source_lang, :target_lang, :source_word, :target_word, :comment)
	on conflict (user_id, source_lang, target_lang, source_word, target_word) do nothing
	`

	// But we pass a SLICE of reqs structs. sqlx will expand this into a batch insert!
	_, err := p.db.db.NamedExecContext(ctx, query, reqs)
	if err != nil {
		return fmt.Errorf("failed to bulk insert words: %w", err)
	}
	return nil
}

func (p *wordsPostgres) GetGeminiWordList(ctx context.Context, sourceLang, targetLang, level, topic string) (*modelsDB.GeminiWordList, error) {
	var wordList modelsDB.GeminiWordList
	query := `SELECT id, source_lang, target_lang, level, topic, response, created_at 
              FROM gemini_word_lists 
              WHERE source_lang = $1 AND target_lang = $2 AND level = $3 AND topic = $4`
	err := p.db.db.GetContext(ctx, &wordList, query, sourceLang, targetLang, level, topic)
	if err != nil {
		return nil, err
	}
	return &wordList, nil
}

func (p *wordsPostgres) SaveGeminiWordList(ctx context.Context, wordList modelsDB.GeminiWordList) error {
	query := `INSERT INTO gemini_word_lists (source_lang, target_lang, level, topic, response) 
              VALUES ($1, $2, $3, $4, $5)
              ON CONFLICT (source_lang, target_lang, level, topic) DO UPDATE 
              SET response = EXCLUDED.response, created_at = CURRENT_TIMESTAMP`
	_, err := p.db.db.ExecContext(ctx, query, wordList.SourceLang, wordList.TargetLang, wordList.Level, wordList.Topic, wordList.Response)
	return err
}

func (p *wordsPostgres) GetProgressStats(ctx context.Context, userID uuid.UUID) (*models.ProgressStats, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE state = 0) as new,
			COUNT(*) FILTER (WHERE state = 1) as learning,
			COUNT(*) FILTER (WHERE state = 2) as review,
			COUNT(*) FILTER (WHERE state = 3) as relearning
		FROM words 
		WHERE user_id = $1
	`

	var stats models.ProgressStats
	err := p.db.db.QueryRowxContext(ctx, query, userID).StructScan(&stats)
	if err != nil {
		return nil, fmt.Errorf("error getting progress stats: %w", err)
	}

	return &stats, nil
}

func (p *wordsPostgres) CreatePublicWordList(ctx context.Context, list modelsDB.PublicWordList, items []modelsDB.PublicWordListItem) (uuid.UUID, error) {
	tx, err := p.db.db.BeginTxx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()

	var listID uuid.UUID
	queryList := `
		INSERT INTO public_word_lists (user_id, title, description, source_lang, target_lang, level)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	err = tx.GetContext(ctx, &listID, queryList, list.UserID, list.Title, list.Description, list.SourceLang, list.TargetLang, list.Level)
	if err != nil {
		return uuid.Nil, err
	}

	queryItems := `
		INSERT INTO public_word_list_items (list_id, source_word, target_word, comment)
		VALUES ($1, $2, $3, $4)`
	for _, item := range items {
		_, err = tx.ExecContext(ctx, queryItems, listID, item.SourceWord, item.TargetWord, item.Comment)
		if err != nil {
			return uuid.Nil, err
		}
	}

	return listID, tx.Commit()
}

func (p *wordsPostgres) GetPublicWordLists(ctx context.Context, sourceLang, targetLang, level string) ([]modelsDB.PublicWordList, error) {
	var lists []modelsDB.PublicWordList
	query := `
		SELECT l.id, l.user_id, COALESCE(u.first_name || ' ' || u.last_name, '') as creator_name, l.title, l.description, l.source_lang, l.target_lang, l.level, l.created_at, l.updated_at 
		FROM public_word_lists l
		JOIN users u ON l.user_id = u.id`
	var args []any
	var conditions []string

	if sourceLang != "" {
		conditions = append(conditions, fmt.Sprintf("l.source_lang = $%d", len(args)+1))
		args = append(args, sourceLang)
	}
	if targetLang != "" {
		conditions = append(conditions, fmt.Sprintf("l.target_lang = $%d", len(args)+1))
		args = append(args, targetLang)
	}
	if level != "" {
		conditions = append(conditions, fmt.Sprintf("l.level = $%d", len(args)+1))
		args = append(args, level)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += ` ORDER BY l.created_at DESC`
	err := p.db.db.SelectContext(ctx, &lists, query, args...)
	return lists, err
}

func (p *wordsPostgres) GetPublicWordListByID(ctx context.Context, listID uuid.UUID) (*modelsDB.PublicWordListDetail, error) {
	var detail modelsDB.PublicWordListDetail
	queryList := `
		SELECT l.id, l.user_id, COALESCE(u.first_name || ' ' || u.last_name, '') as creator_name, l.title, l.description, l.source_lang, l.target_lang, l.level, l.created_at, l.updated_at 
		FROM public_word_lists l
		JOIN users u ON l.user_id = u.id
		WHERE l.id = $1`
	err := p.db.db.GetContext(ctx, &detail.PublicWordList, queryList, listID)
	if err != nil {
		return nil, err
	}

	queryItems := `SELECT id, list_id, source_word, target_word, comment, created_at FROM public_word_list_items WHERE list_id = $1`
	err = p.db.db.SelectContext(ctx, &detail.Items, queryItems, listID)
	if err != nil {
		return nil, err
	}

	return &detail, nil
}

func (p *wordsPostgres) UpdatePublicWordList(ctx context.Context, list modelsDB.PublicWordList, items []modelsDB.PublicWordListItem) error {
	tx, err := p.db.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update list metadata
	queryList := `
		UPDATE public_word_lists 
		SET title = $1, description = $2, level = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4`
	_, err = tx.ExecContext(ctx, queryList, list.Title, list.Description, list.Level, list.ID)
	if err != nil {
		return err
	}

	// Simple approach: delete all items and re-insert
	// This is easier for MVP than tracking diffs
	queryDeleteItems := `DELETE FROM public_word_list_items WHERE list_id = $1`
	_, err = tx.ExecContext(ctx, queryDeleteItems, list.ID)
	if err != nil {
		return err
	}

	queryInsertItems := `
		INSERT INTO public_word_list_items (list_id, source_word, target_word, comment)
		VALUES ($1, $2, $3, $4)`
	for _, item := range items {
		_, err = tx.ExecContext(ctx, queryInsertItems, list.ID, item.SourceWord, item.TargetWord, item.Comment)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (p *wordsPostgres) CreatePlaylist(ctx context.Context, playlist modelsDB.Playlist, listIDs []uuid.UUID) (uuid.UUID, error) {
	tx, err := p.db.db.BeginTxx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()

	var playlistID uuid.UUID
	queryPlaylist := `
		INSERT INTO playlists (user_id, title, description)
		VALUES ($1, $2, $3)
		RETURNING id`
	err = tx.GetContext(ctx, &playlistID, queryPlaylist, playlist.UserID, playlist.Title, playlist.Description)
	if err != nil {
		return uuid.Nil, err
	}

	queryPlaylistLists := `
		INSERT INTO playlist_public_lists (playlist_id, list_id, sort_order)
		VALUES ($1, $2, $3)`
	for i, listID := range listIDs {
		_, err = tx.ExecContext(ctx, queryPlaylistLists, playlistID, listID, i)
		if err != nil {
			return uuid.Nil, err
		}
	}

	return playlistID, tx.Commit()
}

func (p *wordsPostgres) GetPlaylists(ctx context.Context) ([]modelsDB.Playlist, error) {
	var playlists []modelsDB.Playlist
	query := `
		SELECT p.id, p.user_id, COALESCE(u.first_name || ' ' || u.last_name, '') as creator_name, p.title, p.description, p.created_at, p.updated_at 
		FROM playlists p
		JOIN users u ON p.user_id = u.id
		ORDER BY p.created_at DESC`
	err := p.db.db.SelectContext(ctx, &playlists, query)
	return playlists, err
}

func (p *wordsPostgres) GetPlaylistByID(ctx context.Context, playlistID uuid.UUID) (*modelsDB.PlaylistDetail, error) {
	var detail modelsDB.PlaylistDetail
	queryPlaylist := `
		SELECT p.id, p.user_id, COALESCE(u.first_name || ' ' || u.last_name, '') as creator_name, p.title, p.description, p.created_at, p.updated_at 
		FROM playlists p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = $1`
	err := p.db.db.GetContext(ctx, &detail.Playlist, queryPlaylist, playlistID)
	if err != nil {
		return nil, err
	}

	queryLists := `
		SELECT l.id, l.user_id, COALESCE(u.first_name || ' ' || u.last_name, '') as creator_name, l.title, l.description, l.source_lang, l.target_lang, l.level, l.created_at, l.updated_at 
		FROM public_word_lists l
		JOIN users u ON l.user_id = u.id
		JOIN playlist_public_lists pl ON l.id = pl.list_id
		WHERE pl.playlist_id = $1
		ORDER BY pl.sort_order`
	err = p.db.db.SelectContext(ctx, &detail.Lists, queryLists, playlistID)
	if err != nil {
		return nil, err
	}

	return &detail, nil
}

func (p *wordsPostgres) UpdatePlaylist(ctx context.Context, playlist modelsDB.Playlist, listIDs []uuid.UUID) error {
	tx, err := p.db.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryUpdate := `
		UPDATE playlists 
		SET title = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND user_id = $4`
	_, err = tx.ExecContext(ctx, queryUpdate, playlist.Title, playlist.Description, playlist.ID, playlist.UserID)
	if err != nil {
		return err
	}

	queryDelete := `DELETE FROM playlist_public_lists WHERE playlist_id = $1`
	_, err = tx.ExecContext(ctx, queryDelete, playlist.ID)
	if err != nil {
		return err
	}

	queryInsert := `
		INSERT INTO playlist_public_lists (playlist_id, list_id, sort_order)
		VALUES ($1, $2, $3)`
	for i, listID := range listIDs {
		_, err = tx.ExecContext(ctx, queryInsert, playlist.ID, listID, i)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (p *wordsPostgres) DeletePlaylist(ctx context.Context, playlistID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM playlists WHERE id = $1 AND user_id = $2`
	_, err := p.db.db.ExecContext(ctx, query, playlistID, userID)
	return err
}
