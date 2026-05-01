package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"os"
	"github.com/joho/godotenv"

	"github.com/google/generative-ai-go/genai"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/api/option"
)

var statistika = map[string]int{}

type KursResponse struct {
	Rates map[string]float64 `json:"rates"`
}

type VremeResponse struct {
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Name string `json:"name"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

func dohvatiKurs(valuta string) string {
	if valuta == ""{
		valuta="RSD"
	}
	resp, err := http.Get("https://api.exchangerate-api.com/v4/latest/"+valuta)
	if err != nil {
		return "❌ Greška pri dohvatanju kursa."
	}
	defer resp.Body.Close()

	var data KursResponse
	json.NewDecoder(resp.Body).Decode(&data)

	eur := 1.0 / data.Rates["EUR"]
	usd := 1.0 / data.Rates["USD"]

	return fmt.Sprintf("💶 1 EUR = %.2f %s\n💵 1 USD = %.2f %s", eur,valuta, usd,valuta)
}

func dohvatiVreme(grad, apiKey string) string {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric&lang=sr", grad, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return "❌ Greška pri dohvatanju vremena."
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "❌ Grad nije pronađen. Probaj npr: /prognoza Beograd"
	}

	var data VremeResponse
	json.NewDecoder(resp.Body).Decode(&data)

	opis := ""
	if len(data.Weather) > 0 {
		opis = data.Weather[0].Description
	}

	return fmt.Sprintf("🌤 Vreme za %s:\n🌡 Temperatura: %.1f°C (osećaj %.1f°C)\n💧 Vlažnost: %d%%\n💨 Vetar: %.1f m/s\n📋 Opis: %s",
		data.Name, data.Main.Temp, data.Main.FeelsLike, data.Main.Humidity, data.Wind.Speed, opis)
}

func pitajGemini(apiKey, pitanje string) string {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "❌ Greška pri povezivanju sa AI."
	}
	defer client.Close()

model := client.GenerativeModel("gemini-2.0-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(pitanje))
	if err != nil {
    	return "❌ Greška: " + err.Error()
	}

	var rezultat strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		rezultat.WriteString(fmt.Sprintf("%v", part))
	}
	return rezultat.String()
}

---------------------KvizPitanje
func zapocniKviz(chatID int64) string {
	pitanja := make([]KvizPitanje, len(svaPitanja))
	copy(pitanja, svaPitanja)
	rand.Shuffle(len(pitanja), func(i, j int) {
		pitanja[i], pitanja[j] = pitanja[j], pitanja[i]
	})
	if len(pitanja) > 10 {
		pitanja = pitanja[:10]
	}

	kvizovi[chatID] = &KvizStanje{
		Pitanja:  pitanja,
		Trenutno: 0,
		Poeni:    0,
		Aktivno:  true,
	}

	return formatujPitanje(kvizovi[chatID])
}

func formatujPitanje(k *KvizStanje) string {
	p := k.Pitanja[k.Trenutno]
	return fmt.Sprintf("❓ Pitanje %d/%d\n\n%s\n\nA) %s\nB) %s\nC) %s\nD) %s\n\nOdgovori sa A, B, C ili D",
		k.Trenutno+1, len(k.Pitanja),
		p.Pitanje,
		p.Opcije[0], p.Opcije[1], p.Opcije[2], p.Opcije[3])
}

func odgovoriNaKviz(chatID int64, odgovor string) string {
	k, postoji := kvizovi[chatID]
	if !postoji || !k.Aktivno {
		return "Nema aktivnog kviza. Pokreni ga sa /kviz"
	}

	mapa := map[string]int{"A": 0, "B": 1, "C": 2, "D": 3}
	idx, ok := mapa[strings.ToUpper(strings.TrimSpace(odgovor))]
	if !ok {
		return "Odgovori sa A, B, C ili D"
	}

	p := k.Pitanja[k.Trenutno]
	var odg string

	if idx == p.Tacan {
		k.Poeni++
		odg = "✅ Tacno!\n\n"
	} else {
		odg = fmt.Sprintf("❌ Netacno! Tacan odgovor: %s) %s\n\n", []string{"A", "B", "C", "D"}[p.Tacan], p.Opcije[p.Tacan])
	}

	k.Trenutno++

	if k.Trenutno >= len(k.Pitanja) {
		k.Aktivno = false
		poruka := ""
		switch {
		case k.Poeni >= 9:
			poruka = "Sportski genije! 🏆"
		case k.Poeni >= 7:
			poruka = "Odlicno! 🌟"
		case k.Poeni >= 5:
			poruka = "Solidno, moze i bolje! 👍"
		default:
			poruka = "Vežbaj više! 💪"
		}
		return odg + fmt.Sprintf("🏁 Kviz gotov!\n\nRezultat: %d/%d\n%s\n\nZa novi kviz pošalji /kviz", k.Poeni, len(k.Pitanja), poruka)
	}

	return odg + formatujPitanje(k)
}
------------------------------KVIZ


func main() {
	godotenv.Load()
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	//geminiKey     := os.Getenv("GEMINI_KEY")
	weatherKey    := os.Getenv("WEATHER_KEY")

	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Bot pokrenut: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		tekst := update.Message.Text
		statistika[tekst]++

		var odgovor string
		k, kvizAktivan := kvizovi[chatID]
		if kvizAktivan && k.Aktivno && !strings.HasPrefix(tekst, "/") {
			odgovor = odgovoriNaKviz(chatID, tekst)
		} else {
		switch {
		case tekst == "/start":
			odgovor = "Zdravo! 👋 Ja sam tvoj bot!\n\nKomande:\n/vreme — trenutno vreme i datum\n/prognoza [grad] — vremenska prognoza\n/slucajno — nasumičan broj\n/kurs — kurs evra i dolara\n/statistika — statistika korišćenja\n/ai [pitanje] — pitaj AI\n\nPrimeri:\n/prognoza Beograd?"
		case tekst == "/vreme":
			odgovor = "🕐 " + time.Now().Format("02.01.2006. u 15:04:05")
		case tekst == "/slucajno":
			odgovor = fmt.Sprintf("🎲 Tvoj broj je: %d", rand.Intn(100)+1)
		case strings.HasPrefix(tekst, "/kurs "):
			pitanje := strings.TrimPrefix(tekst, "/kurs ")
			odgovor = dohvatiKurs(pitanje)
		case tekst == "/kurs":
			odgovor = "Napiši pitanje posle komande, npr:\n/kurs RSD"
		case tekst == "/statistika":
			odgovor = "📊 Statistika korišćenja:\n"
			for komanda, broj := range statistika {
				odgovor += fmt.Sprintf("%s — %d puta\n", komanda, broj)
			}
		case strings.HasPrefix(tekst, "/prognoza "):
			grad := strings.TrimPrefix(tekst, "/prognoza ")
			odgovor = dohvatiVreme(grad, weatherKey)
		case tekst == "/prognoza":
			odgovor = "Napiši grad posle komande, npr:\n/prognoza Beograd"
		case tekst == "/kviz":
				odgovor = zapocniKviz(chatID)
		// case strings.HasPrefix(tekst, "/ai "):
		// 	pitanje := strings.TrimPrefix(tekst, "/ai ")
			//odgovor = "🤖 " + pitajGemini(geminiKey, pitanje)
		// case tekst == "/ai":
		// 	odgovor = "Napiši pitanje posle komande, npr:\n/ai Kako se pravi pasta?"
		default:
			odgovor = "Ne razumem tu komandu. Probaj /start"
		}
	}
		msg := tgbotapi.NewMessage(chatID, odgovor)
		bot.Send(msg)
	}
}