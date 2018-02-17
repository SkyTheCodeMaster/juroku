package juroku

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
	"text/template"
)

// GenerateCode generates the ComputerCraft code to render the given image
// that must have an underlying color type of color.RGBA and is assumed to have
// ChunkImage already called on it.
func GenerateCode(img image.Image) ([]byte, error) {
	palette := GetPalette(img)
	if len(palette) > 16 {
		return nil, errors.New("juroku: palette must have <= 16 colors")
	}

	colorsCodes := "0123456789abcdef"

	paletteToColor := make(map[color.RGBA]byte)
	for i, col := range palette {
		paletteToColor[col.(color.RGBA)] = colorsCodes[i]
	}

	type rowData struct {
		Text      []byte
		TextColor []byte
		BgColor   []byte
	}

	var rows []rowData

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y += 3 {
		var row rowData
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x += 2 {
			chunk := make([]byte, 0, 6)
			for dy := 0; dy < 3; dy++ {
				for dx := 0; dx < 2; dx++ {
					chunk = append(chunk,
						paletteToColor[img.At(x+dx, y+dy).(color.RGBA)])
				}
			}

			b, textColor, bgColor := chunkToByte(chunk)
			row.Text = append(row.Text, b)
			row.TextColor = append(row.TextColor, textColor)
			row.BgColor = append(row.BgColor, bgColor)
		}
		rows = append(rows, row)
	}

	buf := new(bytes.Buffer)
	err := cc.Execute(buf, struct {
		Rows    []rowData
		Palette color.Palette
		Width   int
		Height  int
	}{
		Rows:    rows,
		Palette: palette,
		Width:   img.Bounds().Dx() / 2,
		Height:  img.Bounds().Dy() / 3,
	})

	return buf.Bytes(), err
}

func chunkToByte(chunk []byte) (b byte, textColor byte, bgColor byte) {
	bgColor = chunk[5]

	var i uint
	for i = 0; i < 6; i++ {
		if chunk[i] != bgColor {
			textColor = chunk[i]
			b |= 1 << i
		} else {
			b |= 0 << i
		}
	}

	if textColor == 0 {
		textColor = '0'
	}

	return
}

var cc = template.Must(template.New("cc").Funcs(template.FuncMap{
	"colorToHex": func(c color.Color) string {
		r, g, b, _ := c.RGBA()
		return fmt.Sprintf("%X%X%X", r>>8, g>>8, b>>8)
	},
	"bToList": func(b []byte) string {
		parts := make([]string, len(b))
		for i, v := range b {
			parts[i] = strconv.Itoa(int(v))
		}

		return strings.Join(parts, ",")
	},
	"bToString": func(b []byte) string {
		return string(b)
	},
}).Parse(`-- This code was automatically generated by
--        _                  __
--       (_)_  ___________  / /____  __
--      / / / / / ___/ __ \/ //_/ / / /
--     / / /_/ / /  / /_/ / ,< / /_/ /
--  __/ /\__,_/_/   \____/_/|_|\__,_/
-- /___/  by 1lann - github.com/1lann/juroku
--
-- Usage:
-- local img = os.loadAPI("image")
-- img.draw(term) or img.draw(monitor)

local function decode(...)
	local result = ""
	local args = {...}
	for _, v in pairs(args) do
		result = result .. string.char(128 + v)
	end
	return result
end

function draw(t)
	local x, y = t.getCursorPos()

	{{range $index, $color := .Palette -}}
	t.setPaletteColor(2^{{$index}}, 0x{{colorToHex $color}})
	{{end}}

	{{range $index, $row := .Rows -}}
	t.setCursorPos(x, y + {{$index}})
	t.blit(decode({{bToList $row.Text}}), "{{bToString $row.TextColor}}", "{{bToString $row.BgColor}}")
	{{end}}
end

function getSize()
	return {{.Width}}, {{.Height}}
end
`))
