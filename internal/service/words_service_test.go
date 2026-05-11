package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"synaply/external/gemini"
	"synaply/internal/models"
)

type fakeGeminiService struct {
	req  *gemini.GemTranslationReq
	resp *gemini.GemTranslationResp
	err  error
}

func (f *fakeGeminiService) WordInfo(ctx context.Context, req gemini.WordInfoRequest) (string, error) {
	return "", nil
}

func (f *fakeGeminiService) StartPracticeWithGemini(ctx context.Context, req *gemini.PracticeWithGemini, wordList string) (*gemini.StartPracticeWithGeminiResponse, error) {
	return nil, nil
}

func (f *fakeGeminiService) CheckAnswerPracticeWithGemini(ctx context.Context, req *gemini.PracticeWithGemini, translate string) (*gemini.CheckAnswerPracticeWithGeminiResponse, error) {
	return nil, nil
}

func (f *fakeGeminiService) WordList(ctx context.Context, req gemini.WordListReq) ([]gemini.WordListResp, error) {
	return nil, nil
}

func (f *fakeGeminiService) WordTranslate(ctx context.Context, req *gemini.GemTranslationReq) (*gemini.GemTranslationResp, error) {
	f.req = req
	return f.resp, f.err
}

func TestWordsServiceTranslate(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name    string
		req     models.TranslateReq
		gemResp *gemini.GemTranslationResp
		gemErr  error
		wantReq *gemini.GemTranslationReq
		want    *models.TranslateResp
		wantErr bool
	}{
		{
			name: "source to target",
			req: models.TranslateReq{
				ID:         id,
				SourceLang: "en",
				TargetLang: "es",
				SourceWord: " cat ",
			},
			gemResp: &gemini.GemTranslationResp{SourceWord: "cat", TargetWord: "gato"},
			wantReq: &gemini.GemTranslationReq{
				SourceLang: "en",
				TargetLang: "es",
				SourceWord: "cat",
			},
			want: &models.TranslateResp{
				ID:         id,
				SourceWord: "cat",
				TargetWord: "gato",
			},
		},
		{
			name: "target to source fills original target when gemini omits it",
			req: models.TranslateReq{
				ID:         id,
				SourceLang: "en",
				TargetLang: "es",
				TargetWord: " gato ",
			},
			gemResp: &gemini.GemTranslationResp{SourceWord: "cat"},
			wantReq: &gemini.GemTranslationReq{
				SourceLang: "en",
				TargetLang: "es",
				TargetWord: "gato",
			},
			want: &models.TranslateResp{
				ID:         id,
				SourceWord: "cat",
				TargetWord: "gato",
			},
		},
		{
			name: "rejects blank words",
			req: models.TranslateReq{
				SourceLang: "en",
				TargetLang: "es",
				SourceWord: " ",
			},
			wantErr: true,
		},
		{
			name: "rejects nil gemini response",
			req: models.TranslateReq{
				SourceLang: "en",
				TargetLang: "es",
				SourceWord: "cat",
			},
			wantReq: &gemini.GemTranslationReq{
				SourceLang: "en",
				TargetLang: "es",
				SourceWord: "cat",
			},
			wantErr: true,
		},
		{
			name: "propagates gemini error",
			req: models.TranslateReq{
				SourceLang: "en",
				TargetLang: "es",
				SourceWord: "cat",
			},
			gemErr: errors.New("gemini failed"),
			wantReq: &gemini.GemTranslationReq{
				SourceLang: "en",
				TargetLang: "es",
				SourceWord: "cat",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gem := &fakeGeminiService{resp: tt.gemResp, err: tt.gemErr}
			svc := NewWordsService(nil, nil, nil, nil, gem)

			got, err := svc.Translate(context.Background(), tt.req)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gem.req == nil {
				t.Fatal("expected gemini request")
			}
			if *gem.req != *tt.wantReq {
				t.Fatalf("gemini request mismatch: got %+v, want %+v", *gem.req, *tt.wantReq)
			}
			if *got != *tt.want {
				t.Fatalf("response mismatch: got %+v, want %+v", *got, *tt.want)
			}
		})
	}
}
