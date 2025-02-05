package main

import (
	"fmt"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	translate "github.com/OwO-Network/DeepLX/translate"
	"github.com/abadojack/whatlanggo"
	"github.com/gin-gonic/gin"
)

func authMiddleware(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.Token != "" {
			providedTokenInQuery := c.Query("token")
			providedTokenInHeader := c.GetHeader("Authorization")

			// Compatability with the Bearer token format
			if providedTokenInHeader != "" {
				parts := strings.Split(providedTokenInHeader, " ")
				if len(parts) == 2 {
					if parts[0] == "Bearer" || parts[0] == "DeepL-Auth-Key" {
						providedTokenInHeader = parts[1]
					} else {
						providedTokenInHeader = ""
					}
				} else {
					providedTokenInHeader = ""
				}
			}

			if providedTokenInHeader != cfg.Token && providedTokenInQuery != cfg.Token {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "Invalid access token",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

type PayloadFree struct {
	TransText   string `json:"text"`
	SourceLang  string `json:"source_lang"`
	TargetLang  string `json:"target_lang"`
	TagHandling string `json:"tag_handling"`
}

type PayloadAPI struct {
	Text        []string `json:"text"`
	TargetLang  string   `json:"target_lang"`
	SourceLang  string   `json:"source_lang"`
	TagHandling string   `json:"tag_handling"`
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Translator")
	myApp.Settings().SetTheme(theme.DarkTheme()) // Set dark theme

	// Language selection
	sourceLang := widget.NewSelect(getLanguageNames(), nil)
	targetLang := widget.NewSelect(getLanguageNames(), nil)

	targetLang.SetSelected("English")

	// Text entry fields
	sourceEntry := widget.NewMultiLineEntry()
	sourceEntry.MultiLine = true
	sourceEntry.Wrapping = fyne.TextWrapWord

	sourceEntry.OnChanged = func(s string) {
		info := whatlanggo.Detect(s)
		sourceLang.SetSelected(info.Lang.String())
	}
	targetEntry := widget.NewMultiLineEntry()
	targetEntry.MultiLine = true
	targetEntry.Wrapping = fyne.TextWrapWord

	// Translate button
	translateBtn := widget.NewButton("Translate", func() {
		sourceLang := getLanguageCode(sourceLang.Selected)
		targetLang := getLanguageCode(targetLang.Selected)
		if sourceLang == "" || targetLang == "" {
			dialog.NewInformation("Error", "Please Select Language!", myWindow)
			return
		}
		translateText := sourceEntry.Text
		if len(translateText) > 1000 {
			translateText = translateText[:1000]
		}
		result, err := translate.TranslateByDeepLX(sourceLang, targetLang, translateText, "", "", "")
		if err != nil {
			targetEntry.SetText("Error Found")
			return
		}
		if result.Code == http.StatusOK {
			targetEntry.SetText(result.Data)
		} else {
			fmt.Println(result.Message)
			targetEntry.SetText(result.Message)
		}
	})

	clearBtn := widget.NewButton("Clear Text", func() {
		sourceEntry.SetText("")
		targetEntry.SetText("")
	})

	// Layout
	content := container.NewVBox(
		widget.NewLabel("Source Language:"), sourceLang,
		widget.NewLabel("Target Language:"), targetLang,
		widget.NewLabel("Input Text:"), sourceEntry,
		translateBtn,
		widget.NewLabel("Translated Text:"), targetEntry,
		clearBtn,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.ShowAndRun()
}

type Language struct {
	Code string
	Name string
}

var languages = []Language{
	{"AR", "Arabic"},
	{"BG", "Bulgarian"},
	{"CS", "Czech"},
	{"DA", "Danish"},
	{"DE", "German"},
	{"EL", "Greek"},
	{"EN", "English"},
	{"ES", "Spanish"},
	{"ET", "Estonian"},
	{"FI", "Finnish"},
	{"FR", "French"},
	{"HU", "Hungarian"},
	{"ID", "Indonesian"},
	{"IT", "Italian"},
	{"JA", "Japanese"},
	{"KO", "Korean"},
	{"LT", "Lithuanian"},
	{"LV", "Latvian"},
	{"NB", "Norwegian Bokmål"},
	{"NL", "Dutch"},
	{"PL", "Polish"},
	{"PT", "Portuguese"},
	{"RO", "Romanian"},
	{"RU", "Russian"},
	{"SK", "Slovak"},
	{"SL", "Slovenian"},
	{"SV", "Swedish"},
	{"TR", "Turkish"},
	{"UK", "Ukrainian"},
	{"ZH", "Chinese"},
}

func getLanguageNames() []string {
	var names []string
	for _, lang := range languages {
		names = append(names, lang.Name)
	}
	return names
}

func getLanguageCode(name string) string {
	for _, lang := range languages {
		if lang.Name == name {
			return lang.Code
		}
	}
	return ""
}
