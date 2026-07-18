package tutor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func BuildSystemPrompt(skillname string) (prompt string, err error) {
	var file strings.Builder
	BasePath, err := filepath.Abs("prompts/skills")
	if err != nil {
		return "", fmt.Errorf("The basepath isn't right.detail: %s", err)
	}
	targetPath, err := filepath.Abs(filepath.Join("prompts", "skills", skillname+".md"))
	if err != nil {
		return "", fmt.Errorf("The targetpath isn't find,detail: %s", targetPath)
	}
	if !strings.HasPrefix(targetPath, BasePath+string(filepath.Separator)) {
		return "", fmt.Errorf("dangerous path detected: %s", targetPath)
	}

	SystemPath := filepath.Clean(filepath.Join("prompts/", "system.md"))
	MemoryPath := filepath.Clean(filepath.Join("prompts/", "memory.md"))
	SkillPath := filepath.Clean(targetPath)

	Systemfile, err := os.ReadFile(SystemPath)
	if err != nil {
		return "", err
	}

	Skillfile, err := os.ReadFile(SkillPath)
	if err != nil {
		return "", err
	}

	Memoryfile, err := os.ReadFile(MemoryPath)
	if err != nil {
		return "", err
	}

	file.Write(Systemfile)
	file.WriteString("\n")
	file.Write(Skillfile)
	file.WriteString("\n")
	file.Write(Memoryfile)

	return file.String(), err
}
