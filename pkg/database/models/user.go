package models

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/phoenix-industries/omega-server/pkg/auth"
	"github.com/phoenix-industries/omega-server/pkg/database"
	"github.com/phoenix-industries/omega-server/pkg/validate"
)

type UserGender string

const (
	GenderMale   UserGender = "male"
	GenderFemale UserGender = "female"
)

func (g UserGender) String() string {
	return string(g)
}

func (g UserGender) Validate() error {
	if g != GenderMale && g != GenderFemale {
		return errors.New("invalid gender")
	}
	return nil
}

type User struct {
	Model
	Name        string     `db:"name" json:"name"`
	Role        auth.Role  `db:"role" json:"role"`
	Email       string     `db:"email" json:"email"`
	Phone       *string    `db:"phone" json:"phone,omitempty"`
	Password    string     `db:"password" json:"-"`
	Gender      UserGender `db:"gender" json:"gender"`
	PictureID   *string    `db:"picture_id" json:"picture_id"`
	City        *string    `db:"city" json:"city,omitempty"`
	Governorate *string    `db:"governorate" json:"governorate,omitempty"`
	Address     *string    `db:"address" json:"address,omitempty"`
	Birthdate   time.Time  `db:"birthdate" json:"birthdate"`
}

func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("name is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Phone != nil {
		if *u.Phone == "" {
			return errors.New("phone is required")
		}
		if err := validate.PhoneNumber(*u.Phone); err != nil {
			return err
		}
	}
	if u.Role == "" {
		return errors.New("role is required")
	}
	if u.Birthdate.IsZero() {
		return errors.New("birthdate is required")
	}
	if err := u.Gender.Validate(); err != nil {
		return err
	}
	if err := validate.Email(u.Email); err != nil {
		return err
	}
	return nil
}

func UserInsert(ctx context.Context, db database.DB, user *User) error {
	if user.ID == "" {
		return errors.New("id is not set")
	}
	if user.Role == "" {
		user.Role = auth.RoleMember
	}
	if err := user.Validate(); err != nil {
		return err
	}
	query := `
		INSERT INTO users
		(id, name, email, phone, role, gender, picture_id, city, governorate, address, password, birthdate)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := db.Exec(ctx, query, user.ID, user.Name, user.Email, user.Phone, user.Role, user.Gender, user.PictureID, user.City, user.Governorate, user.Address, user.Password, user.Birthdate)
	return err
}

func UserGetByID(ctx context.Context, db database.DB, id string) (*User, error) {
	query := `
		SELECT *
		FROM users
		WHERE id = $1 AND deleted_at IS null
	`
	var user User
	if err := pgxscan.Get(ctx, db, &user, query, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func UserGetAll(ctx context.Context, db database.DB, limit int, offset int) ([]User, error) {
	query := `
		SELECT *
		FROM users
		WHERE deleted_at IS null
		ORDER BY created_at DESC
		LIMIT $1
		OFFSET $2
	`
	var users []User
	if err := pgxscan.Select(ctx, db, &users, query, limit, offset); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return users, nil
}

func UserGetByEmail(ctx context.Context, db database.DB, email string) (*User, error) {
	query := `
		SELECT *
		FROM users
		WHERE email = $1 AND deleted_at IS null
	`
	var user User
	if err := pgxscan.Get(ctx, db, &user, query, email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func UserGetByPhone(ctx context.Context, db database.DB, phone string) (*User, error) {
	query := `
		SELECT *
		FROM users
		WHERE phone = $1 AND deleted_at IS null
	`
	var user User
	if err := pgxscan.Get(ctx, db, &user, query, phone); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func UserExistsWithID(ctx context.Context, db database.DB, id string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE id = $1 AND deleted_at IS null
		)
	`
	var exists bool
	if err := pgxscan.Get(ctx, db, &exists, query, id); err != nil {
		return false, err
	}
	return exists, nil
}

func UserExistsWithEmail(ctx context.Context, db database.DB, email string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE email = $1 AND deleted_at IS null
		)
	`
	var exists bool
	if err := pgxscan.Get(ctx, db, &exists, query, email); err != nil {
		return false, err
	}
	return exists, nil
}

func UserExistsWithPhone(ctx context.Context, db database.DB, phone string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE phone = $1 AND deleted_at IS null
		)
	`
	var exists bool
	if err := pgxscan.Get(ctx, db, &exists, query, phone); err != nil {
		return false, err
	}
	return exists, nil
}

func UserUpdate(ctx context.Context, db database.DB, user *User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	query := `
		UPDATE users
		SET name = $2, email = $3, phone = $4, city = $5, governorate = $6,
			address = $7, birthdate = $8, picture_id = $9, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS null
	`
	if _, err := db.Exec(ctx, query, user.ID, user.Name, user.Email, user.Phone, user.City, user.Governorate, user.Address, user.Birthdate, user.PictureID); err != nil {
		return err
	}
	return nil
}

func UserUpdatePassword(ctx context.Context, db database.DB, id, password string) error {
	query := `
		UPDATE users
		SET password = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS null
	`
	_, err := db.Exec(ctx, query, id, password)
	return err
}
