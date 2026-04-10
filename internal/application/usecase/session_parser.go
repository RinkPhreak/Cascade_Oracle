package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

type TelethonSession struct {
	DCID          int
	ServerAddress string
	Port          int
	AuthKey       []byte
	TakeoutID     int
}

func parseTelethonSession(ctx context.Context, filePath string) (*TelethonSession, error) {
	// telethon .session files are sqlite databases
	db, err := sql.Open("sqlite", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open session file: %w", err)
	}
	defer db.Close()

	var sess TelethonSession
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

// ensureTempFile stores raw bytes into a temporary file for sqlite access
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
