package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"cascade/internal/domain"
	_ "modernc.org/sqlite"
)

type sessionStorage struct{}

func NewSessionStorage() *sessionStorage {
	return &sessionStorage{}
}

func (s *sessionStorage) ParseTelethonSession(ctx context.Context, data []byte) (*domain.TelethonSession, error) {
	filePath, err := createTempSessionFile(data)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp session file: %w", err)
	}
	defer os.Remove(filePath)

	db, err := sql.Open("sqlite", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open session file: %w", err)
	}
	defer db.Close()

	var sess domain.TelethonSession
	row := db.QueryRowContext(ctx, "SELECT dc_id, server_address, port, auth_key, takeout_id FROM sessions")
	err = row.Scan(&sess.DCID, &sess.ServerAddress, &sess.Port, &sess.AuthKey, &sess.TakeoutID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no session data found in sqlite")
		}
		return nil, fmt.Errorf("failed to scan session data: %w", err)
	}

	return &sess, nil
}

func createTempSessionFile(data []byte) (string, error) {
	tmpFile, err := os.CreateTemp("", "cascade_session_*.session")
	if err != nil {
		return "", err
	}
	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return "", err
	}
	path := tmpFile.Name()
	_ = tmpFile.Close()
	return path, nil
}
