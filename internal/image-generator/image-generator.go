package imagegenerator

import (
	"bytes"
	"log"

	"github.com/chai2010/webp"
	"github.com/fogleman/gg"
)

func CreateSticker(text string) ([]byte, error) {
	const width = 512
	const height = 512

	// Создаём новый контекст для рисования
	dc := gg.NewContext(width, height)

	// Задаём белый фон
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// Задаём цвет текста (чёрный)
	dc.SetRGB(0, 0, 0)

	// Загружаем шрифт (убедитесь, что файл arial.ttf доступен)
	if err := dc.LoadFontFace("arial.ttf", 36); err != nil {
		log.Println("Ошибка загрузки шрифта:", err)
		// Если шрифт не удалось загрузить, можно продолжить – текст может быть не отцентрирован
	}

	// Рисуем текст по центру с переносом строк при необходимости
	dc.DrawStringWrapped(text, width/2, height/2, 0.5, 0.5, float64(width)-20, 1.5, gg.AlignCenter)

	// Получаем изображение из контекста
	img := dc.Image()

	// Кодируем изображение в формат WebP
	buf := new(bytes.Buffer)
	if err := webp.Encode(buf, img, &webp.Options{Lossless: true}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
