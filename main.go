package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const NEWLINE string = "\n"

const EXIT_ERROR_CODE int = 1
const EXIT_SUCCESS_CODE int = 0

var IsNameForVariable = regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z_$0-9]*$`).MatchString

var (
	EnvDirPath string
	CmdToRun   []string
)

type EnvironmentList map[string]string

// возвращает список строк вида "key=value"
func (e EnvironmentList) stringify() []string {
	result := make([]string, len(e))
	i := 0
	for key, value := range e {
		result[i] = fmt.Sprintf("%s=%s", key, value)
		i += 1
	}
	return result
}

// сканирует указанный каталог и возвращает все переменные окружения (имя файл = содержимое), определенные в нем
func ReadDir(dir string) (EnvironmentList, error) {
	env := make(EnvironmentList)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return env, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		filePath := filepath.Join(dir, fileName)

		if !IsNameForVariable(fileName) {
			continue
		}
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("something bad happened when trying to read file %s\n", fileName)
			continue
		}
		if value := strings.TrimRight(string(content), NEWLINE); len(value) != 0 {
			env[fileName] = value
		}
	}
	return env, nil
}

// запускает программу с аргументами (cmd) c переопределнным окружением.
func RunCmd(cmd []string, env EnvironmentList) error {
	if len(cmd) == 0 {
		return fmt.Errorf("command to run is empty")
	}
	command := exec.Command(cmd[0], cmd[1:]...)

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stdin
	command.Env = append(os.Environ(), env.stringify()...)

	if err := command.Run(); err != nil {
		return err
	}
	return nil
}

func init() {
	if len(os.Args) < 3 {
		return
	}
	EnvDirPath = os.Args[1]
	CmdToRun = os.Args[2:]
}

func main() {
	if len(EnvDirPath) == 0 || len(CmdToRun) == 0 {
		fmt.Fprint(os.Stderr, "you should pass the path to dir for environment variables and command to run\n")
		os.Exit(EXIT_ERROR_CODE)
	}
	env, err := ReadDir(EnvDirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(EXIT_ERROR_CODE)
	}
	if err := RunCmd(CmdToRun, env); err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(EXIT_ERROR_CODE)
	}
	os.Exit(EXIT_SUCCESS_CODE)
}
