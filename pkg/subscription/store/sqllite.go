package store

import (
	"database/sql"
	"errors"
	"fmt"
)

type SubscriptionStore struct {
	db *sql.DB
}

func NewSubscriptionStore(db *sql.DB) *SubscriptionStore {
	return &SubscriptionStore{db: db}
}

func (s *SubscriptionStore) Subscribe(id int64) error {
	_, err := s.db.Exec(
		`INSERT INTO subscribers(telegram_id) VALUES (?) ON CONFLICT (telegram_id) DO NOTHING`,
		id,
	)
	if err != nil {
		return fmt.Errorf("store unable to subscribe err: %v", err)
	}
	return nil
}

func (s *SubscriptionStore) Unsubscribe(id int64) error {
	_, err := s.db.Exec(`DELETE FROM subscribers WHERE telegram_id = ?`, id)
	if err != nil {
		return fmt.Errorf("store unable to unsubscribe err: %v", err)
	}
	return nil
}

func (s *SubscriptionStore) GetAllSubscribedChats() ([]int64, error) {
	var res []int64
	rows, err := s.db.Query(`SELECT telegram_id FROM subscribers`)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("store unable to get all subscribed ids err: %v", err)
	}
	defer rows.Close()
	if errors.Is(err, sql.ErrNoRows) {
		return res, nil
	}

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("store scanning id: err %v", err)
		}
		res = append(res, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: getting rows err: %v", err)
	}
	return res, nil
}
