package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const userIDKey contextKey = "userID"

// Init reads or generates the JWT secret from the given path.
func Init(secretPath string) error {
	if err := os.MkdirAll(filepath.Dir(secretPath), 0700); err != nil {
		return fmt.Errorf("creating secret dir: %w", err)
	}
	data, err := os.ReadFile(secretPath)
	if err == nil && len(data) == 32 {
		secretMu.Lock()
		secret = data
		secretMu.Unlock()
		return nil
	}
	newSecret := make([]byte, 32)
	if _, err := rand.Read(newSecret); err != nil {
		return fmt.Errorf("generating secret: %w", err)
	}
	if err := os.WriteFile(secretPath, newSecret, 0600); err != nil {
		return fmt.Errorf("writing secret: %w", err)
	}
	secretMu.Lock()
	secret = newSecret
	secretMu.Unlock()
	return nil
}

// WithUserID stores userID in the request context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext retrieves the userID stored by WithUserID.
func UserIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey).(string)
	return id
}

var (
	ErrInvalidToken    = errors.New("invalid token")
	ErrExpiredToken    = errors.New("expired token")
	ErrTokenGeneration = errors.New("failed to generate token")
	ErrRateLimited     = errors.New("rate limited")
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

var (
	secret   []byte
	secretMu sync.RWMutex
)

type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*attemptTracker
}

type attemptTracker struct {
	count    int
	blocked  bool
	firstTry time.Time
}

var limiter = &rateLimiter{
	attempts: make(map[string]*attemptTracker),
}

const (
	accessTokenTTL  = 1 * time.Hour
	refreshTokenTTL = 7 * 24 * time.Hour
	maxAttempts     = 5
	windowDuration  = 10 * time.Minute
)

// readOrGenerateSecret returns the JWT secret set by Init. Init must be called
// once at startup (main.go) before any login/token-validation request arrives.
func readOrGenerateSecret() []byte {
	secretMu.RLock()
	defer secretMu.RUnlock()
	return secret
}

func GenerateTokenPair(userID string) (access, refresh string, err error) {
	sec := readOrGenerateSecret()
	if sec == nil || len(sec) != 32 {
		return "", "", ErrTokenGeneration
	}

	var wg sync.WaitGroup
	var accessErr, refreshErr error

	wg.Add(2)
	go func() {
		defer wg.Done()
		access, accessErr = generateToken(userID, accessTokenTTL, sec)
	}()
	go func() {
		defer wg.Done()
		refresh, refreshErr = generateToken(userID, refreshTokenTTL, sec)
	}()
	wg.Wait()

	if accessErr != nil || refreshErr != nil {
		return "", "", ErrTokenGeneration
	}
	return access, refresh, nil
}

func generateToken(userID string, ttl time.Duration, secret []byte) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ValidateToken(tokenString string) (string, error) {
	sec := readOrGenerateSecret()
	if sec == nil || len(sec) != 32 {
		return "", ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return sec, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrExpiredToken
		}
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}

	return claims.UserID, nil
}

func HashPassword(plain string) (string, error) {
	if len(plain) == 0 {
		return "", errors.New("password cannot be empty")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

func CheckPassword(hash, plain string) bool {
	if len(hash) == 0 || len(plain) == 0 {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}

func CheckRateLimit(ip string) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	entry, exists := limiter.attempts[ip]
	if !exists {
		return true
	}

	if entry.blocked {
		return false
	}

	if time.Since(entry.firstTry) > windowDuration {
		delete(limiter.attempts, ip)
		return true
	}

	return entry.count < maxAttempts
}

func RecordFailedAttempt(ip string) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	now := time.Now()
	entry, exists := limiter.attempts[ip]
	if !exists {
		limiter.attempts[ip] = &attemptTracker{
			count:    1,
			firstTry: now,
		}
		return
	}

	if time.Since(entry.firstTry) > windowDuration {
		entry.count = 1
		entry.firstTry = now
		entry.blocked = false
		return
	}

	entry.count++
	if entry.count >= maxAttempts {
		entry.blocked = true
	}
}
