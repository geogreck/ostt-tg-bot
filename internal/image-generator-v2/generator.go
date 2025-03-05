package imagegeneratorv2

import (
	"bytes"
	"context"
	"html/template"
	"image/png"
	"log"
	"net/url"
	"telegram-sticker-bot/internal/models"

	"github.com/chai2010/webp"
	"github.com/chromedp/chromedp"
)

func fontSizeForMessages(messages []models.MessageForSticker) int {
	totalLen := 0
	for _, m := range messages {
		totalLen += len(m.Text)
	}

	switch {
	case totalLen < 50:
		return 50 // Очень мало текста — крупный шрифт
	case totalLen < 100:
		return 40
	case totalLen < 200:
		return 30
	case totalLen < 300:
		return 25
	case totalLen < 500:
		return 20
	case totalLen < 1200:
		return 20
	default:
		return 10 // Очень много текста — мелкий шрифт
	}
}

func renderMessagesHTML(messages []models.MessageForSticker) (string, error) {
	fontSize := fontSizeForMessages(messages)
	data := struct {
		Messages []models.MessageForSticker
		FontSize int
	}{
		Messages: messages,
		FontSize: fontSize,
	}
	tmpl, err := template.New("messages").Parse(strickerHtmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func CreateSticker(messages []models.MessageForSticker) ([]byte, error) {
	htmlContent, err := renderMessagesHTML(messages)
	if err != nil {
		log.Fatal("Error generating HTML:", err)
	}
	encodedHTML := url.PathEscape(htmlContent)

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte

	// Выполняем задачи в браузере
	err = chromedp.Run(ctx,
		chromedp.EmulateViewport(512, 512),
		chromedp.Navigate("data:text/html,"+encodedHTML),
		// chromedp.Sleep(2*time.Second),
		// chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.CaptureScreenshot(&buf),
	)

	if err != nil {
		log.Fatal("Error during chromedp actions:", err)
	}

	img, err := png.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	var webpBuf bytes.Buffer
	if err := webp.Encode(&webpBuf, img, &webp.Options{Lossless: true}); err != nil {
		return nil, err
	}

	return webpBuf.Bytes(), nil
}
