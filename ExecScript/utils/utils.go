package utils

import (
	"bufio"
	"os/exec"
	"strings"
	"sync"

	"errors"
)

func ExecScript(script string) (string, error) {

	cmd := exec.Command("/bin/sh", "-c", "bash")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	wg := sync.WaitGroup{}
	out := &strings.Builder{}

	go func() {
		wg.Add(1)
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			out.WriteString(scanner.Text())
			out.WriteString("\n")
		}
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
		}
	}()

	if err := cmd.Start(); err != nil {
		return "", err
	}

	if _, err := stdin.Write([]byte(script)); err != nil {
		return "", err
	}
	if err := stdin.Close(); err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return out.String(), errors.New("failed to call cmd.Wait()")
	}

	wg.Wait()

	return out.String(), nil
}
