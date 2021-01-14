package sqlHandler

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type TrxIdHandler struct {
	db *sql.DB
}

func (s *TrxIdHandler) ConnectDB() (err error) {
	s.db, err = sql.Open("mysql", "root:root@tcp("+os.Getenv("DB_HOST")+":3306)/bankbe")
	if err != nil {
		return err
	}
	return nil
}

func (s *TrxIdHandler) CloseConnection() (err error) {
	err = s.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *TrxIdHandler) IncreaseAndGetTrxid() (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec("insert into trxid (id) values (NULL);")
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	rows, err := tx.Query("SELECT MAX(id) FROM trxid")

	if err != nil {
		return 0, err
	}

	var id int
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return 0, err
		}
		fmt.Println(id)
	}

	return id, nil
}
