عالیه—بریم نسخه‌ی سینیور mini-twitter رو با **لایک، کامنت، فالو** و **پروفایل** کامل کنیم. زیر هر بخش فایل‌های «جدید/آپدیت‌شده» رو می‌بینی. همه چیز با همان لایه‌بندی (handler → service → repository) و کانتکست‌محور پیاده شده.

> نکته: در `auth.go` یک باگ ایمپورت داشتیم (استفاده از `fmt.Sscanf` بدون `import fmt`) و یک ایمپورت بلااستفاده. همزمان درستش کردم و یک کم به هدر Authorization سفت‌وسخت‌تر نگاه می‌کنیم.

---

# ✅ مایگریشن‌های جدید

### `migrations/0002_social.sql` (جدید)

```sql
-- Profiles
CREATE TABLE profiles (
  user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  name TEXT,
  bio TEXT,
  avatar_url TEXT,
  location TEXT,
  website TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Likes
CREATE TABLE tweet_likes (
  user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  tweet_id BIGINT REFERENCES tweets(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ DEFAULT now(),
  PRIMARY KEY (user_id, tweet_id)
);
CREATE INDEX idx_tweet_likes_tweet ON tweet_likes(tweet_id);

-- Comments
CREATE TABLE comments (
  id BIGSERIAL PRIMARY KEY,
  tweet_id BIGINT REFERENCES tweets(id) ON DELETE CASCADE,
  user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  body TEXT NOT NULL CHECK (char_length(body) <= 280),
  created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_comments_tweet ON comments(tweet_id);
CREATE INDEX idx_comments_user ON comments(user_id);

-- Follows
CREATE TABLE follows (
  follower_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  followee_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ DEFAULT now(),
  PRIMARY KEY (follower_id, followee_id),
  CHECK (follower_id <> followee_id)
);
CREATE INDEX idx_follows_followee ON follows(followee_id);
```

---

# 🧱 مدل‌ها

### `internal/model/profile.go` (جدید)

```go
package model

import "time"

type Profile struct {
	UserID    int64     `db:"user_id" json:"user_id"`
	Name      *string   `db:"name" json:"name,omitempty"`
	Bio       *string   `db:"bio" json:"bio,omitempty"`
	AvatarURL *string   `db:"avatar_url" json:"avatar_url,omitempty"`
	Location  *string   `db:"location" json:"location,omitempty"`
	Website   *string   `db:"website" json:"website,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type UpsertProfileRequest struct {
	Name      *string `json:"name"`
	Bio       *string `json:"bio"`
	AvatarURL *string `json:"avatar_url"`
	Location  *string `json:"location"`
	Website   *string `json:"website"`
}
```

### `internal/model/social.go` (جدید)

```go
package model

import "time"

type Like struct {
	UserID  int64     `db:"user_id" json:"user_id"`
	TweetID int64     `db:"tweet_id" json:"tweet_id"`
	Created time.Time `db:"created_at" json:"created_at"`
}

type Comment struct {
	ID        int64     `db:"id" json:"id"`
	TweetID   int64     `db:"tweet_id" json:"tweet_id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Body      string    `db:"body" json:"body"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type CreateCommentRequest struct {
	Body string `json:"body"`
}
```

### آپدیت روی توییت برای ویو غنی

#### `internal/model/tweet.go` (آپدیت افزایشی)

```go
package model

import "time"

type Tweet struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Body      string    `db:"body" json:"body"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// نمای غنی برای خروجی API
type TweetView struct {
	Tweet
	LikeCount    int64 `db:"like_count" json:"like_count"`
	CommentCount int64 `db:"comment_count" json:"comment_count"`
	// وقتی کاربر لاگین است
	LikedByViewer bool `db:"liked_by_viewer" json:"liked_by_viewer"`
}
```

---

# 🧭 ریپازیتوری‌ها

### `internal/repository/tweet_repo.go` (آپدیت)

```go
package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/example/mini-twitter/internal/model"
)

type TweetRepo interface {
	Create(ctx context.Context, t *model.Tweet) error
	List(ctx context.Context, limit, offset int) ([]*model.Tweet, error)
	GetByID(ctx context.Context, id int64) (*model.Tweet, error)
	// لیست با متادیتا برای نمایش
	ListWithMeta(ctx context.Context, viewerID *int64, limit, offset int) ([]*model.TweetView, error)
}

type tweetRepo struct{
	db *sqlx.DB
}

func NewTweetRepo(db *sqlx.DB) TweetRepo {
	return &tweetRepo{db: db}
}

func (r *tweetRepo) Create(ctx context.Context, t *model.Tweet) error {
	query := `INSERT INTO tweets (user_id, body) VALUES ($1, $2) RETURNING id, created_at`
	return r.db.QueryRowxContext(ctx, query, t.UserID, t.Body).Scan(&t.ID, &t.CreatedAt)
}

func (r *tweetRepo) List(ctx context.Context, limit, offset int) ([]*model.Tweet, error) {
	var tweets []*model.Tweet
	err := r.db.SelectContext(ctx, &tweets, `SELECT * FROM tweets ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	return tweets, err
}

func (r *tweetRepo) GetByID(ctx context.Context, id int64) (*model.Tweet, error) {
	t := &model.Tweet{}
	err := r.db.GetContext(ctx, t, `SELECT * FROM tweets WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *tweetRepo) ListWithMeta(ctx context.Context, viewerID *int64, limit, offset int) ([]*model.TweetView, error) {
	var out []*model.TweetView
	base := `
SELECT 
  t.id, t.user_id, t.body, t.created_at,
  COALESCE(lc.cnt,0) AS like_count,
  COALESCE(cc.cnt,0) AS comment_count,
  CASE WHEN $3 IS NULL THEN false
       ELSE EXISTS(SELECT 1 FROM tweet_likes tl WHERE tl.tweet_id = t.id AND tl.user_id = $3)
  END AS liked_by_viewer
FROM tweets t
LEFT JOIN (
  SELECT tweet_id, COUNT(*) AS cnt FROM tweet_likes GROUP BY tweet_id
) lc ON lc.tweet_id = t.id
LEFT JOIN (
  SELECT tweet_id, COUNT(*) AS cnt FROM comments GROUP BY tweet_id
) cc ON cc.tweet_id = t.id
ORDER BY t.created_at DESC
LIMIT $1 OFFSET $2
`
	err := r.db.SelectContext(ctx, &out, base, limit, offset, viewerID)
	return out, err
}
```

### `internal/repository/social_repo.go` (جدید)

```go
package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/example/mini-twitter/internal/model"
)

type LikeRepo interface {
	Like(ctx context.Context, userID, tweetID int64) error
	Unlike(ctx context.Context, userID, tweetID int64) error
	CountByTweet(ctx context.Context, tweetID int64) (int64, error)
}

type CommentRepo interface {
	Create(ctx context.Context, c *model.Comment) error
	ListByTweet(ctx context.Context, tweetID int64, limit, offset int) ([]*model.Comment, error)
	Delete(ctx context.Context, id int64, ownerID int64) (bool, error) // true if deleted
	CountByTweet(ctx context.Context, tweetID int64) (int64, error)
}

type FollowRepo interface {
	Follow(ctx context.Context, followerID, followeeID int64) error
	Unfollow(ctx context.Context, followerID, followeeID int64) error
	Counts(ctx context.Context, userID int64) (followers, following int64, err error)
	IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error)
}

type likeRepo struct{ db *sqlx.DB }
type commentRepo struct{ db *sqlx.DB }
type followRepo struct{ db *sqlx.DB }

func NewLikeRepo(db *sqlx.DB) LikeRepo       { return &likeRepo{db: db} }
func NewCommentRepo(db *sqlx.DB) CommentRepo { return &commentRepo{db: db} }
func NewFollowRepo(db *sqlx.DB) FollowRepo   { return &followRepo{db: db} }

func (r *likeRepo) Like(ctx context.Context, userID, tweetID int64) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO tweet_likes (user_id, tweet_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, tweetID)
	return err
}
func (r *likeRepo) Unlike(ctx context.Context, userID, tweetID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tweet_likes WHERE user_id=$1 AND tweet_id=$2`, userID, tweetID)
	return err
}
func (r *likeRepo) CountByTweet(ctx context.Context, tweetID int64) (int64, error) {
	var c int64
	err := r.db.GetContext(ctx, &c, `SELECT COUNT(*) FROM tweet_likes WHERE tweet_id=$1`, tweetID)
	return c, err
}

func (r *commentRepo) Create(ctx context.Context, c *model.Comment) error {
	q := `INSERT INTO comments (tweet_id, user_id, body) VALUES ($1,$2,$3) RETURNING id, created_at`
	return r.db.QueryRowxContext(ctx, q, c.TweetID, c.UserID, c.Body).Scan(&c.ID, &c.CreatedAt)
}
func (r *commentRepo) ListByTweet(ctx context.Context, tweetID int64, limit, offset int) ([]*model.Comment, error) {
	var out []*model.Comment
	err := r.db.SelectContext(ctx, &out, `SELECT * FROM comments WHERE tweet_id=$1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`, tweetID, limit, offset)
	return out, err
}
func (r *commentRepo) Delete(ctx context.Context, id int64, ownerID int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM comments WHERE id=$1 AND user_id=$2`, id, ownerID)
	if err != nil { return false, err }
	aff, _ := res.RowsAffected()
	return aff > 0, nil
}
func (r *commentRepo) CountByTweet(ctx context.Context, tweetID int64) (int64, error) {
	var c int64
	err := r.db.GetContext(ctx, &c, `SELECT COUNT(*) FROM comments WHERE tweet_id=$1`, tweetID)
	return c, err
}

func (r *followRepo) Follow(ctx context.Context, followerID, followeeID int64) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO follows (follower_id, followee_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, followerID, followeeID)
	return err
}
func (r *followRepo) Unfollow(ctx context.Context, followerID, followeeID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM follows WHERE follower_id=$1 AND followee_id=$2`, followerID, followeeID)
	return err
}
func (r *followRepo) Counts(ctx context.Context, userID int64) (int64, int64, error) {
	var followers, following int64
	if err := r.db.GetContext(ctx, &followers, `SELECT COUNT(*) FROM follows WHERE followee_id=$1`, userID); err != nil { return 0,0, err }
	if err := r.db.GetContext(ctx, &following, `SELECT COUNT(*) FROM follows WHERE follower_id=$1`, userID); err != nil { return 0,0, err }
	return followers, following, nil
}
func (r *followRepo) IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id=$1 AND followee_id=$2)`, followerID, followeeID)
	return exists, err
}
```

### `internal/repository/profile_repo.go` (جدید)

```go
package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/example/mini-twitter/internal/model"
)

type ProfileRepo interface {
	Upsert(ctx context.Context, p *model.Profile) error
	Get(ctx context.Context, userID int64) (*model.Profile, error)
	GetByUsername(ctx context.Context, username string) (*model.Profile, *int64, error) // + user_id
}

type profileRepo struct{ db *sqlx.DB }

func NewProfileRepo(db *sqlx.DB) ProfileRepo { return &profileRepo{db: db} }

func (r *profileRepo) Upsert(ctx context.Context, p *model.Profile) error {
	q := `
INSERT INTO profiles (user_id, name, bio, avatar_url, location, website)
VALUES (:user_id, :name, :bio, :avatar_url, :location, :website)
ON CONFLICT (user_id) DO UPDATE
SET name = EXCLUDED.name,
    bio = EXCLUDED.bio,
    avatar_url = EXCLUDED.avatar_url,
    location = EXCLUDED.location,
    website = EXCLUDED.website,
    updated_at = now()
`
	_, err := r.db.NamedExecContext(ctx, q, p)
	return err
}

func (r *profileRepo) Get(ctx context.Context, userID int64) (*model.Profile, error) {
	p := &model.Profile{}
	err := r.db.GetContext(ctx, p, `SELECT * FROM profiles WHERE user_id=$1`, userID)
	if err != nil { return nil, err }
	return p, nil
}

func (r *profileRepo) GetByUsername(ctx context.Context, username string) (*model.Profile, *int64, error) {
	var uid int64
	if err := r.db.GetContext(ctx, &uid, `SELECT id FROM users WHERE username=$1`, username); err != nil {
		return nil, nil, err
	}
	p, err := r.Get(ctx, uid)
	if err != nil {
		// اگر پروفایل هنوز ساخته نشده بود، خود uid را برگردانیم
		return nil, &uid, nil
	}
	return p, &uid, nil
}
```

---

# 🧠 سرویس‌ها

### `internal/service/social_service.go` (جدید)

```go
package service

import (
	"context"
	"errors"

	"github.com/example/mini-twitter/internal/model"
	"github.com/example/mini-twitter/internal/repository"
)

var (
	ErrForbidden = errors.New("forbidden")
)

type SocialService interface {
	Like(ctx context.Context, userID, tweetID int64) error
	Unlike(ctx context.Context, userID, tweetID int64) error
	AddComment(ctx context.Context, userID, tweetID int64, body string) (*model.Comment, error)
	ListComments(ctx context.Context, tweetID int64, limit, offset int) ([]*model.Comment, error)
	DeleteComment(ctx context.Context, userID, commentID int64) error
}

type socialService struct {
	likes    repository.LikeRepo
	comments repository.CommentRepo
	tweets   repository.TweetRepo
}

func NewSocialService(l repository.LikeRepo, c repository.CommentRepo, t repository.TweetRepo) SocialService {
	return &socialService{likes: l, comments: c, tweets: t}
}

func (s *socialService) Like(ctx context.Context, userID, tweetID int64) error {
	if _, err := s.tweets.GetByID(ctx, tweetID); err != nil { return err }
	return s.likes.Like(ctx, userID, tweetID)
}
func (s *socialService) Unlike(ctx context.Context, userID, tweetID int64) error {
	return s.likes.Unlike(ctx, userID, tweetID)
}
func (s *socialService) AddComment(ctx context.Context, userID, tweetID int64, body string) (*model.Comment, error) {
	if len(body) == 0 || len(body) > 280 { return nil, ErrTweetTooLong }
	if _, err := s.tweets.GetByID(ctx, tweetID); err != nil { return nil, err }
	c := &model.Comment{TweetID: tweetID, UserID: userID, Body: body}
	if err := s.comments.Create(ctx, c); err != nil { return nil, err }
	return c, nil
}
func (s *socialService) ListComments(ctx context.Context, tweetID int64, limit, offset int) ([]*model.Comment, error) {
	return s.comments.ListByTweet(ctx, tweetID, limit, offset)
}
func (s *socialService) DeleteComment(ctx context.Context, userID, commentID int64) error {
	ok, err := s.comments.Delete(ctx, commentID, userID)
	if err != nil { return err }
	if !ok { return ErrForbidden }
	return nil
}
```

### `internal/service/follow_service.go` (جدید)

```go
package service

import (
	"context"

	"github.com/example/mini-twitter/internal/repository"
)

type FollowService interface {
	Follow(ctx context.Context, followerID, followeeID int64) error
	Unfollow(ctx context.Context, followerID, followeeID int64) error
	Counts(ctx context.Context, userID int64) (followers, following int64, err error)
	IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error)
}

type followService struct{ follows repository.FollowRepo }

func NewFollowService(fr repository.FollowRepo) FollowService { return &followService{follows: fr} }

func (s *followService) Follow(ctx context.Context, followerID, followeeID int64) error {
	return s.follows.Follow(ctx, followerID, followeeID)
}
func (s *followService) Unfollow(ctx context.Context, followerID, followeeID int64) error {
	return s.follows.Unfollow(ctx, followerID, followeeID)
}
func (s *followService) Counts(ctx context.Context, userID int64) (int64, int64, error) {
	return s.follows.Counts(ctx, userID)
}
func (s *followService) IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error) {
	return s.follows.IsFollowing(ctx, followerID, followeeID)
}
```

### `internal/service/profile_service.go` (جدید)

```go
package service

import (
	"context"

	"github.com/example/mini-twitter/internal/model"
	"github.com/example/mini-twitter/internal/repository"
)

type ProfileService interface {
	Upsert(ctx context.Context, userID int64, req *model.UpsertProfileRequest) (*model.Profile, error)
	GetByUsername(ctx context.Context, username string, viewerID *int64) (*model.Profile, int64, int64, int64, bool, error)
}

type profileService struct{
	profiles repository.ProfileRepo
	follows  repository.FollowRepo
}

func NewProfileService(p repository.ProfileRepo, f repository.FollowRepo) ProfileService {
	return &profileService{profiles: p, follows: f}
}

func (s *profileService) Upsert(ctx context.Context, userID int64, req *model.UpsertProfileRequest) (*model.Profile, error) {
	p := &model.Profile{
		UserID: userID,
		Name: req.Name, Bio: req.Bio, AvatarURL: req.AvatarURL, Location: req.Location, Website: req.Website,
	}
	if err := s.profiles.Upsert(ctx, p); err != nil { return nil, err }
	return s.profiles.Get(ctx, userID)
}

func (s *profileService) GetByUsername(ctx context.Context, username string, viewerID *int64) (*model.Profile, int64, int64, int64, bool, error) {
	p, uidPtr, err := s.profiles.GetByUsername(ctx, username)
	if err != nil { return nil, 0,0,0,false, err }
	var uid int64
	if uidPtr != nil { uid = *uidPtr }
	followers, following, err := s.follows.Counts(ctx, uid)
	if err != nil { return p, 0,0,0,false, err }
	isFollowing := false
	if viewerID != nil {
		isFollowing, _ = s.follows.IsFollowing(ctx, *viewerID, uid)
	}
	return p, uid, followers, following, isFollowing, nil
}
```

---

# 🌐 هندلرها

### کمک‌های مشترک

#### `internal/handler/context.go` (جدید)

```go
package handler

import "context"

type ctxKey string

const userIDKey ctxKey = "user_id"

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
func UserIDFrom(ctx context.Context) (int64, bool) {
	v := ctx.Value(userIDKey)
	if v == nil { return 0, false }
	id, ok := v.(int64)
	return id, ok
}
```

### `internal/handler/auth.go` (فیکس کوچک)

```go
package handler

import (
	"context"
	"encoding/json"
	"fmt" // ← اضافه شد
	"net/http"

	"github.com/example/mini-twitter/internal/model"
	"github.com/example/mini-twitter/internal/service"
)

type AuthHandler struct{
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	u, err := h.svc.Register(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(u)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	token, err := h.svc.Login(r.Context(), &req)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *AuthHandler) JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		var token string
		if n, _ := fmt.Sscanf(authHeader, "Bearer %s", &token); n != 1 || token == "" {
			http.Error(w, "missing or invalid token", http.StatusUnauthorized)
			return
		}
		userID, err := h.svc.ValidateToken(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		ctx := WithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

### `internal/handler/social.go` (جدید: لایک/آن‌لایک و کامنت)

```go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/example/mini-twitter/internal/model"
	"github.com/example/mini-twitter/internal/service"
)

type SocialHandler struct{
	s service.SocialService
}

func NewSocialHandler(s service.SocialService) *SocialHandler { return &SocialHandler{s: s} }

func (h *SocialHandler) Like(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFrom(r.Context()); if !ok { http.Error(w, "unauthenticated", http.StatusUnauthorized); return }
	tid, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64); if err != nil { http.Error(w, "bad id", http.StatusBadRequest); return }
	if err := h.s.Like(r.Context(), uid, tid); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	w.WriteHeader(http.StatusNoContent)
}

func (h *SocialHandler) Unlike(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFrom(r.Context()); if !ok { http.Error(w, "unauthenticated", http.StatusUnauthorized); return }
	tid, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64); if err != nil { http.Error(w, "bad id", http.StatusBadRequest); return }
	if err := h.s.Unlike(r.Context(), uid, tid); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	w.WriteHeader(http.StatusNoContent)
}

func (h *SocialHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFrom(r.Context()); if !ok { http.Error(w, "unauthenticated", http.StatusUnauthorized); return }
	tid, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64); if err != nil { http.Error(w, "bad id", http.StatusBadRequest); return }
	var req model.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "bad request", http.StatusBadRequest); return }
	c, err := h.s.AddComment(r.Context(), uid, tid, req.Body)
	if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(c)
}

func (h *SocialHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64); if err != nil { http.Error(w, "bad id", http.StatusBadRequest); return }
	limit, offset := parseLimitOffset(r, 20, 0)
	list, err := h.s.ListComments(r.Context(), tid, limit, offset)
	if err != nil { http.Error(w, "failed", http.StatusInternalServerError); return }
	_ = json.NewEncoder(w).Encode(list)
}

func (h *SocialHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFrom(r.Context()); if !ok { http.Error(w, "unauthenticated", http.StatusUnauthorized); return }
	cid, err := strconv.ParseInt(chi.URLParam(r, "commentID"), 10, 64); if err != nil { http.Error(w, "bad id", http.StatusBadRequest); return }
	if err := h.s.DeleteComment(r.Context(), uid, cid); err != nil { http.Error(w, err.Error(), http.StatusForbidden); return }
	w.WriteHeader(http.StatusNoContent)
}

func parseLimitOffset(r *http.Request, defL, defO int) (int, int) {
	q := r.URL.Query()
	limit, offset := defL, defO
	if v := q.Get("limit"); v != "" { if n, err := strconv.Atoi(v); err == nil { limit = n } }
	if v := q.Get("offset"); v != "" { if n, err := strconv.Atoi(v); err == nil { offset = n } }
	return limit, offset
}
```

### `internal/handler/follow.go` (جدید)

```go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/example/mini-twitter/internal/service"
)

type FollowHandler struct{ s service.FollowService }

func NewFollowHandler(s service.FollowService) *FollowHandler { return &FollowHandler{s: s} }

func (h *FollowHandler) Follow(w http.ResponseWriter, r *http.Request) {
	followerID, ok := UserIDFrom(r.Context()); if !ok { http.Error(w, "unauthenticated", http.StatusUnauthorized); return }
	uid, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64); if err != nil { http.Error(w, "bad id", http.StatusBadRequest); return }
	if err := h.s.Follow(r.Context(), followerID, uid); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	w.WriteHeader(http.StatusNoContent)
}
func (h *FollowHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	followerID, ok := UserIDFrom(r.Context()); if !ok { http.Error(w, "unauthenticated", http.StatusUnauthorized); return }
	uid, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64); if err != nil { http.Error(w, "bad id", http.StatusBadRequest); return }
	if err := h.s.Unfollow(r.Context(), followerID, uid); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	w.WriteHeader(http.StatusNoContent)
}

func (h *FollowHandler) Counts(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64); if err != nil { http.Error(w, "bad id", http.StatusBadRequest); return }
	followers, following, err := h.s.Counts(r.Context(), uid)
	if err != nil { http.Error(w, "failed", http.StatusInternalServerError); return }
	_ = json.NewEncoder(w).Encode(map[string]int64{
		"followers": followers,
		"following": following,
	})
}
```

### `internal/handler/profile.go` (جدید)

```go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/example/mini-twitter/internal/model"
	"github.com/example/mini-twitter/internal/service"
)

type ProfileHandler struct{ s service.ProfileService }

func NewProfileHandler(s service.ProfileService) *ProfileHandler { return &ProfileHandler{s: s} }

func (h *ProfileHandler) UpsertMe(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFrom(r.Context()); if !ok { http.Error(w, "unauthenticated", http.StatusUnauthorized); return }
	var req model.UpsertProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "bad request", http.StatusBadRequest); return }
	p, err := h.s.Upsert(r.Context(), uid, &req)
	if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
	_ = json.NewEncoder(w).Encode(p)
}

func (h *ProfileHandler) GetByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	var viewerID *int64
	if id, ok := UserIDFrom(r.Context()); ok { viewerID = &id }
	p, uid, followers, following, isFollowing, err := h.s.GetByUsername(r.Context(), username, viewerID)
	if err != nil { http.Error(w, "not found", http.StatusNotFound); return }
	resp := map[string]any{
		"user_id": uid, "profile": p, "followers": followers, "following": following, "is_following": isFollowing,
	}
	_ = json.NewEncoder(w).Encode(resp)
}
```

### `internal/handler/tweet.go` (آپدیت: لیست غنی)

```go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/example/mini-twitter/internal/model"
	"github.com/example/mini-twitter/internal/service"
)

type TweetHandler struct{
	svc service.TweetService
}

func NewTweetHandler(svc service.TweetService) *TweetHandler {
	return &TweetHandler{svc: svc}
}

func (h *TweetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTweetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid, ok := UserIDFrom(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	t, err := h.svc.Create(r.Context(), uid, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(t)
}

func (h *TweetHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r, 20, 0)
	var viewer *int64
	if id, ok := UserIDFrom(r.Context()); ok { viewer = &id }
	list, err := h.svc.ListWithMeta(r.Context(), viewer, limit, offset)
	if err != nil {
		http.Error(w, "failed to list tweets", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(list)
}
```

---

# 🧩 سرویس توییت (آپدیت برای ویو غنی)

### `internal/service/tweet_service.go` (آپدیت)

```go
package service

import (
	"context"
	"errors"

	"github.com/example/mini-twitter/internal/model"
	"github.com/example/mini-twitter/internal/repository"
)

var (
	ErrTweetTooLong = errors.New("tweet too long")
)

type TweetService interface {
	Create(ctx context.Context, userID int64, req *model.CreateTweetRequest) (*model.Tweet, error)
	List(ctx context.Context, limit, offset int) ([]*model.Tweet, error)
	ListWithMeta(ctx context.Context, viewerID *int64, limit, offset int) ([]*model.TweetView, error)
}

type tweetService struct{
	tweets repository.TweetRepo
}

func NewTweetService(t repository.TweetRepo) TweetService {
	return &tweetService{tweets: t}
}

func (s *tweetService) Create(ctx context.Context, userID int64, req *model.CreateTweetRequest) (*model.Tweet, error) {
	if len(req.Body) == 0 {
		return nil, errors.New("empty body")
	}
	if len(req.Body) > 280 {
		return nil, ErrTweetTooLong
	}
	t := &model.Tweet{UserID: userID, Body: req.Body}
	if err := s.tweets.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *tweetService) List(ctx context.Context, limit, offset int) ([]*model.Tweet, error) {
	return s.tweets.List(ctx, limit, offset)
}

func (s *tweetService) ListWithMeta(ctx context.Context, viewerID *int64, limit, offset int) ([]*model.TweetView, error) {
	return s.tweets.ListWithMeta(ctx, viewerID, limit, offset)
}
```

---

# 🚏 سرور و روت‌ها

### `internal/server/server.go` (آپدیت)

```go
package server

import (
	"time"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"

	"github.com/example/mini-twitter/internal/handler"
	"github.com/example/mini-twitter/internal/repository"
	"github.com/example/mini-twitter/internal/service"
	"github.com/example/mini-twitter/pkg/config"
)

func New(db *sqlx.DB, cfg *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// repos
	userRepo := repository.NewUserRepo(db)
	tweetRepo := repository.NewTweetRepo(db)
	likeRepo := repository.NewLikeRepo(db)
	commentRepo := repository.NewCommentRepo(db)
	followRepo := repository.NewFollowRepo(db)
	profileRepo := repository.NewProfileRepo(db)

	// services
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiry)
	tweetSvc := service.NewTweetService(tweetRepo)
	socialSvc := service.NewSocialService(likeRepo, commentRepo, tweetRepo)
	followSvc := service.NewFollowService(followRepo)
	profileSvc := service.NewProfileService(profileRepo, followRepo)

	// handlers
	authH := handler.NewAuthHandler(authSvc)
	tweetH := handler.NewTweetHandler(tweetSvc)
	socialH := handler.NewSocialHandler(socialSvc)
	followH := handler.NewFollowHandler(followSvc)
	profileH := handler.NewProfileHandler(profileSvc)

	// public
	r.Route("/api", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)

		// timeline (public ولی اگر توکن بده، liked_by_viewer پر می‌شود)
		r.Get("/tweets", tweetH.List)

		// پروفایل پابلیک با یوزرنیم
		r.Get("/profiles/{username}", profileH.GetByUsername)
		// کامنت‌ها پابلیک لیست می‌شوند
		r.Get("/tweets/{id:[0-9]+}/comments", socialH.ListComments)

		// protected
		r.Group(func(r chi.Router) {
			r.Use(authH.JWTAuthMiddleware)

			// tweeting
			r.Post("/tweets", tweetH.Create)

			// likes
			r.Post("/tweets/{id:[0-9]+}/like", socialH.Like)
			r.Delete("/tweets/{id:[0-9]+}/like", socialH.Unlike)

			// comments
			r.Post("/tweets/{id:[0-9]+}/comments", socialH.AddComment)
			r.Delete("/comments/{commentID:[0-9]+}", socialH.DeleteComment)

			// follow
			r.Post("/users/{userID:[0-9]+}/follow", followH.Follow)
			r.Delete("/users/{userID:[0-9]+}/follow", followH.Unfollow)
			r.Get("/users/{userID:[0-9]+}/follow_counts", followH.Counts)

			// profile (خودم)
			r.Put("/me/profile", profileH.UpsertMe)
		})
	})

	return r
}
```

---

# 🧪 نکات اجرای سریع

1. مایگریشن جدید را اجرا کن:

   ```bash
   psql "$DATABASE_URL" < migrations/0002_social.sql
   ```
2. سرور را ران کن:

   ```bash
   make run
   ```
3. اندپوینت‌های جدید/به‌روز:

   * `GET  /api/tweets?limit&offset` → لیست با `like_count, comment_count, liked_by_viewer`
   * `POST /api/tweets/{id}/like` / `DELETE /api/tweets/{id}/like`
   * `GET  /api/tweets/{id}/comments` (public)
   * `POST /api/tweets/{id}/comments` (protected)
   * `DELETE /api/comments/{commentID}` (مالک کامنت)
   * `POST /api/users/{userID}/follow` / `DELETE /api/users/{userID}/follow`
   * `GET  /api/users/{userID}/follow_counts`
   * `PUT  /api/me/profile`
   * `GET  /api/profiles/{username}` → `{ user_id, profile, followers, following, is_following }`

---

## ایده‌های بهبود بعدی (در مسیر Production)

* **Transaction/Outbox** برای شمارش‌های سنگین (اگر نیاز شد)، یا **materialized view**‌های دوره‌ای.
* **Cursor-based pagination** برای کامنت‌ها و تایم‌لاین.
* **Etag/Conditional Requests** برای پروفایل و توییت‌ها.
* **Validation** دقیق‌تر (مثلاً trim/normalize ورودی‌ها).
* **Policy layer** (e.g., دوستان فقط، بلاک‌کردن کاربران).
* **Indexes پوششی** روی `(tweet_id, created_at)` برای کامنت‌ها.

---

اگر بخوای، می‌تونم همین تغییرات رو به صورت یک ریپو کامل (با `docker-compose` برای Postgres، تست‌های واحد/اینtegration با `testcontainers`, و نمونه ریکوئست‌های `httpie`/`restclient`) بسازم تا یک‌کلیکه بیاد بالا. کد دیگری هم می‌خوای اضافه کنیم (مثلاً ری‌توییت، منشن، یا نوتیفیکیشن)، بگو انجامش بدم.
