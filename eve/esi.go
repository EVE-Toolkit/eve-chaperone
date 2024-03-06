package eve

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dvsekhvalnov/jose2go/base64url"
	"github.com/golang-jwt/jwt/v5"
)

const BaseESIRoute = "https://esi.evetech.net/latest"

type ESIAuthResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

type ESIAuth struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	CharacterName string `json:"name"`
}

type SolarSystemResponse struct {
	SolarSystemId int `json:"solar_system_id"`
	StationId     int `json:"station_id"`
	StructureId   int `json:"structure_id"`
}

type LocationResponse struct {
	Name          string `json:"name"`
	SolarSystemId int    `json:"solar_system_id"`
	StationId     int    `json:"station_id"`
	StructureId   int    `json:"structure_id"`
}

type JWKS struct {
	Keys []jsonWebKey `json:"keys"`
}

type jsonWebKey struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
}

type AccessTokenJWT struct {
	Name string `json:"name"`
	Sub  string `json:"sub"`
	Iss  string `json:"iss"`
}

func CheckAuth(ctx context.Context) bool {
	return ctx.Value("current_character") != nil
}

func GetCurrentCharacter(ctx context.Context) AccessTokenJWT {
	return ctx.Value("current_character").(AccessTokenJWT)
}

func SetCurrentCharacter(ctx context.Context, character AccessTokenJWT) context.Context {
	return context.WithValue(ctx, "current_character", character)
}

func GetAccessToken(ctx context.Context) (context.Context, error) {

	done := make(chan bool)

	var state string
	var challengeBytes string
	var codeChallenge string

	esiAuthResponse := ESIAuthResponse{}

	mux := http.NewServeMux()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		cb, err := generateState()

		if err != nil {
			fmt.Println(err)

			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		st, err := generateState()

		if err != nil {
			fmt.Println(err)

			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		hash := sha256.New()

		hash.Write([]byte(cb))

		cc := base64url.Encode(hash.Sum(nil))

		state = st
		challengeBytes = cb
		codeChallenge = cc

		data := url.Values{}

		data.Add("Response_type", "code")
		data.Add("redirect_uri", "http://localhost:3003/callback")
		data.Add("client_id", os.Getenv("ESI_CLIENT_ID"))
		data.Add("scope", "esi-location.read_location.v1")
		data.Add("code_challenge", codeChallenge)
		data.Add("code_challenge_method", "S256")
		data.Add("state", state)

		http.Redirect(w, r, "https://login.eveonline.com/v2/oauth/authorize?"+data.Encode(), http.StatusTemporaryRedirect) // redirect user to oauth endpoint
	})

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		st := r.URL.Query().Get("state")

		if st != state {
			http.Error(w, "the state is not valid.", http.StatusInternalServerError)

			return
		}

		tokenData := url.Values{}

		tokenData.Add("grant_type", "authorization_code")
		tokenData.Add("code", code)
		tokenData.Add("client_id", os.Getenv("ESI_CLIENT_ID"))
		tokenData.Add("code_verifier", challengeBytes)

		req, err := http.NewRequest(
			"POST",
			"https://login.eveonline.com/v2/oauth/token",
			bytes.NewReader([]byte(tokenData.Encode())),
		)

		if err != nil {
			fmt.Println(err)

			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Host", "login.eveonline.com")

		res, err := http.DefaultClient.Do(req)

		if err != nil {
			fmt.Println(err)

			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		defer res.Body.Close()

		fmt.Println(res.Status)
		fmt.Println(res.StatusCode)

		err = ProcessBody(res.Body, &esiAuthResponse)

		if err != nil {
			fmt.Println(err)

			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		if esiAuthResponse.AccessToken == "" {
			fmt.Println(err)

			http.Error(w, errors.New("there was an error authenticating").Error(), http.StatusInternalServerError)

			return
		}

		character, err := GetCharacter(esiAuthResponse.AccessToken)

		if err != nil {
			fmt.Println(err)

			http.Error(w, errors.New("there was an error authenticating").Error(), http.StatusInternalServerError)

			return
		}

		ctx = context.WithValue(ctx, "access_token_"+character.Name, esiAuthResponse.AccessToken)
		ctx = context.WithValue(ctx, "refresh_token_"+character.Name, esiAuthResponse.RefreshToken)
		ctx = SetCurrentCharacter(ctx, character)

		done <- true

		w.WriteHeader(200)
		w.Write([]byte("You were authenticated."))
	})

	server := &http.Server{Addr: ":3003", Handler: mux}

	go func() {
		_, err := net.DialTimeout("tcp", "localhost:3003", 1*time.Second)

		if err != nil {
			err := server.ListenAndServe()

			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	<-done

	err := server.Close()

	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func GetLocation(ctx context.Context) (LocationResponse, error) {
	if ctx.Value("current_character") == nil {
		return LocationResponse{}, errors.New("no character is logged in right now")
	}

	character := ctx.Value("current_character").(AccessTokenJWT)

	characterId := character.Sub[strings.LastIndex(character.Sub, ":")+1:]

	data := url.Values{}

	data.Add("datasource", "tranquility")
	data.Add("token", ctx.Value("access_token_"+character.Name).(string))

	res, err := http.Get(BaseESIRoute + "/characters/" + characterId + "/location?" + data.Encode())

	if err != nil {
		return LocationResponse{}, err
	}

	if res.StatusCode != 200 {
		bytes, _ := io.ReadAll(res.Body)

		fmt.Println(string(bytes))

		return LocationResponse{}, errors.New(res.Status)
	}

	solarSystemResponse := SolarSystemResponse{}

	err = ProcessBody(res.Body, &solarSystemResponse)

	if err != nil {
		return LocationResponse{}, err
	}

	location, err := http.Get(BaseESIRoute + "/universe/systems/" + strconv.Itoa(solarSystemResponse.SolarSystemId))

	if err != nil {
		return LocationResponse{}, err
	}

	locationResponse := LocationResponse{}

	err = ProcessBody(location.Body, &locationResponse)

	if err != nil {
		return LocationResponse{}, err
	}

	locationResponse.SolarSystemId = solarSystemResponse.SolarSystemId
	locationResponse.StationId = solarSystemResponse.StationId
	locationResponse.StructureId = solarSystemResponse.StructureId

	return locationResponse, nil
}

func GetCharacter(jwtString string) (AccessTokenJWT, error) {
	token, _, err := jwt.NewParser().ParseUnverified(jwtString, jwt.MapClaims{})

	if err != nil {
		return AccessTokenJWT{}, err
	}

	jwks, err := fetchJWKS()

	if err != nil {
		return AccessTokenJWT{}, err
	}

	key, err := extractAndValidateKey(token, jwks)

	if err != nil {
		return AccessTokenJWT{}, err
	}

	claims := jwt.MapClaims{}

	_, err = jwt.ParseWithClaims(jwtString, claims, func(t *jwt.Token) (interface{}, error) {
		return key, nil
	})

	if err != nil {
		return AccessTokenJWT{}, err
	}

	character := AccessTokenJWT{
		Sub:  token.Claims.(jwt.MapClaims)["sub"].(string),
		Name: token.Claims.(jwt.MapClaims)["name"].(string),
		Iss:  token.Claims.(jwt.MapClaims)["iss"].(string),
	}

	return character, nil
}

func RefreshToken(ctx context.Context, character ESIAuth) (context.Context, error) {
	data := url.Values{}

	refreshToken := ctx.Value("refresh_token_" + character.CharacterName)

	if refreshToken == nil {
		return ctx, errors.New("the refresh token does not exist")
	}

	data.Add("grant_type", "refresh_token")
	data.Add("refresh_token", refreshToken.(string))
	data.Add("client_id", os.Getenv("ESI_CLIENT_ID"))

	req, err := http.NewRequest(
		"POST",
		"https://login.eveonline.com/v2/oauth/token",
		strings.NewReader(data.Encode()),
	)

	if err != nil {
		return ctx, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Host", "login.eveonline.com")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return ctx, err
	}

	esiAuthResponse := ESIAuthResponse{}

	err = ProcessBody(res.Body, &esiAuthResponse)

	if err != nil {
		return ctx, err
	}

	fmt.Println("refreshing")

	ctx = context.WithValue(ctx, "access_token_"+character.CharacterName, esiAuthResponse.AccessToken)
	ctx = context.WithValue(ctx, "refresh_token_"+character.CharacterName, esiAuthResponse.RefreshToken)

	return ctx, nil
}

func ProcessBody(body io.ReadCloser, s interface{}) error {
	bytes, err := io.ReadAll(body)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	return nil
}

func generateState() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	return base64url.Encode(bytes), err
}

func fetchJWKS() (*JWKS, error) {
	res, err := http.Get("https://login.eveonline.com/oauth/jwks")

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var jwks JWKS

	err = ProcessBody(res.Body, &jwks)

	if err != nil {
		return nil, err
	}

	return &jwks, nil
}

func extractAndValidateKey(token *jwt.Token, jwks *JWKS) (*rsa.PublicKey, error) {
	kid := token.Header["kid"].(string)

	for _, key := range jwks.Keys {
		if key.Kid == kid && key.Kty == "RSA" && key.Use == "sig" {
			modulus, err := base64url.Decode(key.N)

			if err != nil {
				return nil, err
			}

			exponent, err := base64url.Decode(key.E)

			if err != nil {
				return nil, err
			}

			rsaPublicKey := &rsa.PublicKey{
				N: new(big.Int).SetBytes(modulus),
				E: int(new(big.Int).SetBytes(exponent).Int64()),
			}
			return rsaPublicKey, nil
		}
	}

	return nil, fmt.Errorf("unable to find RSA public key for kid: %s", kid)
}
