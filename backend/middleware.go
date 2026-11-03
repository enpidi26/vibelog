package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ctxKey string

const userIDKey ctxKey = "userID"

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

// decodeJSON membaca body dengan batas 1 MB dan menolak JSON yang bukan
// objek atau punya field yang tidak dikenal (membantu menangkap typo).
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		switch {
		case errors.As(err, &maxErr):
			writeError(w, http.StatusRequestEntityTooLarge, "body_too_large", "body request terlalu besar")
		case errors.Is(err, io.EOF):
			writeError(w, http.StatusBadRequest, "empty_body", "body request kosong")
		default:
			writeError(w, http.StatusBadRequest, "invalid_json", "body JSON tidak valid")
		}
		return false
	}
	if dec.More() {
		writeError(w, http.StatusBadRequest, "invalid_json", "body JSON harus satu objek")
		return false
	}
	return true
}

func parseID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func userIDFromContext(ctx context.Context) int64 {
	id, _ := ctx.Value(userIDKey).(int64)
	return id
}

func authRequired(cfg Config, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "unauthorized", "token tidak ada")
			return
		}
		uid, _, _, err := ParseToken(cfg.EffectiveJWTSecret(), strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "token tidak valid atau kedaluwarsa")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey, uid)))
	})
}

func cors(cfg Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Hanya origin frontend yang dikonfigurasi yang diizinkan. Request
		// tanpa Origin (curl, server-to-server) tetap dilayani.
		if origin != "" && origin == cfg.FrontendURL {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				writeError(w, http.StatusInternalServerError, "internal_error", "terjadi kesalahan di server")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func requestTimeout(d time.Duration, next http.Handler) http.Handler {
	return http.TimeoutHandler(next, d, `{"error":{"code":"timeout","message":"request timeout"}}`)
}
