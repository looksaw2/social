package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

// post模型
type Post struct {
	ID        int64    `json:"id"`
	Content   string   `json:"content"`
	Title     string   `json:"title"`
	UserID    int64    `json:"user_id"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`

	//连接的外表
	Comments []Comment `json:"comment"`
	User     User      `json:"user"`
	//乐观锁
	Version int64 `json:"version"`
}

// Post的元数据
type PostWithMetadata struct {
	Post
	CommentCount int `json:"comments_count"`
}

// post存储
type PostStore struct {
	db *sql.DB
}

// 实现Create接口
func (s *PostStore) Create(ctx context.Context, post *Post) error {
	//SQL语句
	query := `
		INSERT INTO posts(content , title , user_id , tags)
		VALUES($1,$2,$3,$4) RETURNING id ,created_at ,updated_at
	`
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, QueryDuration)
	defer cancel()
	//开始查询
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

// 实现getById接口 posts
func (s *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	query :=
		`
		SELECT id , user_id , title , created_at , updated_at , tags ,version
		FROM posts WHERE id = $1
	`
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, QueryDuration)
	defer cancel()
	//执行
	var post Post
	err := s.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags),
		&post.Version,
	)
	//错误处理
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, err
		default:
			return nil, err
		}
	}
	return &post, nil
}

// Delete方法
func (s *PostStore) Delete(ctx context.Context, postID int64) error {
	//请求
	query :=
		`
		DELETE FROM posts WHERE id = $1
	`
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	//执行删除操作
	res, err := s.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}
	//查看删除了多少行
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	//删除0行，没有找到
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// Patch方法
func (s *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1 , content = $2 , version = version + 1
		WHERE id = $3 AND version = $4
		RETURNING version
	`
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, QueryDuration)
	defer cancel()
	//执行Patch操作
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.ID,
		post.Version).Scan(
		&post.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}

// 实现接口
func (s *PostStore) GetUserFeed(ctx context.Context, userID int64, fq PaginationFeedQuery) ([]PostWithMetadata, error) {
	query := `
	SELECT 
		p.id,
		p.user_id,
		p.title,
		p.content,
		p.created_at,
		p.updated_at,
		p.version,
		p.tags,
		u.username,
		COUNT(c.id) AS comments_count
	FROM posts p
	LEFT JOIN comments c ON c.post_id = p.id
	LEFT JOIN users u ON p.user_id = u.id
	JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
	WHERE f.user_id = $1 OR p.user_id = $1 AND 
		  (p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')  AND
		  (p.tags @> $5 OR $5 = '{}' )
	GROUP BY p.id , u.username
	ORDER BY p.created_at ` + fq.Sort + `
	LIMIT $2 OFFSET $3
	`
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	//执行查询
	rows, err := s.db.QueryContext(
		ctx,
		query,
		userID,
		fq.Limit,
		fq.Offset,
		fq.Search,
		pq.Array(fq.Tags),
	)
	if err != nil {
		return nil, err
	}
	//关闭资源
	defer rows.Close()
	var feed []PostWithMetadata
	//开始遍历
	for rows.Next() {
		var post PostWithMetadata
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Version,
			pq.Array(&post.Tags),
			&post.User.Username,
			&post.CommentCount,
		)
		if err != nil {
			return nil, err
		}
		feed = append(feed, post)
	}
	return feed, nil
}
