package helper

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"final-task/model"

	"github.com/golang-jwt/jwt/v5"
)

func NewJWTManager(jwtConfig model.JWTConfig) *model.JWTManager {
	return &model.JWTManager{Config: jwtConfig}
}

func RefreshTTL(jwtManager *model.JWTManager) time.Duration {
	return jwtManager.Config.RefreshTokenTTL
}

func GenerateAccessToken(jwtManager *model.JWTManager, userID int64) (string, error) {
	generatedToken, err := generateToken(jwtManager, userID, model.TokenTypeAccess, "", jwtManager.Config.AccessTokenTTL)
	if err != nil {
		return "", err
	}

	return ProtectToken(generatedToken), nil
}

func GenerateRefreshToken(jwtManager *model.JWTManager, userID int64) (token string, tokenID string, err error) {
	tokenID, err = newTokenID()
	if err != nil {
		return "", "", err
	}

	token, err = generateToken(jwtManager, userID, model.TokenTypeRefresh, tokenID, jwtManager.Config.RefreshTokenTTL)
	if err != nil {
		return "", "", err
	}

	return ProtectToken(token), tokenID, nil
}

func VerifyToken(jwtManager *model.JWTManager, tokenString string, expectedTokenType string) (*model.CustomClaims, error) {
	rawToken, err := UnprotectToken(tokenString)
	if err != nil {
		return nil, err
	}

	parsedToken, err := jwt.ParseWithClaims(rawToken, &model.CustomClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New(model.MessageUnexpectedSigning)
		}
		return []byte(jwtManager.Config.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, isValidClaimType := parsedToken.Claims.(*model.CustomClaims)
	if !isValidClaimType || !parsedToken.Valid {
		return nil, errors.New(model.MessageInvalidToken)
	}

	if claims.TokenType != expectedTokenType {
		return nil, errors.New(model.MessageInvalidTokenType)
	}

	return claims, nil
}

func generateToken(
	jwtManager *model.JWTManager,
	userID int64,
	tokenType string,
	tokenID string,
	tokenTTL time.Duration,
) (string, error) {
	currentTime := time.Now()

	customClaims := model.CustomClaims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Issuer:    jwtManager.Config.Issuer,
			Subject:   strconv.FormatInt(userID, 10),
			ExpiresAt: jwt.NewNumericDate(currentTime.Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(currentTime),
			NotBefore: jwt.NewNumericDate(currentTime),
		},
	}

	generatedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	return generatedToken.SignedString([]byte(jwtManager.Config.SecretKey))
}

func newTokenID() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}

func ProtectToken(token string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(token))
}

func UnprotectToken(encodedToken string) (string, error) {
	decodedToken, err := base64.RawURLEncoding.DecodeString(encodedToken)
	if err == nil {
		return string(decodedToken), nil
	}

	decodedToken, err = base64.URLEncoding.DecodeString(encodedToken)
	if err == nil {
		return string(decodedToken), nil
	}

	decodedToken, err = base64.StdEncoding.DecodeString(encodedToken)
	if err == nil {
		return string(decodedToken), nil
	}

	return "", errors.New(model.MessageInvalidToken)
}
