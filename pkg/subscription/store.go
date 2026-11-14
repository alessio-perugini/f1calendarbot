package subscription

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Subscribe(ctx context.Context, id int64, chatType string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO subscribers(telegram_id, telegram_chat_type) VALUES (?, ?)
		ON CONFLICT (telegram_id) DO UPDATE SET telegram_chat_type = excluded.telegram_chat_type`,
		id, chatType,
	)
	if err != nil {
		return fmt.Errorf("subscribing chat %d: %w", id, err)
	}
	return nil
}

func (s *Store) Unsubscribe(ctx context.Context, id int64) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM subscribers WHERE telegram_id = ?`, id); err != nil {
		return fmt.Errorf("unsubscribing chat %d: %w", id, err)
	}
	return nil
}

// Migrate moves a subscription to the new chat id telegram assigns when a group
// is upgraded to a supergroup. An existing row for the new id is replaced.
func (s *Store) Migrate(ctx context.Context, oldID, newID int64) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE OR REPLACE subscribers SET telegram_id = ?, telegram_chat_type = 'supergroup' WHERE telegram_id = ?`,
		newID, oldID,
	)
	if err != nil {
		return fmt.Errorf("migrating chat %d to %d: %w", oldID, newID, err)
	}
	return nil
}

func (s *Store) GetAllSubscribedChats(ctx context.Context) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT telegram_id FROM subscribers`)
	if err != nil {
		return nil, fmt.Errorf("listing subscribed chats: %w", err)
	}
	defer rows.Close()

	var res []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning subscribed chat: %w", err)
		}
		res = append(res, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating subscribed chats: %w", err)
	}
	return res, nil
}
