// ini_editor.go
// INI Color Editor на Go

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

// ANSI-цвета
const (
	reset  = "\033[0m"
	cyan   = "\033[96m"
	green  = "\033[92m"
	yellow = "\033[93m"
	gray   = "\033[90m"
	white  = "\033[37m"
	blue   = "\033[94m"
	black  = "\033[30m"
)

type ColorTheme struct {
	section string
	key     string
	value   string
	comment string
	equals  string
	bracket string
}

var darkTheme = ColorTheme{
	section: cyan,
	key:     green,
	value:   yellow,
	comment: gray,
	equals:  white,
	bracket: white,
}

var lightTheme = ColorTheme{
	section: blue,
	key:     green,
	value:   yellow,
	comment: white,
	equals:  black,
	bracket: black,
}

type INIFile struct {
	sections map[string]map[string]string
	order    []string // порядок секций
}

func NewINIFile() *INIFile {
	return &INIFile{
		sections: make(map[string]map[string]string),
		order:    []string{},
	}
}

func (ini *INIFile) Parse(content string) {
	lines := strings.Split(content, "\n")
	var currentSection string
	ini.sections = make(map[string]map[string]string)
	ini.order = []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
			if _, ok := ini.sections[section]; !ok {
				ini.sections[section] = make(map[string]string)
				ini.order = append(ini.order, section)
			}
			currentSection = section
		} else if strings.Contains(trimmed, "=") && currentSection != "" {
			parts := strings.SplitN(trimmed, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			ini.sections[currentSection][key] = value
		}
	}
	if len(ini.sections) == 0 {
		ini.sections[""] = make(map[string]string)
		ini.order = append(ini.order, "")
	}
}

func (ini *INIFile) GetValue(section, key string) (string, bool) {
	if s, ok := ini.sections[section]; ok {
		val, ok := s[key]
		return val, ok
	}
	return "", false
}

func (ini *INIFile) SetValue(section, key, value string) {
	if _, ok := ini.sections[section]; !ok {
		ini.sections[section] = make(map[string]string)
		ini.order = append(ini.order, section)
	}
	ini.sections[section][key] = value
}

func (ini *INIFile) DeleteKey(section, key string) bool {
	if s, ok := ini.sections[section]; ok {
		if _, exists := s[key]; exists {
			delete(s, key)
			return true
		}
	}
	return false
}

func (ini *INIFile) DeleteSection(section string) bool {
	if _, ok := ini.sections[section]; ok {
		delete(ini.sections, section)
		// Удаляем из порядка
		newOrder := []string{}
		for _, s := range ini.order {
			if s != section {
				newOrder = append(newOrder, s)
			}
		}
		ini.order = newOrder
		return true
	}
	return false
}

func (ini *INIFile) AddSection(section string) bool {
	if _, ok := ini.sections[section]; !ok {
		ini.sections[section] = make(map[string]string)
		ini.order = append(ini.order, section)
		return true
	}
	return false
}

func (ini *INIFile) ToString(useColor bool, theme string) string {
	var t ColorTheme
	if theme == "light" {
		t = lightTheme
	} else {
		t = darkTheme
	}
	var lines []string
	for _, section := range ini.order {
		if section != "" {
			line := t.bracket + "[" + reset + t.section + section + reset + t.bracket + "]" + reset
			lines = append(lines, line)
		}
		keys := ini.sections[section]
		for key, value := range keys {
			line := t.key + key + reset + t.equals + " = " + reset + t.value + value + reset
			lines = append(lines, line)
		}
		if len(keys) > 0 {
			lines = append(lines, "")
		}
	}
	return strings.Join(lines, "\n")
}

func (ini *INIFile) Save(filename string) error {
	return os.WriteFile(filename, []byte(ini.ToString(false, "")), 0644)
}

func main() {
	var file string
	var view bool
	var setVal string
	var getKey string
	var delKey string
	var addSec string
	var remSec string
	var output string
	var noColor bool
	var theme string

	flag.StringVar(&file, "file", "", "INI файл")
	flag.BoolVar(&view, "v", false, "Показать содержимое")
	flag.StringVar(&setVal, "s", "", "Установить ключ (section.key=value)")
	flag.StringVar(&getKey, "g", "", "Получить значение ключа (section.key)")
	flag.StringVar(&delKey, "d", "", "Удалить ключ (section.key)")
	flag.StringVar(&addSec, "a", "", "Добавить секцию")
	flag.StringVar(&remSec, "r", "", "Удалить секцию")
	flag.StringVar(&output, "o", "", "Сохранить в другой файл")
	flag.BoolVar(&noColor, "no-color", false, "Отключить цвета")
	flag.StringVar(&theme, "theme", "dark", "Цветовая тема (dark/light)")
	flag.Parse()

	if file == "" && flag.NArg() > 0 {
		file = flag.Arg(0)
	}

	if file == "" {
		fmt.Print("Введите путь к INI файлу: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			file = strings.TrimSpace(scanner.Text())
		}
		if file == "" {
			fmt.Println("Файл не указан.")
			os.Exit(1)
		}
	}

	ini := NewINIFile()
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Файл %s не найден. Создаём новый.\n", file)
	} else {
		ini.Parse(string(content))
	}

	// Обработка команд
	if setVal != "" {
		parts := strings.SplitN(setVal, "=", 2)
		if len(parts) != 2 {
			fmt.Println("Неверный формат --set. Используйте section.key=value")
			os.Exit(1)
		}
		sectionKey := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		var section, key string
		if dot := strings.Index(sectionKey, "."); dot != -1 {
			section = sectionKey[:dot]
			key = sectionKey[dot+1:]
		} else {
			section = ""
			key = sectionKey
		}
		ini.SetValue(section, key, value)
		fmt.Printf("Установлено: %s.%s = %s\n", section, key, value)
	}

	if getKey != "" {
		var section, key string
		if dot := strings.Index(getKey, "."); dot != -1 {
			section = getKey[:dot]
			key = getKey[dot+1:]
		} else {
			section = ""
			key = getKey
		}
		if val, ok := ini.GetValue(section, key); ok {
			fmt.Println(val)
		} else {
			fmt.Printf("Ключ %s не найден.\n", getKey)
		}
	}

	if delKey != "" {
		var section, key string
		if dot := strings.Index(delKey, "."); dot != -1 {
			section = delKey[:dot]
			key = delKey[dot+1:]
		} else {
			section = ""
			key = delKey
		}
		if ini.DeleteKey(section, key) {
			fmt.Printf("Ключ %s удалён.\n", delKey)
		} else {
			fmt.Printf("Ключ %s не найден.\n", delKey)
		}
	}

	if addSec != "" {
		if ini.AddSection(addSec) {
			fmt.Printf("Секция %s добавлена.\n", addSec)
		} else {
			fmt.Printf("Секция %s уже существует.\n", addSec)
		}
	}

	if remSec != "" {
		if ini.DeleteSection(remSec) {
			fmt.Printf("Секция %s удалена.\n", remSec)
		} else {
			fmt.Printf("Секция %s не найдена.\n", remSec)
		}
	}

	// Вывод
	if view || (setVal == "" && getKey == "" && delKey == "" && addSec == "" && remSec == "") {
		useColor := !noColor
		out := ini.ToString(useColor, theme)
		if output != "" {
			ini.Save(output)
			fmt.Printf("Сохранено в %s\n", output)
		} else {
			fmt.Println(out)
		}
	}

	// Сохранение изменений
	if (setVal != "" || delKey != "" || addSec != "" || remSec != "") && output == "" {
		ini.Save(file)
		fmt.Printf("Изменения сохранены в %s\n", file)
	}
}
