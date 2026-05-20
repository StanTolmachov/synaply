Лучший вариант без лишних миграций: хранить в gemini_word_lists.response накопленный список по ключу:
source_lang + target_lang + level + topic
А клиенту отдавать порциями по 30.
Схема:
1.
Первый запрос: offset=0, limit=30
В БД списка нет -> Gemini генерирует 30 слов -> сохраняешь эти 30 -> отдаёшь клиенту.
2.
Второй запрос: offset=30, limit=30
В БД уже есть 30 слов, но клиент просит слова 30:60 -> сервер видит, что данных не хватает -> отправляет Gemini запрос: “сгенерируй ещё 30, НЕ используя эти слова” -> добавляет новые слова в старый JSON -> сохраняет 60 -> отдаёт вторые 30.
3.
Третий запрос: offset=60, limit=30
Аналогично: передаёшь Gemini уже существующие 60 как exclusion list.
То есть повторов избегает сервер, а не фронт.
Минимальное изменение API:
type WordListReq struct {
SourceLang string `json:"source_lang"`
TargetLang string `json:"target_lang"`
Level      string `json:"level"`
Topic      string `json:"topic"`
UserTopic  string `json:"user_topic"`
Limit      int    `json:"limit"`
Offset     int    `json:"offset"`
}
По умолчанию:
if req.Limit <= 0 {
req.Limit = 30
}
if req.Limit > 30 {
req.Limit = 30
}
В Gemini prompt добавить параметры:
Generate exactly %[6]d NEW words.

Already generated words, DO NOT repeat or paraphrase them:
%[7]s
В сервисе логика такая:
cachedWords := loadFromDB(...)

needUntil := req.Offset + req.Limit

if len(cachedWords) < needUntil {
missing := needUntil - len(cachedWords)

	exclude := cachedWords
	newWords := s.gem.WordList(ctx, req, missing, exclude)

	newWords = removeDuplicates(cachedWords, newWords)
	cachedWords = append(cachedWords, newWords...)

	saveFullListToDB(cachedWords)
}

return cachedWords[req.Offset:min(req.Offset+req.Limit, len(cachedWords))]
Важно сделать dedupe после Gemini, потому что модель всё равно иногда повторит:
func removeDuplicates(existing, incoming []models.WordListResp) []models.WordListResp {
seen := make(map[string]struct{}, len(existing))

```go
	for _, w := range existing {
		key := strings.ToLower(strings.TrimSpace(w.SourceWord)) + "|" +
			strings.ToLower(strings.TrimSpace(w.TargetWord))
		seen[key] = struct{}{}
	}

	result := make([]models.WordListResp, 0, len(incoming))
	for _, w := range incoming {
		key := strings.ToLower(strings.TrimSpace(w.SourceWord)) + "|" +
			strings.ToLower(strings.TrimSpace(w.TargetWord))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, w)
	}

	return result
}
```

Я бы не делал page=1/2/3 в БД отдельными строками. Текущая таблица уже подходит: один JSONB со всеми накопленными словами по теме. Это проще, и можно отдавать любые порции: первые 30, следующие 30, показать всё, сохранить всё.