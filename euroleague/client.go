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

type TeamStanding struct {
    Team    string `json:"club_name"`
    Wins    int    `json:"wins"`
    Losses  int    `json:"losses"`
    Position int   `json:"position"`
}

type StandingsResponse struct {
    Standings []TeamStanding `json:"data"`
}

func GetStandings() ([]TeamStanding, error) {
    url := "https://api-live.euroleague.net/v1/standings?seasonCode=E2024"

    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("greška pri pozivu API-ja: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API vratio status: %d", resp.StatusCode)
    }

    var result StandingsResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("greška pri parsiranju: %w", err)
    }

    if len(result.Standings) == 0 {
        return nil, fmt.Errorf("nema podataka za tabelu")
    }

    return result.Standings, nil
}

func FormatStandings(standings []TeamStanding) string {
    msg := "🏀 *Evroliga — Tabela*\n\n"
    msg += "`#   Tim              W  L`\n"

    for _, t := range standings {
        // Top 8 ide u playoff
        marker := ""
        if t.Position <= 8 {
            marker = "🟢"
        } else {
            marker = "🔴"
        }

        msg += fmt.Sprintf("%s `%2d. %-16s %2d %2d`\n",
            marker, t.Position, t.Team, t.Wins, t.Losses)
    }

    msg += "\n🟢 Playoff  🔴 Eliminisan"
    return msg
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