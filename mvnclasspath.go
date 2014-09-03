package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func escapePath(p string) string {
	var buf bytes.Buffer
	for _, r := range []rune(strings.ToLower(p)) {
		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') {
			buf.WriteRune(r)
		} else {
			buf.WriteString(fmt.Sprintf("%%%02x", r))
		}
	}
	return buf.String()
}

func main() {
	home := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		fmt.Fprintln(os.Stderr, "$HOME is required")
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cacheDir := filepath.Join(home, ".mvncachepath")
	cacheKey := escapePath(cwd)
	cacheFile := filepath.Join(cacheDir, cacheKey)

	sfi, err := os.Stat("pom.xml")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cfi, err := os.Stat(cacheFile)
	if err == nil {
		if cfi.ModTime().Unix() > sfi.ModTime().Unix() {
			b, err := ioutil.ReadFile(cacheFile)
			if err == nil {
				fmt.Println(string(b))
				return
			}
		}
	} else if _, err = os.Stat(cacheDir); os.IsNotExist(err) {
		err = os.MkdirAll(cacheDir, 0755)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	cmd := exec.Command("mvn", "dependency:build-classpath", "-DincludeScope=test")
	b, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	br := bufio.NewReader(bytes.NewReader(b))
	for {
		b, _, err = br.ReadLine()
		if err != nil {
			break
		}
		if strings.HasSuffix(string(b), " Dependencies classpath:") {
			b, _, err = br.ReadLine()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			classpath := string(b)
			fmt.Println(classpath)
			err = ioutil.WriteFile(cacheFile, b, 0644)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			break
		}
	}
}
