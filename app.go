package main

import (
	"context"
	"encoding/json"
	"errors"
	"eve-chaperone/eve"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/pkg/browser"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx   context.Context
	Cache eve.Cache
}

type Location struct {
}

func NewApp() *App {
	return &App{
		Cache: eve.NewCache(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) CheckAuth() bool {
	return eve.CheckAuth(a.ctx)
}

func (a *App) OpenAuth() {
	browser.OpenURL("http://localhost:3003/login")

	ctx, err := eve.GetAccessToken(a.ctx)

	if err != nil {
		log.Fatal(err)
	}

	a.ctx = ctx

	initialCharacter := ctx.Value("current_character").(eve.AccessTokenJWT)

	charAuth := eve.ESIAuth{
		AccessToken:   ctx.Value("access_token_" + initialCharacter.Name).(string),
		RefreshToken:  ctx.Value("refresh_token_" + initialCharacter.Name).(string),
		CharacterName: initialCharacter.Name,
	}

	writeCharacters(charAuth)
}

func (a *App) GetZkill(systemId int64) ([]eve.FrontendKillmail, error) {
	killmails, err := eve.GetSystemKills(int(systemId), 0, a.Cache)

	if err != nil {
		return nil, err
	}

	return killmails, nil
}

func (a *App) GetLocation() (eve.LocationResponse, error) {
	locationResponse, err := eve.GetLocation(a.ctx)

	if err != nil {
		if err.Error() == "403 forbidden" {
			if a.ctx.Value("current_character") == nil {
				return eve.LocationResponse{}, errors.New("there are no characters logged in right now")
			}

			character := a.ctx.Value("current_character").(eve.AccessTokenJWT).Name

			ctx, err := eve.RefreshToken(a.ctx, eve.ESIAuth{
				CharacterName: character,
				AccessToken:   a.ctx.Value("access_token_" + character).(string),
				RefreshToken:  a.ctx.Value("refresh_token_" + character).(string),
			})

			if err != nil {
				return eve.LocationResponse{}, err
			}

			a.ctx = ctx
		}

		return eve.LocationResponse{}, err
	}

	return locationResponse, nil
}

func (a *App) SwitchCurrentCharacter(characterName string) error {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	chaperonePath := homeDir + "/.eve-chaperone"

	err = os.MkdirAll(chaperonePath, os.ModePerm)

	if err != nil {
		return err
	}

	var esiAuths []eve.ESIAuth

	currentAuths, err := os.ReadFile(chaperonePath + "./config.json")

	if err != nil {
		return err
	}

	err = json.Unmarshal(currentAuths, &esiAuths)

	if err != nil {
		return err
	}

	var esiAuth eve.ESIAuth

	for _, auth := range esiAuths {
		if auth.CharacterName == characterName {
			esiAuth = auth
		}
	}

	ctx := a.ctx

	fmt.Println(ctx.Value("access_token_" + characterName))

	ctx, err = eve.RefreshToken(ctx, esiAuth)

	if err != nil {
		return err
	}

	esiAuth.AccessToken = ctx.Value("access_token_" + characterName).(string)
	esiAuth.RefreshToken = ctx.Value("refresh_token_" + characterName).(string)

	fmt.Println("refreshed")

	writeCharacters(esiAuth)

	a.ctx = ctx

	character, err := eve.GetCharacter(ctx.Value("access_token_" + characterName).(string))

	if err != nil {
		return err
	}

	contex := eve.SetCurrentCharacter(ctx, character)

	a.ctx = contex

	return nil
}

func (a *App) GetRegisteredCharacters() ([]eve.ESIAuth, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(err)
	}

	chaperonePath := homeDir + "/.eve-chaperone"

	err = os.MkdirAll(chaperonePath, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}

	var esiAuths []eve.ESIAuth

	currentAuths, err := os.ReadFile(chaperonePath + "./config.json")

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(currentAuths, &esiAuths)

	if err != nil {
		log.Fatal(err)
	}

	for _, auth := range esiAuths {
		ctx := a.ctx

		ctx = context.WithValue(ctx, "access_token_"+auth.CharacterName, auth.AccessToken)
		ctx = context.WithValue(ctx, "refresh_token_"+auth.CharacterName, auth.RefreshToken)

		a.ctx = ctx
	}

	return esiAuths, nil
}

func (a *App) LogOut() {
	resetCharacters()

	a.ctx = context.WithValue(a.ctx, "current_character", nil)

	runtime.WindowReload(a.ctx)
}

func (a *App) OnDomReady(ctx context.Context) {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, os.Kill)

	var esiAuths []eve.ESIAuth

	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(err)
	}

	chaperonePath := homeDir + "/.eve-chaperone"

	err = os.MkdirAll(chaperonePath, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}

	currentAuths, err := os.ReadFile(chaperonePath + "./config.json")

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(currentAuths, &esiAuths)

	if err != nil {
		log.Fatal(err)
	}

	for _, auth := range esiAuths {
		ctx := a.ctx

		fmt.Println(auth.CharacterName)

		ctx = context.WithValue(ctx, "access_token_"+auth.CharacterName, auth.AccessToken)
		ctx = context.WithValue(ctx, "refresh_token_"+auth.CharacterName, auth.RefreshToken)

		ctx, err = eve.RefreshToken(ctx, auth)

		if err != nil {
			log.Fatal(err)
		}

		auth.AccessToken = ctx.Value("access_token_" + auth.CharacterName).(string)
		auth.RefreshToken = ctx.Value("refresh_token_" + auth.CharacterName).(string)

		writeCharacters(auth)

		a.ctx = ctx
	}

	go func() {
		ticker := time.NewTicker(15 * time.Minute)

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if !eve.CheckAuth(a.ctx) {
					continue
				}

				ctx = a.ctx

				currentCharacter := ctx.Value("current_character").(eve.AccessTokenJWT)

				ctx, err := eve.RefreshToken(ctx, eve.ESIAuth{
					CharacterName: currentCharacter.Name,
				})

				if err != nil {
					log.Fatal(err)
				}

				a.ctx = ctx

				charAuth := eve.ESIAuth{
					AccessToken:   ctx.Value("access_token_" + currentCharacter.Name).(string),
					RefreshToken:  ctx.Value("refresh_token_" + currentCharacter.Name).(string),
					CharacterName: currentCharacter.Name,
				}

				writeCharacters(charAuth)
			}
		}
	}()
}

func resetCharacters() {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(err)
	}

	chaperonePath := homeDir + "/.eve-chaperone"

	err = os.MkdirAll(chaperonePath, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(chaperonePath+"./config.json", []byte("[]"), fs.ModePerm)

	if err != nil {
		log.Fatal(err)
	}
}

func writeCharacters(charAuth eve.ESIAuth) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(err)
	}

	chaperonePath := homeDir + "/.eve-chaperone"

	err = os.MkdirAll(chaperonePath, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}

	var esiAuths []eve.ESIAuth

	currentAuths, err := os.ReadFile(chaperonePath + "./config.json")

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(currentAuths, &esiAuths)

	if err != nil {
		log.Fatal(err)
	}

	charFound := false

	for i, esiAuth := range esiAuths {
		if esiAuth.CharacterName == charAuth.CharacterName {
			charFound = true

			fmt.Println(esiAuth.AccessToken)
			fmt.Println(charAuth.AccessToken)

			esiAuths[i] = charAuth
		}
	}

	if !charFound {
		esiAuths = append(esiAuths, charAuth)
	}

	bytes, err := json.MarshalIndent(esiAuths, "", "  ")

	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(chaperonePath+"./config.json", bytes, fs.ModePerm)

	if err != nil {
		log.Fatal(err)
	}
}
