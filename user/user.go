package user

import (
	"database/sql"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/steveperjesi/integra-demo/internal/db"
)

func GetAllUsers(dbcon *sql.DB) ([]User, error) {
	query, args, err := sq.Select(db.AllColumns).
		From(DbName).
		ToSql()
	if err != nil {
		log.Print("failed to build select sql: ", err)
		return nil, err
	}

	var results []User

	rows, err := dbcon.Query(query, args...)
	if err != nil {
		log.Print("query failure: ", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var udb db.UserDB
		if err := rows.Scan(
			&udb.UserID,
			&udb.UserName,
			&udb.FirstName,
			&udb.LastName,
			&udb.Email,
			&udb.UserStatus,
			&udb.Department,
		); err != nil {
			log.Print("row scan failure: ", err)
			return nil, err
		}

		// Need to convert the UserDB into User
		results = append(results, ConvertToUser(&udb))
	}

	if err := rows.Err(); err != nil {
		log.Print("rows iteration error: ", err)
		return nil, err
	}

	return results, nil
}

func GetUser(dbcon *sql.DB, id int64) (*User, error) {
	if id == 0 {
		return nil, ErrMissingUserID
	}

	query, args, err := sq.Select(db.AllColumns).
		From(DbName).
		Where(sq.Eq{"user_id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Print("failed to build select sql: ", err)
		return nil, err
	}

	var result db.UserDB

	err = dbcon.QueryRow(query, args...).Scan(
		&result.UserID,
		&result.UserName,
		&result.FirstName,
		&result.LastName,
		&result.Email,
		&result.UserStatus,
		&result.Department,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	} else if err != nil {
		log.Print("row scan error: ", err)
		return nil, err
	}

	user := ConvertToUser(&result)
	return &user, nil
}

// Returns true if `user_name` exists
func CheckUserNameExists(dbcon *sql.DB, userName string) (bool, error) {
	if userName == "" {
		return false, ErrMissingUserName
	}

	query, args, err := sq.Select("COUNT(*)").
		From(DbName).
		Where(sq.Eq{"user_name": userName}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Print("failed to build select sql: ", err)
		return false, err
	}

	var count int

	err = dbcon.QueryRow(query, args...).Scan(&count)
	if err != nil {
		log.Print("row scan error: ", err)
		return false, err
	}

	return (count == 1), nil
}

func CreateUser(dbcon *sql.DB, u *User) (*User, error) {
	// Check if the `user_name` already exists
	userExists, err := CheckUserNameExists(dbcon, u.UserName)
	if err != nil {
		return nil, err
	}

	if userExists {
		return nil, ErrUserExists
	}

	// Need to convert the User into UserDB
	userDB := u.ConvertToUserDB()

	query, args, err := sq.Insert(DbName).
		Columns("user_name", "first_name", "last_name", "email", "user_status", "department").
		Values(userDB.UserName, userDB.FirstName, userDB.LastName, userDB.Email, userDB.UserStatus, userDB.Department).
		Suffix("RETURNING user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Print("failed to build create sql: ", err)
		return nil, err
	}

	// Execute the update and add the `user_id` to the result
	var lastInsertID int64
	err = dbcon.QueryRow(query, args...).Scan(&lastInsertID)
	if err != nil {
		log.Print("query failure: ", err)
		return nil, err
	}

	u.ID = lastInsertID

	return u, nil
}

func UpdateUser(dbcon *sql.DB, u *User) (*User, error) {
	if u.ID == 0 {
		return nil, ErrMissingUserID
	}

	updateValues := make(map[string]interface{})

	// Only update the values given
	if u.UserName != "" {
		updateValues["user_name"] = u.UserName
	}

	if u.FirstName != "" {
		updateValues["first_name"] = u.FirstName
	}

	if u.LastName != "" {
		updateValues["last_name"] = u.LastName
	}

	if u.Email != "" {
		updateValues["email"] = u.Email
	}

	if u.UserStatus != "" {
		updateValues["user_status"] = u.UserStatus
	}

	if u.Department != nil {
		updateValues["department"] = u.Department
	}

	if len(updateValues) == 0 {
		return nil, ErrUpdateUserMissingValues
	}

	query, args, err := sq.Update(DbName).
		SetMap(updateValues).
		Where(sq.Eq{"user_id": u.ID}).
		Suffix("RETURNING user_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Print("failed to build update sql: ", err)
		return nil, err
	}

	result, err := dbcon.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrUpdateUserNoRows
	}

	// Pull the updated user's data
	user, err := GetUser(dbcon, u.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func DeleteUser(dbcon *sql.DB, id int64) error {
	query, args, err := sq.Delete(DbName).
		Where(sq.Eq{"user_id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		log.Print("failed to build delete sql: ", err)
		return err
	}

	result, err := dbcon.Exec(query, args...)
	if err != nil {
		log.Print("query failure: ", err)
		return err
	}

	numRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if numRows == 0 {
		return ErrUserNotFound
	}

	return nil
}
