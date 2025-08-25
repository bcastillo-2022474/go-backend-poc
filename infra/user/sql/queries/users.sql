-- name: CreateUser :one
INSERT INTO users (id, name, email, password_hash)
VALUES (@id, @name, @email, @password_hash)
RETURNING id, name, email, created_at, updated_at;

-- name: ExistsByEmail :one  
SELECT EXISTS(SELECT 1 FROM users WHERE email = @email);

-- name: FindByEmail :one
SELECT id, name, email, created_at, updated_at 
FROM users 
WHERE email = @email;