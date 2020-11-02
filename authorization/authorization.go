package authorization

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	jwtV4 "github.com/dgrijalva/jwt-go/v4"
)

type Response struct {
	Message string `json:"message"`
}

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

//CreateAuthorizationMiddleware creates authorization middleware with the given parameters
func CreateAuthorizationMiddleware(audience string, authorizationServer string) *jwtmiddleware.JWTMiddleware {
	return jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			// Verify 'aud' claim
			aud := audience
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
			if !checkAud {
				return token, errors.New("invalid audience")
			}
			// Verify 'iss' claim
			iss := authorizationServer
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return token, errors.New("invalid issuer")
			}

			cert, err := getPemCert(token.Header, authorizationServer)
			if err != nil {
				panic(err.Error())
			}

			result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			return result, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
	})
}

func getPemCert(tokenHeader map[string]interface{}, authorizationServer string) (string, error) {
	cert := ""
	resp, err := http.Get(authorizationServer + ".well-known/jwks.json")

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k := range jwks.Keys {
		if tokenHeader["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}

//CreateScopeMiddleware creates a middleware the check the given scope
func CreateScopeMiddleware(scope string, authorizationServer string, audience string) func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		authHeaderParts := strings.Split(r.Header.Get("Authorization"), " ")
		token := authHeaderParts[1]

		hasScope := checkScope(scope, token, authorizationServer, audience)

		if !hasScope {
			message := "Insufficient scope."
			response := Response{message}

			jsonResponse, err := json.Marshal(response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write(jsonResponse)
			return
		}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	}
}

type customClaims struct {
	Scope string `json:"scope"`
	jwtV4.StandardClaims
}

func checkScope(scope string, tokenString string, authorizationServer string, audience string) bool {
	token, err := jwtV4.ParseWithClaims(tokenString, &customClaims{}, func(token *jwtV4.Token) (interface{}, error) {
		cert, err := getPemCert(token.Header, authorizationServer)
		if err != nil {
			return nil, err
		}
		result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
		return result, nil
	}, jwtV4.WithAudience(audience))

	if err != nil {
		log.Println(err)
		return false
	}

	claims, ok := token.Claims.(*customClaims)

	hasScope := false
	if ok && token.Valid {
		result := strings.Split(claims.Scope, " ")
		for i := range result {
			if result[i] == scope {
				hasScope = true
			}
		}
	}

	return hasScope
}

func DefaultExtractUserID(r *http.Request) string {
	user := r.Context().Value("user")
	email := user.(*jwt.Token).Claims.(jwt.MapClaims)["sub"].(string)
	return email
}
