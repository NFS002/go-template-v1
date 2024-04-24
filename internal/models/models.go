package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// DBModel is the type for database connection values
type DBModel struct {
	DB *sql.DB
}

// Models is the wrapper for all models
type Models struct {
	DB DBModel
}

// NewModels returns a model type with database connection pool
func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{
			DB: db,
		},
	}
}

// User is the type for all users
type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Scope     string    `json:"scope"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (u User) CanRequestScope(requestedScope []string) error {
	userScope := strings.Split(u.Scope, ",")
	for _, rs := range requestedScope {
		if !slices.Contains(userScope, rs) {
			return fmt.Errorf("requested scope '%s' is invalid for user", rs)
		}
	}
	return nil
}

// GetUserByEmail gets a user by email address
func (m *DBModel) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	email = strings.ToLower(email)

	var u User

	stmt := `
		select
		    id, first_name, last_name, email, password, scope, created_at, updated_at 
		from 
			users
		where email = $1
	`

	row := m.DB.QueryRowContext(ctx, stmt, email)

	// the *Row's Scan will return ErrNoRows.
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.Scope,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return u, errors.New("invalid credentials")
	}

	return u, nil
}

func (m *DBModel) Authenticate(email, password string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	email = strings.ToLower(email)

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "select id, password from users where email = ?", email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, errors.New("incorrect password")
	} else if err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (m *DBModel) UpdatePasswordForUser(userId int, u UpdateUserRequest, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update users set password = $1 where id = $2`

	_, err := m.DB.ExecContext(ctx, stmt, hash, userId)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) GetAllUsers() ([]*GetUserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var users []*GetUserResponse

	query := `
		select
			first_name, last_name, email, created_at
		from
			users
		order by
			last_name, first_name
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u GetUserResponse
		err = rows.Scan(
			&u.FirstName,
			&u.LastName,
			&u.Email,
			&u.CreatedAt)

		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	return users, nil
}

func (m *DBModel) GetOneUser(id int) (GetUserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u GetUserResponse

	query := `
		select
			first_name, last_name, email, created_at
		from
			users
		where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.CreatedAt)

	if err != nil {
		return u, err
	}
	return u, nil
}

func (m *DBModel) EditUser(userId int, u UpdateUserRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// If at least 1 field set
	stmt := `UPDATE users 
		SET first_name = COALESCE(NULLIF($1, ''), first_name),
		last_name = COALESCE(NULLIF($2, ''), last_name),
		email = COALESCE(NULLIF($3, ''), email)
		WHERE id = $4`

	res, err := m.DB.ExecContext(ctx, stmt,
		u.FirstName,
		u.LastName,
		u.Email,
		userId)

	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("user not found")
	}

	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) AddUser(u User, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	scope := "read:a,write:a,read:b,write:b"
	stmt := `
		insert into users (first_name, last_name, email, password, scope)
		values ($1, $2, $3, $4, $5)`

	_, err := m.DB.ExecContext(ctx, stmt,
		u.FirstName,
		u.LastName,
		u.Email,
		hash,
		scope)

	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) DeleteUser(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// foreign key constraint (on delete cascade) will also
	// delete any associated tokens with the user
	stmt := `delete from users where id = $1`

	res, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("user not found")
	}

	return nil

}
