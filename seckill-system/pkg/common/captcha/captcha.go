package captcha

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	TTL       = 5 * time.Minute
	keyPrefix = "auth:captcha:"
	alphabet  = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
)

type Challenge struct {
	ID        string `json:"captcha_id"`
	Image     string `json:"captcha_image"`
	ExpiresIn int    `json:"expires_in"`
}

func Generate(ctx context.Context, rdb *redis.Client) (*Challenge, error) {
	code, err := randomCode(5)
	if err != nil {
		return nil, err
	}

	id := uuid.NewString()
	if err := rdb.Set(ctx, keyPrefix+id, hash(code), TTL).Err(); err != nil {
		return nil, err
	}

	return &Challenge{
		ID:        id,
		Image:     "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg(code))),
		ExpiresIn: int(TTL.Seconds()),
	}, nil
}

func Validate(ctx context.Context, rdb *redis.Client, id, answer string) error {
	id = strings.TrimSpace(id)
	answer = strings.ToUpper(strings.TrimSpace(answer))
	if id == "" || answer == "" {
		return fmt.Errorf("captcha is required")
	}

	key := keyPrefix + id
	expected, err := rdb.GetDel(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("captcha expired or invalid")
	}
	if err != nil {
		return err
	}
	if expected != hash(answer) {
		return fmt.Errorf("captcha expired or invalid")
	}
	return nil
}

func randomCode(length int) (string, error) {
	var b strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b.WriteByte(alphabet[n.Int64()])
	}
	return b.String(), nil
}

func hash(value string) string {
	sum := sha256.Sum256([]byte(strings.ToUpper(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:])
}

func svg(code string) string {
	escaped := html.EscapeString(code)
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="140" height="44" viewBox="0 0 140 44">
<rect width="140" height="44" rx="6" fill="#f7fafc"/>
<path d="M8 12 C38 2, 62 44, 132 12" stroke="#94a3b8" stroke-width="1.4" fill="none"/>
<path d="M10 34 C42 18, 84 54, 130 24" stroke="#c084fc" stroke-width="1.2" fill="none"/>
<text x="70" y="29" text-anchor="middle" font-family="Verdana,Arial,sans-serif" font-size="22" font-weight="700" letter-spacing="4" fill="#111827">%s</text>
</svg>`, escaped)
}
