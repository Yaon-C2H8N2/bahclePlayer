package utils

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var varRegex = regexp.MustCompile(`\$\{([^}]+)\}`)

func LoadEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open env file: %w", err)
	}
	defer file.Close()

	envMap := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

		envMap[key] = value
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	for key, val := range envMap {
		resolved := resolveEnvVars(val, envMap)
		os.Setenv(key, resolved)
		fmt.Println("Loaded:", key, "=", resolved)
	}

	return nil
}

func resolveEnvVars(value string, env map[string]string) string {
	return varRegex.ReplaceAllStringFunc(value, func(match string) string {
		varName := varRegex.FindStringSubmatch(match)[1]

		if v, ok := env[varName]; ok {
			return resolveEnvVars(v, env)
		}
		return os.Getenv(varName)
	})
}
