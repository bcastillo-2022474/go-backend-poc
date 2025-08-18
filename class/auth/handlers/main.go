package handlers

import (
	authv1 "class-backend/proto/generated/go/auth/v1"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	pool *pgxpool.Pool
}

func NewAuthHandler(pool *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{
		pool: pool,
	}
}
