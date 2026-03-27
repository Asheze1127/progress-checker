package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"

	_ "github.com/lib/pq"

	"github.com/Asheze1127/progress-checker/backend/util"
)

type seedStaff struct {
	Name     string
	Email    string
	Password string
}

var staffMembers = []seedStaff{
	{
		Name:     "Tsubasa Ito",
		Email:    "ito.tsubasa577@gmail.com",
		Password: "ito.tsubasa577@gmail.com",
	},
}

func main() {
	cfg, err := loadDBConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	hasher := util.NewPasswordHasher()

	for _, s := range staffMembers {
		hash, err := hasher.Hash(s.Password)
		if err != nil {
			log.Fatalf("failed to hash password for %s: %v", s.Email, err)
		}

		result, err := db.Exec(
			`INSERT INTO staff (name, email, password_hash)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (email) DO UPDATE SET
			   password_hash = EXCLUDED.password_hash,
			   updated_at = now()`,
			s.Name, s.Email, hash,
		)
		if err != nil {
			log.Fatalf("failed to insert staff %s: %v", s.Email, err)
		}

		rows, _ := result.RowsAffected()
		fmt.Printf("Seeded staff: %s (%s) [rows affected: %d]\n", s.Name, s.Email, rows)
	}

	fmt.Println("Seed completed successfully.")
}

func loadDBConfig() (string, error) {
	host := envOrDefault("DATABASE_HOST", "localhost")
	port := envOrDefault("DATABASE_PORT", "5432")
	name := envOrDefault("DATABASE_NAME", "progress_checker")
	user := envOrDefault("DATABASE_USER", "postgres")
	pass := envOrDefault("DATABASE_PASS", "postgres")
	sslMode := envOrDefault("DATABASE_SSL_MODE", "disable")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.PathEscape(user),
		url.PathEscape(pass),
		url.PathEscape(host),
		url.PathEscape(port),
		url.PathEscape(name),
		url.QueryEscape(sslMode),
	)
	return dsn, nil
}

func envOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
