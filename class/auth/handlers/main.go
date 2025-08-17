package handlers

import (
	authv1 "class-backend/proto/generated/go/auth/v1"

	"github.com/jackc/pgx/v5"
)

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	db *pgx.Conn
}

func NewAuthHandler(db *pgx.Conn) *AuthHandler {
	return &AuthHandler{
		db: db,
	}
}
