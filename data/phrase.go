package data

import (
	"context"
	"phrase_bot/types"
)

func GetAllPhrases(db DB) (*[]types.Phrase, error) {
	rows, err := db.Query(context.Background(), "select id, phrase from phrase order by inserted_at desc")
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

func SearchPhrases(db DB, search string) (*[]types.Phrase, error) {
	searchStmt := `
  select id, phrase from phrase
  where to_tsvector('english', phrase) @@ plainto_tsquery('english', $1)
  order by ts_rank_cd(to_tsvector('english', phrase), plainto_tsquery('english', $1))
  `
	rows, err := db.Query(context.Background(), searchStmt, search)
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

func GetRandomPhrase(db DB) (types.Phrase, error) {
	// This will be slow if the table gets really big, which it probably won't
	row := db.QueryRow(context.Background(), "select id, phrase from phrase order by random() limit 1")
	var id int
	var phrase string
	err := row.Scan(&id, &phrase)
	if err != nil {
		return types.Phrase{}, err
	}
	return types.Phrase{Id: id, Phrase: phrase}, nil
}

func CreatePhrase(db DB, phrase string) (types.Phrase, error) {
	row := db.QueryRow(context.Background(), "insert into phrase (phrase) values ($1) returning id", phrase)
	var id int
	err := row.Scan(&id)
	if err != nil {
		return types.Phrase{}, err
	}
	return types.Phrase{Id: id, Phrase: phrase}, nil
}

func DeletePhrase(db DB, id int) error {
	_, err := db.Exec(context.Background(), "delete from phrase where id = $1", id)
	return err
}
