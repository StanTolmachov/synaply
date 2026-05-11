package handler

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"synaply/internal/middleware"
)

func RegisterRoutes(h *Handler, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"https://synaply.me",
			"http://localhost:3000", // for local frontend development
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.LoggerMiddleware)

	r.Get("/health", h.Health)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/v1/", func(r chi.Router) {
		r.Use(httprate.LimitByIP(100000, 1*time.Minute))

		r.Group(func(r chi.Router) {
			r.Use(middleware.TimeoutMiddleware(time.Second * 5))

			r.Route("/users", func(r chi.Router) {
				r.Get("/lang", h.Lang)
				r.With(httprate.LimitByIP(1000, 1*time.Hour)).Post("/create", h.Create)
				r.With(httprate.LimitByIP(1000, 1*time.Hour)).Post("/login", h.Login)

				r.Group(func(r chi.Router) {
					r.Use(middleware.AuthMidleware(jwtSecret))

					//r.Get("/{id}", h.GetUserByID)
					//r.Get("/all", h.GetUsers)

					r.Put("/{id}", h.Update)
					r.Delete("/{id}", h.Delete)

				})
			})
		})
		r.Route("/public-lists", func(r chi.Router) {
			// Publicly accessible GET routes
			r.Get("/", h.GetPublicWordLists)
			r.Get("/{id}", h.GetPublicWordListByID)

			// Authenticated routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMidleware(jwtSecret))
				r.Post("/", h.CreatePublicWordList)
				r.Put("/{id}", h.UpdatePublicWordList)
				r.Post("/{id}/add", h.AddPublicListToUser)
			})
		})

		r.Route("/playlists", func(r chi.Router) {
			r.Get("/", h.GetPlaylists)
			r.Get("/{id}", h.GetPlaylistByID)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMidleware(jwtSecret))
				r.Post("/", h.CreatePlaylist)
				r.Put("/{id}", h.UpdatePlaylist)
				r.Delete("/{id}", h.DeletePlaylist)
			})
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.AuthMidleware(jwtSecret))
			r.Use(middleware.AdminOnly)

			r.Get("/stats", h.GetAdminStats)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMidleware(jwtSecret))

			r.Group(func(r chi.Router) {
				r.Use(middleware.TimeoutMiddleware(time.Second * 5))

				r.Route("/words", func(r chi.Router) {
					r.With(httprate.LimitByIP(3000, 1*time.Minute)).Post("/create", h.NewWord)
					r.With(httprate.LimitByIP(2000, 1*time.Minute)).Post("/translate", h.Translate)
					r.Get("/GetMe", h.GetMe)
					r.Get("/", h.GetWordsList)
					r.Get("/stats", h.GetProgressStats)
					r.Post("/import", h.ImportWords)
					r.Delete("/all", h.DeleteAllWords)
					r.Put("/{id}", h.UpdateWordFields)
					r.Delete("/{id}", h.DeleteWord)
				})
				r.Route("/lesson", func(r chi.Router) {
					r.Get("/start", h.StartLesson)
					r.Post("/check", h.CheckAnswer)
					r.Post("/finish", h.Finish)
				})
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.TimeoutMiddleware(time.Second * 30))

				r.With(httprate.LimitByIP(1000, 1*time.Minute)).Post("/words/wordInfo", h.WordInfo)

				r.With(httprate.LimitByIP(1000, 1*time.Minute)).Post("/practice/startPractice", h.StartPracticeWithGemini)
				r.With(httprate.LimitByIP(1000, 1*time.Minute)).Post("/practice/checkAnswerPractice", h.CheckAnswerPracticeWithGemini)
				r.Post("/practice/finishPractice", h.FinishPracticeWithGemini)
				r.With(httprate.LimitByIP(1000, 1*time.Minute)).Post("/words/wordList", h.WordList)
				r.With(httprate.LimitByIP(1000, 1*time.Minute)).Post("/words/create-batch", h.CreateBatchWords)
			})
		})
	})

	return r
}
