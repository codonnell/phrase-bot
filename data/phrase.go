package data

import (
	"context"
	"phrase_bot/types"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetAllPhrases(pool *pgxpool.Pool) (*[]types.Phrase, error) {
	rows, err := pool.Query(context.Background(), "select id, phrase from phrase order by inserted_at desc")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	phrases := make([]types.Phrase, 0)
	for rows.Next() {
		var id int
		var phrase string
		err = rows.Scan(&id, &phrase)
		if err != nil {
			return nil, err
		}
		phrases = append(phrases, types.Phrase{Id: id, Phrase: phrase})
	}
	return &phrases, nil
}

func SearchPhrases(pool *pgxpool.Pool, search string) (*[]types.Phrase, error) {
	searchStmt := `
  select id, phrase from phrase
  where to_tsvector('english', phrase) @@ plainto_tsquery('english', $1)
  order by ts_rank_cd(to_tsvector('english', phrase), plainto_tsquery('english', $1))
  `
	rows, err := pool.Query(context.Background(), searchStmt, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	phrases := make([]types.Phrase, 0)
	for rows.Next() {
		var id int
		var phrase string
		err = rows.Scan(&id, &phrase)
		if err != nil {
			return nil, err
		}
		phrases = append(phrases, types.Phrase{Id: id, Phrase: phrase})
	}
	return &phrases, nil
}

func GetRandomPhrase(pool *pgxpool.Pool) (types.Phrase, error) {
	// This will be slow if the table gets really big, which it probably won't
	row := pool.QueryRow(context.Background(), "select id, phrase from phrase order by random() limit 1")
	var id int
	var phrase string
	err := row.Scan(&id, &phrase)
	if err != nil {
		return types.Phrase{}, err
	}
	return types.Phrase{Id: id, Phrase: phrase}, nil
}

func CreatePhrase(pool *pgxpool.Pool, phrase string) (types.Phrase, error) {
	row := pool.QueryRow(context.Background(), "insert into phrase (phrase) values ($1) returning id", phrase)
	var id int
	err := row.Scan(&id)
	if err != nil {
		return types.Phrase{}, err
	}
	return types.Phrase{Id: id, Phrase: phrase}, nil
}

func DeletePhrase(pool *pgxpool.Pool, id int) error {
	_, err := pool.Exec(context.Background(), "delete from phrase where id = $1", id)
	return err
}
