package user_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	db "github.com/steveperjesi/integra-demo/internal/db"
	. "github.com/steveperjesi/integra-demo/user"
)

func ptr(s string) *string {
	return &s
}

func convertToDriverArgs(args []interface{}) []driver.Value {
	driverArgs := make([]driver.Value, len(args))
	for i, v := range args {
		driverArgs[i] = driver.Value(v)
	}

	return driverArgs
}

// ValidateUserID
var _ = Describe("ValidateUserID", func() {
	Context("with valid integer strings", func() {
		It("should return int64 for valid string", func() {
			id, err := ValidateUserID("12345")
			Expect(err).To(BeNil())
			Expect(id).To(Equal(int64(12345)))
		})
	})

	Context("with invalid strings", func() {
		It("should return an error for non-numeric input", func() {
			_, err := ValidateUserID("abc")
			Expect(err).To(HaveOccurred())
		})

		It("should return an error for float-like strings", func() {
			_, err := ValidateUserID("12.34")
			Expect(err).To(HaveOccurred())
		})

		It("should return an error for empty string", func() {
			_, err := ValidateUserID("")
			Expect(err).To(HaveOccurred())
		})
	})
})

// ValidateNewUserRequest
var _ = Describe("ValidateNewUserRequest", func() {
	var user User

	BeforeEach(func() {
		user = User{
			UserName:   "jdoe",
			FirstName:  "John",
			LastName:   "Doe",
			Email:      "john@example.com",
			UserStatus: "A",
		}
	})

	It("should validate a complete and valid user", func() {
		err := user.ValidateNewUserRequest()
		Expect(err).To(BeNil())
	})

	It("should return error when user_name is missing", func() {
		user.UserName = ""
		err := user.ValidateNewUserRequest()
		Expect(err).To(Equal(ErrMissingUserName))
	})

	It("should return error when first_name is missing", func() {
		user.FirstName = ""
		err := user.ValidateNewUserRequest()
		Expect(err).To(Equal(ErrMissingFirstName))
	})

	It("should return error when last_name is missing", func() {
		user.LastName = ""
		err := user.ValidateNewUserRequest()
		Expect(err).To(Equal(ErrMissingLastName))
	})

	It("should return error when email is missing", func() {
		user.Email = ""
		err := user.ValidateNewUserRequest()
		Expect(err).To(Equal(ErrMissingEmail))
	})

	It("should allow department to be nil", func() {
		user.Department = nil
		err := user.ValidateNewUserRequest()
		Expect(err).To(BeNil())
	})
})

// ConvertToUserDB
var _ = Describe("ConvertToUserDB", func() {
	var input User
	var expected db.UserDB

	BeforeEach(func() {
		dept := "Engineering"
		input = User{
			ID:         1,
			UserName:   "jdoe",
			FirstName:  "John",
			LastName:   "Doe",
			Email:      "john@example.com",
			UserStatus: "A",
			Department: &dept,
		}

		expected = db.UserDB{
			UserID:     1,
			UserName:   "jdoe",
			FirstName:  "John",
			LastName:   "Doe",
			Email:      "john@example.com",
			UserStatus: "A",
			Department: sql.NullString{String: "Engineering", Valid: true},
		}
	})

	It("should convert all fields including department", func() {
		result := input.ConvertToUserDB()
		Expect(result).To(Equal(expected))
	})

	It("should handle nil department", func() {
		input.Department = nil
		result := input.ConvertToUserDB()
		Expect(result.Department.Valid).To(BeFalse())
		Expect(result.Department.String).To(BeEmpty())
	})
})

// ConvertToUser
var _ = Describe("ConvertToUser", func() {
	var input db.UserDB
	var expected User

	Context("when Department is valid", func() {
		BeforeEach(func() {
			input = db.UserDB{
				UserID:     10,
				UserName:   "asmith",
				FirstName:  "Alice",
				LastName:   "Smith",
				Email:      "alice@example.com",
				UserStatus: "A",
				Department: sql.NullString{String: "HR", Valid: true},
			}

			expected = User{
				ID:         10,
				UserName:   "asmith",
				FirstName:  "Alice",
				LastName:   "Smith",
				Email:      "alice@example.com",
				UserStatus: "A",
				Department: ptr("HR"),
			}
		})

		It("should convert all fields correctly including department", func() {
			result := ConvertToUser(&input)
			Expect(result).To(Equal(expected))
		})
	})

	Context("when Department is NULL", func() {
		BeforeEach(func() {
			input = db.UserDB{
				UserID:     11,
				UserName:   "bwayne",
				FirstName:  "Bruce",
				LastName:   "Wayne",
				Email:      "bruce@wayneenterprises.com",
				UserStatus: "I",
				Department: sql.NullString{Valid: false},
			}

			expected = User{
				ID:         11,
				UserName:   "bwayne",
				FirstName:  "Bruce",
				LastName:   "Wayne",
				Email:      "bruce@wayneenterprises.com",
				UserStatus: "I",
				Department: nil,
			}
		})

		It("should convert all fields and omit department", func() {
			result := ConvertToUser(&input)
			Expect(result).To(Equal(expected))
			Expect(result.Department).To(BeNil())
		})
	})
})

// GetAllUsers
var _ = Describe("GetAllUsers", func() {
	var (
		mockDB *sql.DB
		mock   sqlmock.Sqlmock
		err    error
	)

	BeforeEach(func() {
		mockDB, mock, err = sqlmock.New()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		mockDB.Close()
	})

	It("returns all users on success", func() {
		columns := []string{
			"user_id", "user_name", "first_name", "last_name", "email", "user_status", "department",
		}

		mockRows := sqlmock.NewRows(columns).AddRow(
			1, "jdoe", "John", "Doe", "jdoe@example.com", "A", sql.NullString{String: "IT", Valid: true},
		).AddRow(
			2, "asmith", "Alice", "Smith", "asmith@example.com", "I", sql.NullString{Valid: false},
		)

		query, args, buildErr := sq.Select(db.AllColumns).From(DbName).ToSql()
		Expect(buildErr).To(BeNil())

		driverArgs := convertToDriverArgs(args)

		mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(mockRows)

		users, err := GetAllUsers(mockDB)
		Expect(err).ToNot(HaveOccurred())
		Expect(users).To(HaveLen(2))

		Expect(users[0].UserName).To(Equal("jdoe"))
		Expect(users[0].Department).ToNot(BeNil())
		Expect(*users[0].Department).To(Equal("IT"))

		Expect(users[1].UserName).To(Equal("asmith"))
		Expect(users[1].Department).To(BeNil())
	})

	It("returns error on query failure", func() {
		query, args, buildErr := sq.Select(db.AllColumns).From(DbName).ToSql()
		Expect(buildErr).To(BeNil())

		driverArgs := convertToDriverArgs(args)

		mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnError(errors.New("query failed"))

		users, err := GetAllUsers(mockDB)
		Expect(err).To(HaveOccurred())
		Expect(users).To(BeNil())
	})

	It("returns error on row scan failure", func() {
		columns := []string{
			"user_id", "user_name", "first_name", "last_name", "email", "user_status", "department",
		}

		// Invalid data type to trigger scan error
		mockRows := sqlmock.NewRows(columns).AddRow(
			"not_an_int", "jdoe", "John", "Doe", "jdoe@example.com", "A", sql.NullString{String: "IT", Valid: true},
		)

		query, args, buildErr := sq.Select(db.AllColumns).From(DbName).ToSql()
		Expect(buildErr).To(BeNil())

		driverArgs := convertToDriverArgs(args)

		mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(mockRows)

		users, err := GetAllUsers(mockDB)
		Expect(err).To(HaveOccurred())
		Expect(users).To(BeNil())
	})

	It("returns error when rows iteration has an error", func() {
		columns := []string{
			"user_id", "user_name", "first_name", "last_name", "email", "user_status", "department",
		}

		mockRows := sqlmock.NewRows(columns).
			AddRow(1, "jdoe", "John", "Doe", "jdoe@example.com", "A", sql.NullString{String: "IT", Valid: true}).
			RowError(0, errors.New("row iteration error"))

		query, args, buildErr := sq.Select(db.AllColumns).From(DbName).ToSql()
		Expect(buildErr).To(BeNil())

		driverArgs := convertToDriverArgs(args)

		mock.ExpectQuery(query).WithArgs(driverArgs...).WillReturnRows(mockRows)

		users, err := GetAllUsers(mockDB)
		Expect(err).To(HaveOccurred())
		Expect(users).To(BeNil())
	})
})

// GetUser
var _ = Describe("GetUser", func() {
	var (
		mockDB   *sql.DB
		mock     sqlmock.Sqlmock
		userID   int64
		expected *User
	)

	BeforeEach(func() {
		var err error
		mockDB, mock, err = sqlmock.New()
		Expect(err).To(BeNil())

		userID = 1
		dept := "Engineering"
		expected = &User{
			ID:         userID,
			UserName:   "jdoe",
			FirstName:  "John",
			LastName:   "Doe",
			Email:      "jdoe@example.com",
			UserStatus: "A",
			Department: &dept,
		}
	})

	AfterEach(func() {
		mock.ExpectClose()
		mockDB.Close()
	})

	It("should return user when found", func() {
		query, args, err := sq.Select(db.AllColumns).
			From(DbName).
			Where(sq.Eq{"user_id": userID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		Expect(err).To(BeNil())

		driverArgs := convertToDriverArgs(args)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{
				"user_id", "user_name", "first_name", "last_name", "email", "user_status", "department",
			}).AddRow(
				expected.ID, expected.UserName, expected.FirstName, expected.LastName,
				expected.Email, expected.UserStatus, expected.Department,
			))

		user, err := GetUser(mockDB, userID)
		Expect(err).To(BeNil())
		Expect(user).To(Equal(expected))
	})

	It("should return error if user not found", func() {
		query, args, err := sq.Select(db.AllColumns).
			From(DbName).
			Where(sq.Eq{"user_id": userID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		Expect(err).To(BeNil())

		driverArgs := convertToDriverArgs(args)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnError(sql.ErrNoRows)

		user, err := GetUser(mockDB, userID)
		Expect(user).To(BeNil())
		Expect(err).To(Equal(ErrUserNotFound))
	})

	It("should return error if scan fails", func() {
		query, args, err := sq.Select(db.AllColumns).
			From(DbName).
			Where(sq.Eq{"user_id": userID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		Expect(err).To(BeNil())

		driverArgs := convertToDriverArgs(args)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("invalid"))

		user, err := GetUser(mockDB, userID)
		Expect(user).To(BeNil())
		Expect(err).To(HaveOccurred())
	})

	It("should return user with nil Department if DB returns NULL", func() {
		query, args, err := sq.Select(db.AllColumns).
			From(DbName).
			Where(sq.Eq{"user_id": userID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		Expect(err).To(BeNil())

		driverArgs := convertToDriverArgs(args)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{
				"user_id", "user_name", "first_name", "last_name", "email", "user_status", "department",
			}).AddRow(
				userID, "jdoe", "John", "Doe", "jdoe@example.com", "A", nil, // department is NULL
			))

		user, err := GetUser(mockDB, userID)
		Expect(err).To(BeNil())
		Expect(user).ToNot(BeNil())
		Expect(user.Department).To(BeNil())
	})

	It("should return ErrMissingUserID when id == 0", func() {
		user, err := GetUser(mockDB, 0)
		Expect(user).To(BeNil())
		Expect(err).To(Equal(ErrMissingUserID))
	})

})

// CheckUserNameExists
var _ = Describe("CheckUserNameExists", func() {
	var (
		mockDB   *sql.DB
		mock     sqlmock.Sqlmock
		userName string
	)

	BeforeEach(func() {
		var err error
		mockDB, mock, err = sqlmock.New()
		Expect(err).To(BeNil())
		userName = "jdoe"
	})

	AfterEach(func() {
		mock.ExpectClose()
		mockDB.Close()
	})

	It("should return true when username exists", func() {
		query, args, err := sq.Select("COUNT(*)").
			From(DbName).
			Where(sq.Eq{"user_name": userName}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		Expect(err).To(BeNil())

		driverArgs := convertToDriverArgs(args)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		exists, err := CheckUserNameExists(mockDB, userName)
		Expect(err).To(BeNil())
		Expect(exists).To(BeTrue())
	})

	It("should return false when username does not exist", func() {
		query, args, err := sq.Select("COUNT(*)").
			From(DbName).
			Where(sq.Eq{"user_name": userName}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		Expect(err).To(BeNil())

		driverArgs := convertToDriverArgs(args)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		exists, err := CheckUserNameExists(mockDB, userName)
		Expect(err).To(BeNil())
		Expect(exists).To(BeFalse())
	})

	It("should return error if scan fails", func() {
		query, args, err := sq.Select("COUNT(*)").
			From(DbName).
			Where(sq.Eq{"user_name": userName}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		Expect(err).To(BeNil())

		driverArgs := convertToDriverArgs(args)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow("invalid"))

		exists, err := CheckUserNameExists(mockDB, userName)
		Expect(err).To(HaveOccurred())
		Expect(exists).To(BeFalse())
	})

	It("should return error if username is empty", func() {
		exists, err := CheckUserNameExists(mockDB, "")
		Expect(err).To(Equal(ErrMissingUserName))
		Expect(exists).To(BeFalse())
	})
})

// CreateNewUser
var _ = Describe("CreateNewUser", func() {
	var (
		mockDB *sql.DB
		mock   sqlmock.Sqlmock
		user   *User
	)

	BeforeEach(func() {
		var err error
		mockDB, mock, err = sqlmock.New()
		Expect(err).To(BeNil())

		dept := "Engineering"
		user = &User{
			UserName:   "jdoe",
			FirstName:  "John",
			LastName:   "Doe",
			Email:      "jdoe@example.com",
			UserStatus: "A",
			Department: &dept,
		}
	})

	AfterEach(func() {
		mock.ExpectClose()
		mockDB.Close()
	})

	It("should insert new user and return user with ID", func() {
		// Expect the CheckUserNameExists subquery
		checkQuery, checkArgs, err := sq.Select("COUNT(*)").
			From(DbName).
			Where(sq.Eq{"user_name": user.UserName}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		Expect(err).To(BeNil())

		driverArgs := convertToDriverArgs(checkArgs)

		mock.ExpectQuery(regexp.QuoteMeta(checkQuery)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		// Expect the INSERT query
		userDB := user.ConvertToUserDB()
		insertQuery, insertArgs, err := sq.Insert(DbName).
			Columns("user_name", "first_name", "last_name", "email", "user_status", "department").
			Values(userDB.UserName, userDB.FirstName, userDB.LastName, userDB.Email, userDB.UserStatus, userDB.Department).
			Suffix("RETURNING user_id").
			PlaceholderFormat(sq.Dollar).
			ToSql()
		Expect(err).To(BeNil())

		driverArgs = convertToDriverArgs(insertArgs)

		mock.ExpectQuery(regexp.QuoteMeta(insertQuery)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(123))

		createdUser, err := CreateUser(mockDB, user)
		Expect(err).To(BeNil())
		Expect(createdUser.ID).To(Equal(int64(123)))
		Expect(createdUser.UserName).To(Equal("jdoe"))
	})

	It("should return error if username already exists", func() {
		checkQuery, checkArgs, err := sq.Select("COUNT(*)").
			From(DbName).
			Where(sq.Eq{"user_name": user.UserName}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		Expect(err).To(BeNil())

		driverArgs := convertToDriverArgs(checkArgs)

		mock.ExpectQuery(regexp.QuoteMeta(checkQuery)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		newUser, err := CreateUser(mockDB, user)
		Expect(err).To(Equal(ErrUserExists))
		Expect(newUser).To(BeNil())
	})

	It("should return error on insert scan failure", func() {
		// Username does not exist
		checkQuery, checkArgs, _ := sq.Select("COUNT(*)").
			From(DbName).
			Where(sq.Eq{"user_name": user.UserName}).
			PlaceholderFormat(sq.Dollar).
			ToSql()

		driverArgs := convertToDriverArgs(checkArgs)

		mock.ExpectQuery(regexp.QuoteMeta(checkQuery)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		// Insert query
		userDB := user.ConvertToUserDB()
		insertQuery, insertArgs, _ := sq.Insert(DbName).
			Columns("user_name", "first_name", "last_name", "email", "user_status", "department").
			Values(userDB.UserName, userDB.FirstName, userDB.LastName, userDB.Email, userDB.UserStatus, userDB.Department).
			Suffix("RETURNING user_id").
			PlaceholderFormat(sq.Dollar).
			ToSql()

		driverArgs = convertToDriverArgs(insertArgs)

		mock.ExpectQuery(regexp.QuoteMeta(insertQuery)).
			WithArgs(driverArgs...).
			WillReturnError(sql.ErrConnDone) // simulate scan or connection failure

		newUser, err := CreateUser(mockDB, user)
		Expect(err).To(HaveOccurred())
		Expect(newUser).To(BeNil())
	})
})

// UpdateUser
var _ = Describe("UpdateUser", func() {
	var (
		mockDB *sql.DB
		mock   sqlmock.Sqlmock
		user   *User
	)

	BeforeEach(func() {
		var err error
		mockDB, mock, err = sqlmock.New()
		Expect(err).To(BeNil())

		dept := "Engineering"
		user = &User{
			ID:         1,
			UserName:   "jdoe",
			FirstName:  "John",
			LastName:   "Doe",
			Email:      "jdoe@example.com",
			UserStatus: "A",
			Department: &dept,
		}
	})

	AfterEach(func() {
		Expect(mock.ExpectationsWereMet()).To(Succeed())
		mockDB.Close()
	})

	It("successfully updates and returns user", func() {
		updateQuery, updateArgs, _ := sq.Update(DbName).
			SetMap(map[string]interface{}{
				"user_name":   user.UserName,
				"first_name":  user.FirstName,
				"last_name":   user.LastName,
				"email":       user.Email,
				"user_status": user.UserStatus,
				"department":  user.Department,
			}).
			Where(sq.Eq{"user_id": user.ID}).
			Suffix("RETURNING user_id").
			PlaceholderFormat(sq.Dollar).
			ToSql()

		driverArgs := convertToDriverArgs(updateArgs)

		mock.ExpectExec(regexp.QuoteMeta(updateQuery)).
			WithArgs(driverArgs...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// GetUser is called after update
		selectQuery, selectArgs, _ := sq.Select(db.AllColumns).
			From(DbName).
			Where(sq.Eq{"user_id": user.ID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()

		driverArgs = convertToDriverArgs(selectArgs)

		mock.ExpectQuery(regexp.QuoteMeta(selectQuery)).
			WithArgs(driverArgs...).
			WillReturnRows(sqlmock.NewRows([]string{
				"user_id", "user_name", "first_name", "last_name", "email", "user_status", "department",
			}).AddRow(user.ID, user.UserName, user.FirstName, user.LastName, user.Email, user.UserStatus, *user.Department))

		updatedUser, err := UpdateUser(mockDB, user)
		Expect(err).To(BeNil())
		Expect(updatedUser.ID).To(Equal(user.ID))
	})

	It("returns error when scan in GetUser fails", func() {
		updateQuery, updateArgs, _ := sq.Update(DbName).
			SetMap(map[string]interface{}{
				"user_name":   user.UserName,
				"first_name":  user.FirstName,
				"last_name":   user.LastName,
				"email":       user.Email,
				"user_status": user.UserStatus,
				"department":  user.Department,
			}).
			Where(sq.Eq{"user_id": user.ID}).
			Suffix("RETURNING user_id").
			PlaceholderFormat(sq.Dollar).
			ToSql()

		driverArgs := convertToDriverArgs(updateArgs)

		mock.ExpectExec(regexp.QuoteMeta(updateQuery)).
			WithArgs(driverArgs...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Simulate scan failure on GetUser
		selectQuery, selectArgs, _ := sq.Select(db.AllColumns).
			From(DbName).
			Where(sq.Eq{"user_id": user.ID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()

		driverArgs = convertToDriverArgs(selectArgs)

		mock.ExpectQuery(regexp.QuoteMeta(selectQuery)).
			WithArgs(driverArgs...).
			WillReturnError(sql.ErrConnDone) // simulate failure in QueryRow().Scan()

		updatedUser, err := UpdateUser(mockDB, user)
		Expect(err).To(HaveOccurred())
		Expect(updatedUser).To(BeNil())
	})
})

// DeleteUser
var _ = Describe("DeleteUser", func() {
	var (
		mockDB *sql.DB
		mock   sqlmock.Sqlmock
		userID int64 = 1
	)

	BeforeEach(func() {
		var err error
		mockDB, mock, err = sqlmock.New()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(mock.ExpectationsWereMet()).To(Succeed())
		mockDB.Close()
	})

	It("successfully deletes a user", func() {
		delQuery, delArgs, _ := sq.Delete(DbName).
			Where(sq.Eq{"user_id": userID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()

		driverArgs := convertToDriverArgs(delArgs)

		mock.ExpectExec(regexp.QuoteMeta(delQuery)).
			WithArgs(driverArgs...).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

		err := DeleteUser(mockDB, userID)
		Expect(err).To(BeNil())
	})

	It("returns ErrUserNotFound when no rows affected", func() {
		delQuery, delArgs, _ := sq.Delete(DbName).
			Where(sq.Eq{"user_id": userID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()

		driverArgs := convertToDriverArgs(delArgs)

		mock.ExpectExec(regexp.QuoteMeta(delQuery)).
			WithArgs(driverArgs...).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

		err := DeleteUser(mockDB, userID)
		Expect(err).To(MatchError(ErrUserNotFound))
	})

	It("returns error if Exec fails", func() {
		delQuery, delArgs, _ := sq.Delete(DbName).
			Where(sq.Eq{"user_id": userID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()

		driverArgs := convertToDriverArgs(delArgs)

		mock.ExpectExec(regexp.QuoteMeta(delQuery)).
			WithArgs(driverArgs...).
			WillReturnError(fmt.Errorf("exec failure"))

		err := DeleteUser(mockDB, userID)
		Expect(err).To(MatchError("exec failure"))
	})

	It("returns error if RowsAffected fails", func() {
		delQuery, delArgs, _ := sq.Delete(DbName).
			Where(sq.Eq{"user_id": userID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()

		driverArgs := convertToDriverArgs(delArgs)

		// Create a result that returns error on RowsAffected()
		_ = sqlmock.NewResult(0, 1)
		// Wrap result to simulate RowsAffected error
		mock.ExpectExec(regexp.QuoteMeta(delQuery)).
			WithArgs(driverArgs...).
			WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected failure")))

		err := DeleteUser(mockDB, userID)
		Expect(err).To(MatchError("rows affected failure"))
	})
})

// GetByID
