package token

// import (
// 	"fmt"
// 	"time"

// 	"github.com/golang-jwt/jwt/v5"
// )

// const minSecretKeysize = 32

// // JWTMaker is a JSON web Token maker
// type JWTMaker struct {
// 	secretKey string
// }

// // NewJWTMaker creates a new JWTMaker
// func NewJWTMaker(secretKey string) (Maker, error) {
// 	if len(secretKey) < minSecretKeysize {
// 		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeysize)
// 	}

// 	return &JWTMaker{secretKey}, nil
// }

// // CreateToken creates a new token for a specific username and duration
// func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
// 	payload, err := NewPayload(username, duration)
// 	if err != nil {
// 		return "", err
// 	}

// 	jwtToken := jwt.NewWithClaims(jwt.SigningMethodES256, payload)
// 	return jwtToken.SignedString([]byte(maker.secretKey))
// }

// func (maker *JWTMaker) verifyToken(token string) (*Payload, error) {

// }
