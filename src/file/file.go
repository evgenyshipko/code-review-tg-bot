package file

import (
	"path/filepath"
	"strings"
)

// Список расширений для файлов с кодом
var codeExtensions = map[string]bool{
	".js":   true,
	".html": true,
	".css":  true,
	".ts":   true,
	".tsx":  true,
	".go":   true,
}

// Функция для проверки, является ли файл кодом
func IsCodeFile(fileName string) bool {
	// Проверка по расширению
	ext := strings.ToLower(filepath.Ext(fileName))
	if codeExtensions[ext] {
		return true
	}

	return false
}
