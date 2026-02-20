package sqlconnect

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDB(dbname string) (*sql.DB, error) {
	fmt.Println("Connecting to database...")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	connectionString := fmt.Sprintf("%s:%s@tcp(%s%s)/%s", user, password, host, port, dbname)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to database successfully")
	return db, nil
}
