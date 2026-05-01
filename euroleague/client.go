package euroleague

import (
    "encoding/json"
    "fmt"
    "net/http"
)

const baseURL = "https://live.euroleague.net/api"

type Game struct {
    HomeTeam  string `json:"hometeam"`
    AwayTeam  string `json:"awayteam"`
    HomeScore int    `json:"homescore"`
    AwayScore int    `json:"awayscore"`
    GameDay   int    `json:"gameday"`
    Date      string `json:"gamedate"`
    Home      bool   `json:"hometeamwin"`
}

type APIResponse struct {
    Games []Game `json:"data"`
}

func GetRoundResults(round int) ([]Game, error) {
    url := fmt.Sprintf("%s/Results?seasonCode=E2024&gameNumber=%d", baseURL, round)

    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("greška pri pozivu API-ja: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API vratio status: %d", resp.StatusCode)
    }

    var result APIResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("greška pri parsiranju: %w", err)
    }

    if len(result.Games) == 0 {
        return nil, fmt.Errorf("nema podataka za kolo %d", round)
    }

    return result.Games, nil
}

func FormatResults(round int, games []Game) string {
    msg := fmt.Sprintf("🏀 *Evroliga — Kolo %d*\n\n", round)

    for _, g := range games {
        winner := ""
        if g.HomeScore > g.AwayScore {
            winner = "🏠"
        } else {
            winner = "✈️"
        }
        msg += fmt.Sprintf("%s `%s %d - %d %s`\n",
            winner, g.HomeTeam, g.HomeScore, g.AwayScore, g.AwayTeam)
    }

    return msg
}