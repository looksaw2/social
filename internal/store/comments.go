package store

import (
	"context"
	"database/sql"
	"time"
)

// Comments的模型
type Comment struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	//链接的外表
	User User `json:"user"`
}

// Comments的存储
type CommentsStore struct {
	db *sql.DB
}

func (s *CommentsStore) GetPostByID(ctx context.Context, postID int64) ([]Comment, error) {
	//和user连表查询
	query :=
		`
		SELECT c.id,c.post_id,c.user_id,c.content,c.created_at,c.updated_at,users.username,users.id FROM comments c
		JOIN users on users.id = c.user_id
		WHERE c.post_id = $1
		ORDER BY c.created_at DESC;
	`
	// 开始查询
	rows, err := s.db.QueryContext(
		ctx,
		query,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	//return的结果
	comments := []Comment{}
	//将得到的结果遍历到数组
	for rows.Next() {
		var c Comment
		c.User = User{}
		err := rows.Scan(
			&c.ID,
			&c.PostID,
			&c.UserID,
			&c.Content,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.User.Username,
			&c.User.ID,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil

}

// 创建评论
func (s *CommentsStore) Create(ctx context.Context, comment *Comment) error {
	//创建的SQL语句
	query := `
		INSERT INTO comments (post_id, user_id, content)
		VALUES( $1, $2 , $3)
		RETURNING id , created_at, updated_at
	`
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	//执行SQL
	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.PostID,
		comment.UserID,
		comment.Content,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}
