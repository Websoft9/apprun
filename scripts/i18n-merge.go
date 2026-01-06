// i18n-merge: 合并模板到现有翻译文件（不覆盖）
// Usage: go run scripts/i18n-merge.go -template=./core/locales/template.toml -target=./core/locales/active.en-US.toml
//go:build ignore
// +build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

var (
	templateFile = flag.String("template", "./core/locales/template.toml", "Template file")
	targetFile   = flag.String("target", "", "Target translation file (required)")
	backup       = flag.Bool("backup", true, "Create backup before merge")
)

type Translation map[string]Message
type Message struct {
	Other string `toml:"other"`
}

func main() {
	flag.Parse()

	if *targetFile == "" {
		fmt.Println("❌ Error: -target is required")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("🔄 Merging translations...\n")
	fmt.Printf("   Template: %s\n", *templateFile)
	fmt.Printf("   Target:   %s\n", *targetFile)

	// 读取模板
	template, err := loadTranslations(*templateFile)
	if err != nil {
		fmt.Printf("❌ Failed to load template: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Template loaded: %d keys\n", len(template))

	// 读取目标文件（如果存在）
	var target Translation
	if _, err := os.Stat(*targetFile); err == nil {
		target, err = loadTranslations(*targetFile)
		if err != nil {
			fmt.Printf("❌ Failed to load target: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Target loaded: %d keys\n", len(target))

		// 备份
		if *backup {
			backupPath := fmt.Sprintf("%s.backup.%s", *targetFile, time.Now().Format("20060102-150405"))
			if err := copyFile(*targetFile, backupPath); err != nil {
				fmt.Printf("⚠️  Failed to create backup: %v\n", err)
			} else {
				fmt.Printf("📦 Backup created: %s\n", backupPath)
			}
		}
	} else {
		target = make(Translation)
		fmt.Printf("📝 Target file does not exist, will create new\n")
	}

	// 合并：仅添加新键，不覆盖现有键
	addedCount := 0
	for key, value := range template {
		if _, exists := target[key]; !exists {
			target[key] = value
			addedCount++
		}
	}

	fmt.Printf("➕ Added %d new keys\n", addedCount)

	// 保存
	if err := saveTranslations(*targetFile, target); err != nil {
		fmt.Printf("❌ Failed to save: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Merged successfully: %s\n", *targetFile)
	fmt.Printf("📊 Total keys: %d\n", len(target))
}

func loadTranslations(path string) (Translation, error) {
	var trans Translation
	if _, err := toml.DecodeFile(path, &trans); err != nil {
		return nil, err
	}
	return trans, nil
}

func saveTranslations(path string, trans Translation) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("# Translation file\n")
	f.WriteString(fmt.Sprintf("# Last updated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	encoder := toml.NewEncoder(f)
	return encoder.Encode(trans)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
