package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User模型
type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	IsActive  bool     `json:"is_active"`
	RoleID    int64    `json:"role_id"`
	Role      Role     `json:"role"`
}

// password 结构体
type password struct {
	text *string //原始的文本
	hash []byte
}

// 给text加密
func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.text = &text
	p.hash = hash
	return nil
}

// User存储
type UserStore struct {
	db *sql.DB
}

// Create方法实现
func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query :=
		`
		INSERT INTO users (username , email , password ,role_id)
		VALUES ($1,$2,$3,$4) RETURNING id,created_at,updated_at
	`
	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash,
		user.RoleID,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint :users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}
	return nil
}

// 实现GetByID方法
func (s *UserStore) GetByID(ctx context.Context, userID int64) (*User, error) {
	//SQL语句
	query :=
		`
		SELECT users.id , username , email , password , created_at , updated_at,roles.*
		FROM users
		JOIN roles ON (users.role_id = roles.id)
		WHERE users.id = $1	AND is_active = true
	`
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	//需要返回的user
	user := &User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Description,
		&user.Role.Level,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

// 创建用户并且发送邀请
func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	//wapper事务
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		//创建用户
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}
		//创建用户并且邀请
		if err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}
		return nil
	})

}

func (s *UserStore) createUserInvitation(
	ctx context.Context,
	tx *sql.Tx,
	token string,
	exp time.Duration,
	userID int64,
) error {
	query := `
		INSERT INTO user_invitations (token ,user_id , expiry)  VALUES ($1,$2,$3)
	`
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, QueryDuration)
	defer cancel()
	//执行Query
	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}
	return nil
}

// 激活用户
func (s *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		//得到user
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}
		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}
		if err := s.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query :=
		`
		SELECT u.id ,u.username , u.email , u.created_at , u.updated_at , u.is_active
		FROM users u
		JOIN user_invitations ui ON u.id = ui.user_id
		WHERE ui.token = $1 AND ui.expiry > $2
	`
	//hash the token
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, QueryDuration)
	defer cancel()

	user := &User{}
	err := tx.QueryRowContext(
		ctx,
		query,
		hashToken,
		time.Now(),
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil

}

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		UPDATE users
		SET username = $1,
			email = $2,
			is_active = $3
		WHERE id = $4
	`
	ctx, cancel := context.WithTimeout(ctx, QueryDuration)
	defer cancel()
	_, err := tx.ExecContext(
		ctx,
		query,
		&user.Username,
		&user.Email,
		&user.IsActive,
		&user.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, QueryDuration)
	defer cancel()
	_, err := tx.ExecContext(
		ctx,
		query,
		userID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) Delete(ctx context.Context, userID int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.delete(ctx, tx, userID); err != nil {
			return err
		}
		if err := s.deleteUserInvitations(ctx, tx, userID); err != nil {
			return err
		}
		return nil
	})
}
func (s *UserStore) delete(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM users WHERE id = $1`
	ctx, cancel := context.WithTimeout(ctx, QueryDuration)
	defer cancel()
	_, err := tx.ExecContext(
		ctx,
		query,
		userID,
	)
	if err != nil {
		return err
	}
	return nil
}

// 通过email得到User
func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query :=
		`
		SELECT id ,username,email,password,created_at,updated_at
		FROM users
		WHERE email = $1 AND is_active = true	
	`
	user := &User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil

}
