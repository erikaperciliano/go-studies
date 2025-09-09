package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Response struct {
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

func sendJSON(w http.ResponseWriter, resp Response, status int) {
	data, err := json.Marshal(resp)
	if err != nil {
		slog.Error("error ao fazer mashal de json", "error", err)
		fmt.Println("error ao fazer marshal de json", err)
		sendJSON(w, Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		slog.Error("error ao enviar resposta:", "error", err)
		return
	}
}

type User struct {
	Username string
	ID       int64 `json:"id,string"`
	Role     string
	Password string `json:"-"`
}

type Password string

func (p Password) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}

func (u User) LogValue() slog.Value {
	return slog.GroupValue(slog.Int64("id", u.ID), slog.String("role", u.Role))
}

const LevelFoo = slog.Level(-50)

func main() {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     LevelFoo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "level" {
				level := a.Value.String()
				if level == "DEBUG-46" {
					a.Value = slog.StringValue("FOO")
				}
			}

			return a
		},
	}
	l := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	p := Password("123456")
	slog.Info("password", "p", p)
	slog.SetDefault(l)
	slog.Info("Serviço sendo iniciado", "time", time.Now(), "version", "1.0.0")
	l.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		"tivemos um http request",
		slog.Group("http_data",
			slog.String("method", http.MethodDelete),
			slog.Int("status", http.StatusOK),
		),
		slog.Duration("time_taken", time.Second),
		slog.String("user_agent", "hasuida"),
	)

	r := chi.NewMux()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	db := map[int64]User{
		1: {
			Username: "admin",
			Password: "admin",
			Role:     "admin",
			ID:       1,
		},
	}

	r.Group(func(r chi.Router) {
		r.Use((jsonMiddleware))
		r.Get("/users/{id:[0-9]+}", handleGetUsers(db))
		r.Post("/users", handlePostUsers(db))
	})

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	})
}

func handleGetUsers(db map[int64]User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseInt(idStr, 10, 64)

		user, ok := db[id]

		if !ok {
			sendJSON(w, Response{Error: "usuário não encontrado"}, http.StatusNotFound)
			return
		}
		sendJSON(w, Response{Data: user}, http.StatusOK)
	}
}

func handlePostUsers(db map[int64]User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1000)
		data, err := io.ReadAll(r.Body)

		if err != nil {
			var maxErr *http.MaxBytesError

			if errors.As(err, &maxErr) {
				sendJSON(w, Response{Error: "body too large"}, http.StatusRequestEntityTooLarge)
				return
			}

			slog.Error("falha ao ler o json do usuário", "error", err)
			sendJSON(w, Response{Error: "something went wrong"}, http.StatusInternalServerError)
			return
		}

		var user User
		if err := json.Unmarshal(data, &user); err != nil {
			sendJSON(w, Response{Error: "invalid body"}, http.StatusUnprocessableEntity)
			return
		}

		db[user.ID] = user

		w.WriteHeader(http.StatusCreated)
	}
}
