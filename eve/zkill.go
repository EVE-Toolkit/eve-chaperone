package eve

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
)

type Ship struct {
	Name struct {
		En string `json:"en"`
	} `json:"name"`
	Id string `json:"id"`
}

type ZKillboardSystemIDResponse struct {
	KillmailID int `json:"killmail_id"`
	ZKB        struct {
		Hash string `json:"hash"`
	} `json:"zkb"`
}

type Killmail struct {
	Attackers []struct {
		CharacterId   int `json:"character_id"`
		AllianceId    int `json:"alliance_id"`
		CorporationId int `json:"corporation_id"`
		ShipTypeId    int `json:"ship_type_id"`
	} `json:"attackers"`
	Victim struct {
		CharacterId   int `json:"character_id"`
		AllianceId    int `json:"alliance_id"`
		CorporationId int `json:"corporation_id"`
		ShipTypeId    int `json:"ship_type_id"`
	} `json:"victim"`
	KillmailTime string `json:"killmail_time"`
}

type FrontendKillmail struct {
	Victim       FrontendKillmailVictim      `json:"victim"`
	Attackers    []FrontendKillmailAttackers `json:"attackers"`
	KillmailId   int64                       `json:"killmailId"`
	KillmailTime string                      `json:"killmail_time"`
}

type FrontendKillmailAttackers struct {
	Character   string `json:"character_id"`
	Alliance    string `json:"alliance_id"`
	Corporation string `json:"corporation_id"`
	ShipType    string `json:"ship_type_id"`
}

type FrontendKillmailVictim struct {
	Character   string `json:"character_id"`
	Alliance    string `json:"alliance_id"`
	Corporation string `json:"corporation_id"`
	ShipType    string `json:"ship_type_id"`
}

type ESIResourceResponse struct {
	Name string `json:"name"`
}

//TODO: use cache and mutate killmail to contain human readable info

func GetSystemKills(systemID int, pageNumber int, cache Cache) ([]FrontendKillmail, error) {
	fmt.Println(systemID)

	requestCount := 0

	var frontendKillmails []FrontendKillmail

	var zKillboardSystemIDResponses []ZKillboardSystemIDResponse

	res, err := http.Get("https://zkillboard.com/api/kills/systemID/" + strconv.Itoa(systemID) + "/")

	if err != nil {
		return frontendKillmails, err
	}

	requestCount += 1

	err = ProcessBody(res.Body, &zKillboardSystemIDResponses)

	if err != nil {
		return frontendKillmails, err
	}

	var kills [][]ZKillboardSystemIDResponse

	chunkSize := 5

	for i := 0; i < len(zKillboardSystemIDResponses); i += chunkSize {
		end := i + chunkSize

		if end > len(zKillboardSystemIDResponses) {
			end = len(zKillboardSystemIDResponses)
		}

		kills = append(kills, zKillboardSystemIDResponses[i:end])
	}

	if len(kills) == 0 {
		return frontendKillmails, errors.New("no killmails found")
	}

	page := kills[pageNumber]

	fmt.Println("page created")

	for _, kill := range page {
		if cache.Killmails[int64(kill.KillmailID)].KillmailTime != "" {
			frontendKillmails = append(frontendKillmails, cache.Killmails[int64(kill.KillmailID)])

			fmt.Println("found kill in cache")

			continue
		}

		killmailRes, err := http.Get(
			fmt.Sprintf(
				BaseESIRoute+"/killmails/%s/%s",
				strconv.Itoa(kill.KillmailID),
				kill.ZKB.Hash,
			),
		)

		if err != nil {
			return frontendKillmails, err
		}

		requestCount += 1

		killmail := Killmail{}

		err = ProcessBody(killmailRes.Body, &killmail)

		if err != nil {
			return frontendKillmails, err
		}

		t, err := time.Parse(time.RFC3339, killmail.KillmailTime)

		if err != nil {
			return frontendKillmails, err
		}

		frontendKillmail := FrontendKillmail{
			KillmailTime: t.Format("2006-01-02 15:04 AM"),
		}

		for _, attacker := range killmail.Attackers {
			var errors []error

			shipName, err := GetShipName(int64(attacker.ShipTypeId))
			errors = append(errors, err)

			characterName, err := GetCharacterName(int64(attacker.CharacterId))
			errors = append(errors, err)

			/*
				corporationName, err := GetCorporationName(int64(attacker.CorporationId))
				errors = append(errors, err)

				allianceName, err := GetAllianceName(int64(attacker.AllianceId))
				errors = append(errors, err)
			*/

			for _, err := range errors {
				requestCount += 1

				if err != nil {
					return frontendKillmails, err
				}
			}

			frontendKillmail.Attackers = append(frontendKillmail.Attackers, FrontendKillmailAttackers{
				ShipType:    shipName,
				Character:   characterName,
				Corporation: "",
				Alliance:    "",
			})
		}

		var errors []error

		shipName, err := GetShipName(int64(killmail.Victim.ShipTypeId))
		errors = append(errors, err)

		characterName, err := GetCharacterName(int64(killmail.Victim.CharacterId))
		errors = append(errors, err)

		/*
			corporationName, err := GetCorporationName(int64(killmail.Victim.CorporationId))
			errors = append(errors, err)

			allianceName, err := GetAllianceName(int64(killmail.Victim.AllianceId))
			errors = append(errors, err)
		*/

		for _, err := range errors {
			requestCount += 1
			if err != nil {
				return frontendKillmails, err
			}
		}

		frontendKillmailVictim := FrontendKillmailVictim{
			ShipType:    shipName,
			Character:   characterName,
			Corporation: "",
			Alliance:    "",
		}

		frontendKillmail.Victim = frontendKillmailVictim
		frontendKillmail.KillmailId = int64(kill.KillmailID)

		frontendKillmails = append(frontendKillmails, frontendKillmail)

		cache.Killmails[int64(kill.KillmailID)] = frontendKillmail
	}

	fmt.Println(requestCount)

	return frontendKillmails, nil
}

func GetCharacterName(characterId int64) (string, error) {
	esiResourceResponse := ESIResourceResponse{}

	res, err := http.Get(BaseESIRoute + "/characters/" + strconv.Itoa(int(characterId)))

	if err != nil {
		return "", err
	}

	err = ProcessBody(res.Body, &esiResourceResponse)

	if err != nil {
		return "", err
	}

	return esiResourceResponse.Name, nil
}

func GetAllianceName(allianceId int64) (string, error) {
	esiResourceResponse := ESIResourceResponse{}

	res, err := http.Get(BaseESIRoute + "/alliances/" + strconv.Itoa(int(allianceId)))

	if err != nil {
		return "", err
	}

	err = ProcessBody(res.Body, &esiResourceResponse)

	if err != nil {
		return "", err
	}

	return esiResourceResponse.Name, nil
}

func GetCorporationName(corporationId int64) (string, error) {
	esiResourceResponse := ESIResourceResponse{}

	res, err := http.Get(BaseESIRoute + "/corporations/" + strconv.Itoa(int(corporationId)))

	if err != nil {
		return "", err
	}

	err = ProcessBody(res.Body, &esiResourceResponse)

	if err != nil {
		return "", err
	}

	return esiResourceResponse.Name, nil
}

func GetShipName(shipId int64) (string, error) {
	data, err := os.ReadFile("./ships.json")

	if err != nil {
		return "", err
	}

	s, _, _, err := jsonparser.Get(data, strconv.Itoa(int(shipId)))

	if err != nil {
		return "Unknown", nil
	}

	ship := Ship{}

	err = json.Unmarshal(s, &ship)

	if err != nil {
		return "Unknown", nil
	}

	return ship.Name.En, nil
}
