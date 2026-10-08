package islands

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Yendric/geny/util"
)

const (
	checkFile    = "islands.check.ts"
	checkConfig  = "tsconfig.islands.json"
	rootTSConfig = "tsconfig.json"
)

type checkLine struct {
	usage  Usage
	island string
}

func Check(stateDir string, reg *Registry, usages []Usage) error {
	if len(usages) == 0 {
		return nil
	}
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", stateDir, err)
	}

	source, lines, err := checkSource(stateDir, reg, dedupe(usages))
	if err != nil {
		return err
	}
	checkPath := filepath.Join(stateDir, checkFile)
	if err := writeIfChanged(checkPath, source); err != nil {
		return err
	}

	configPath := filepath.Join(stateDir, checkConfig)
	config, err := checkTSConfig(stateDir)
	if err != nil {
		return err
	}
	if err := writeIfChanged(configPath, config); err != nil {
		return err
	}

	out, runErr := util.ShellCommand("npx --no-install tsc --pretty false -p " + filepath.ToSlash(configPath)).CombinedOutput()
	if runErr == nil {
		return nil
	}

	diagnostics := parseDiagnostics(out)
	if len(diagnostics) == 0 {
		return fmt.Errorf("checking island props with tsc failed, is typescript installed? (npm install -D typescript)\n%s", bytes.TrimSpace(out))
	}

	var errs []error
	for _, d := range diagnostics {
		errs = append(errs, describe(d, filepath.ToSlash(checkPath), lines))
	}
	return errors.Join(errs...)
}

func writeIfChanged(path string, data []byte) error {
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, data) {
		return nil
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

func dedupe(usages []Usage) []Usage {
	seen := map[string]bool{}
	var out []Usage
	for _, u := range usages {
		key := u.Island.Name + "\x00" + u.Island.PropsJSON()
		if u.Line > 0 {
			key += "\x00" + u.Source + "\x00" + strconv.Itoa(u.Line)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, u)
	}
	return out
}

func checkSource(stateDir string, reg *Registry, usages []Usage) ([]byte, map[int]checkLine, error) {
	var b bytes.Buffer
	lines := map[int]checkLine{}
	line := 1
	writeLine := func(s string) {
		b.WriteString(s)
		b.WriteByte('\n')
		line++
	}

	writeLine(`import type { ComponentProps } from "react";`)

	idents := map[string]string{}
	for _, u := range usages {
		name := u.Island.Name
		if _, ok := idents[name]; ok {
			continue
		}
		file, ok := reg.File(name)
		if !ok {
			return nil, nil, fmt.Errorf("unknown island %s", name)
		}
		rel, err := filepath.Rel(stateDir, file)
		if err != nil {
			return nil, nil, err
		}
		spec, err := json.Marshal(filepath.ToSlash(strings.TrimSuffix(rel, filepath.Ext(rel))))
		if err != nil {
			return nil, nil, err
		}
		ident := "Island" + strconv.Itoa(len(idents))
		idents[name] = ident
		lines[line] = checkLine{island: name}
		writeLine(fmt.Sprintf("import type %s from %s;", ident, spec))
	}

	for _, u := range usages {
		lines[line] = checkLine{usage: u, island: u.Island.Name}
		writeLine(fmt.Sprintf("(%s) satisfies ComponentProps<typeof %s>;", u.Island.PropsJSON(), idents[u.Island.Name]))
	}
	return b.Bytes(), lines, nil
}

func checkTSConfig(stateDir string) ([]byte, error) {
	config := map[string]json.RawMessage{
		"files":   json.RawMessage(`["` + checkFile + `"]`),
		"include": json.RawMessage(`[]`),
	}
	if _, err := os.Stat(rootTSConfig); err == nil {
		rel, err := filepath.Rel(stateDir, rootTSConfig)
		if err != nil {
			return nil, err
		}
		extends, err := json.Marshal(filepath.ToSlash(rel))
		if err != nil {
			return nil, err
		}
		config["extends"] = extends
		config["compilerOptions"] = json.RawMessage(`{"noEmit":true}`)
	} else {
		config["compilerOptions"] = json.RawMessage(`{"noEmit":true,"strict":true,"jsx":"react-jsx","module":"esnext","moduleResolution":"bundler","target":"es2022","skipLibCheck":true,"allowJs":true}`)
	}
	return json.MarshalIndent(config, "", "  ")
}

type diagnostic struct {
	file    string
	line    int
	message string
}

var diagnosticPattern = regexp.MustCompile(`^(.+?)\((\d+),\d+\): error TS\d+: (.*)$`)

func parseDiagnostics(out []byte) []diagnostic {
	var diagnostics []diagnostic
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		text := scanner.Text()
		if m := diagnosticPattern.FindStringSubmatch(text); m != nil {
			line, _ := strconv.Atoi(m[2])
			diagnostics = append(diagnostics, diagnostic{file: m[1], line: line, message: m[3]})
			continue
		}
		if n := len(diagnostics); n > 0 && strings.TrimSpace(text) != "" {
			diagnostics[n-1].message += "\n  " + strings.TrimSpace(text)
		}
	}
	return diagnostics
}

func describe(d diagnostic, checkPath string, lines map[int]checkLine) error {
	if filepath.ToSlash(d.file) != checkPath {
		return fmt.Errorf("%s:%d: %s", d.file, d.line, d.message)
	}
	l, ok := lines[d.line]
	switch {
	case !ok:
		return fmt.Errorf("islands: %s", d.message)
	case l.usage.Source == "":
		return fmt.Errorf("island %s: %s", l.island, d.message)
	case l.usage.Line > 0:
		return fmt.Errorf("%s:%d: island %s: %s", l.usage.Source, l.usage.Line, l.island, d.message)
	default:
		return fmt.Errorf("%s: island %s: %s", l.usage.Source, l.island, d.message)
	}
}
