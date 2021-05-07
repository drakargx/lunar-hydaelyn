package lunar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	graphQLApi = "https://www.fflogs.com/api/v2/client"
	authURL    = "https://www.fflogs.com/oauth/authorize"
	tokenURL   = "https://www.fflogs.com/oauth/token"

	//Use this query to obtain the ids, start and end times of all the fights in the report.
	//It also grabs the name of the fight, which is usually the name of the boss.
	//We need the start and end time to retrieve information about damage downs and deaths.
	fightsQuery = `
	query($report: String!) {
		reportData {
			report(code: $report) {
				fights {
					id
					name
					startTime
					endTime
				}
			}
		}
	}
	`

	query = `
	query($report: String!, $start: Float!, $end: Float!, $fid: [Int!]!) {
		reportData {
			report(code: $report) {
				...composition
				...rdpsRanking
				...dmgDowns
			}
		}
	}

	fragment dmgDowns on Report {
		events(startTime: $start, endTime: $end, dataType: Debuffs, abilityID: 1002092) {
		  data
		}
	  }
	  
	  fragment composition on Report {
		composition: table(fightIDs: $fid, startTime: $start, endTime: $end, dataType: Summary)
	  }
	  
	  fragment rdpsRanking on Report {
		rankings(fightIDs: $fid, playerMetric: rdps)
	  }
	  
	`
)

var (
	JobTickers = map[string]string{
		"Paladin":    "PLD",
		"Warrior":    "WAR",
		"DarkKnight": "DRK",
		"Gunbreaker": "GNB",

		"Monk":    "MNK",
		"Dragoon": "DRG",
		"Ninja":   "NIN",
		"Samurai": "SAM",

		"WhiteMage":   "WHM",
		"Scholar":     "SCH",
		"Astrologian": "AST",

		"Bard":      "BRD",
		"Machinist": "MCH",
		"Dancer":    "DNC",

		"BlackMage": "BLM",
		"Summoner":  "SMN",
		"RedMage":   "RDM",
		"BlueMage":  "BLU",
	}
)

type CompositionPlayer struct {
	Name string `json:"name"`
	Job  string `json:"type"`
	ID   int    `json:"id"`
}

type DeathEvent struct {
	PlayerName string `json:"name"`
	PlayerID   int    `json:"id"`
}

type DamageDowns struct {
	DebuffMode string `json:"type"`
	TargetID   int    `json:"targetID"`
}

type Fight struct {
	ID        int    `json:"id"`
	BossName  string `json:"name"`
	StartTime int    `json:"startTime"`
	EndTime   int    `json:"endTime"`
}

type RdpsRankingCharacter struct {
	PlayerName  string  `json:"name"`
	Rdps        float64 `json:"amount"`
	FightRank   string  `json:"rank"`
	BestRank    string  `json:"best"`
	RankPercent int     `json:"rankPercent"`
	//Class2 is to detect if the player is a combined healer or tank ranking
	Class2 *string `json:"class_2"`
}

type SpeedRanking struct {
	Rank        string `json:"rank"`
	TotalParses int    `json:"totalParses"`
	RankPercent int    `json:"rankPercent"`
}

type FightsResponse struct {
	QueryData struct {
		ReportData struct {
			Report struct {
				Fights []Fight `json:"fights"`
			} `json:"report"`
		} `json:"reportData"`
	} `json:"data"`
}

type QueryResponse struct {
	Data struct {
		ReportData struct {
			Report struct {
				Composition struct {
					CompositionData struct {
						TotalTime   float64             `json:"totalTime"`
						Party       []CompositionPlayer `json:"composition"`
						DeathEvents []DeathEvent        `json:"deathEvents"`
					} `json:"data"`
				} `json:"composition"`
				Rankings struct {
					RankingData []RankingData `json:"data"`
				} `json:"rankings"`
				DamageDownEvents struct {
					DmgDownInformation []DamageDowns `json:"data"`
				} `json:"events"`
			} `json:"report"`
		} `json:"reportData"`
	} `json:"data"`
}

type RankingData struct {
	Roles struct {
		Tanks struct {
			TankCharacters []RdpsRankingCharacter `json:"characters"`
		} `json:"tanks"`
		Healers struct {
			HealerCharacters []RdpsRankingCharacter `json:"characters"`
		} `json:"healers"`
		DPS struct {
			DPSCharacters []RdpsRankingCharacter `json:"characters"`
		} `json:"dps"`
	} `json:"roles"`
	SpeedRank SpeedRanking `json:"speed"`
}

type DeconstructedQueryResponse struct {
	TotalTime float64
	Deaths    map[string]int
	Players   map[string]RdpsRankingCharacter
	//RdpsRankings []RdpsRankingCharacter
	SpeedRank   SpeedRanking
	DamageDowns map[string]int
	Jobs        map[string]string
}

type FFLogsClient struct {
	client      *http.Client
	AccessToken Token
}

type Token struct {
	AccessToken string `json:"access_token"`
}

func NewFFLogsClient() FFLogsClient {
	hc := http.Client{Timeout: time.Second * 10}

	//Retrieve the access token using Oauth2
	clientID := os.Getenv("LUNAR_HYDAELYN_ID")
	clientSecret := os.Getenv("LUNAR_HYDAELYN_SECRET")

	multiForm := url.Values{}
	multiForm.Add("grant_type", `client_credentials`)

	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(multiForm.Encode()))

	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := hc.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}

	var token Token
	json.Unmarshal(bodyBytes, &token)
	return FFLogsClient{
		client:      &http.Client{Timeout: time.Second * 10},
		AccessToken: token,
	}
}

func (c FFLogsClient) runRequest(query string, variables map[string]interface{}) []byte {
	requestBody := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     query,
		Variables: variables,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
	}

	body := bytes.NewReader(payload)
	req, err := http.NewRequest(http.MethodGet, graphQLApi, body)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprint("Bearer ", c.AccessToken.AccessToken))

	hc := http.Client{Timeout: time.Second * 10}
	resp, err := hc.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return bodyBytes
}

func (c FFLogsClient) GetLastFightInfo(report string) (*Fight, int) {
	requestVariables := map[string]interface{}{
		"report": report,
	}

	var response FightsResponse
	json.Unmarshal(c.runRequest(fightsQuery, requestVariables), &response)

	fights := response.QueryData.ReportData.Report.Fights
	numFights := len(fights)

	highestTime := 0
	var lastFight *Fight
	for i, f := range fights {
		if f.EndTime > highestTime {
			highestTime = f.EndTime
			lastFight = &fights[i]
		}
	}

	if numFights > 0 {
		return lastFight, numFights
	} else {
		return nil, 0
	}
}

func (c FFLogsClient) GrabReportInfo(report string, fight Fight) DeconstructedQueryResponse {
	requestVariables := map[string]interface{}{
		"report": report,
		"start":  fight.StartTime,
		"end":    fight.EndTime,
		"fid":    [1]int{fight.ID},
	}

	var response QueryResponse
	json.Unmarshal(c.runRequest(query, requestVariables), &response)

	var dr DeconstructedQueryResponse
	dr.TotalTime = response.Data.ReportData.Report.Composition.CompositionData.TotalTime
	dr.SpeedRank = response.Data.ReportData.Report.Rankings.RankingData[0].SpeedRank

	dr.Players = make(map[string]RdpsRankingCharacter)
	for _, player := range response.Data.ReportData.Report.Rankings.RankingData[0].Roles.Tanks.TankCharacters {
		if player.Class2 == nil {
			dr.Players[player.PlayerName] = player
		}
	}
	for _, player := range response.Data.ReportData.Report.Rankings.RankingData[0].Roles.Healers.HealerCharacters {
		if player.Class2 == nil {
			dr.Players[player.PlayerName] = player
		}
	}
	for _, player := range response.Data.ReportData.Report.Rankings.RankingData[0].Roles.DPS.DPSCharacters {
		dr.Players[player.PlayerName] = player
	}

	dr.Deaths = make(map[string]int)
	dr.DamageDowns = make(map[string]int)
	for _, s := range dr.Players {
		dr.Deaths[s.PlayerName] = 0
		dr.DamageDowns[s.PlayerName] = 0
	}

	for _, s := range response.Data.ReportData.Report.Composition.CompositionData.DeathEvents {
		dr.Deaths[s.PlayerName]++
	}

	dr.Jobs = make(map[string]string)
	for _, player := range response.Data.ReportData.Report.Composition.CompositionData.Party {
		dr.Jobs[player.Name] = JobTickers[player.Job]
	}

	return dr
}
