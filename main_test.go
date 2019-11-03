package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/kami-zh/go-capturer"
	"github.com/stretchr/testify/require"
)

type TempDir struct {
	prefix string
	path   string
}

func (td *TempDir) create() error {
	dir, err := ioutil.TempDir(os.TempDir(), td.prefix)
	if err != nil {
		return err
	}
	td.path = dir
	return nil
}

func (td *TempDir) addFile(filename string, content string) error {
	file, err := os.Create(filepath.Join(td.path, filename))
	if err != nil {
		return err
	}
	if _, err := file.WriteString(content); err != nil {
		return err
	}
	return nil
}

func (td *TempDir) addSubDir(subdir string) error {
	if err := os.Mkdir(filepath.Join(td.path, subdir), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (td *TempDir) clean() error {
	return os.RemoveAll(td.path)
}

func newTempDir(prefix string) (*TempDir, error) {
	td := new(TempDir)
	td.prefix = prefix

	err := td.create()
	if err != nil {
		return td, err
	}
	return td, nil
}

func TestReadDir(t *testing.T) {
	td, err := newTempDir(t.Name())
	defer td.clean()
	require.Nil(t, err)

	td.addFile("A", "123")
	td.addFile("B_B", "456\n")
	td.addFile("CcC", "789\n\n")
	td.addFile("Dd12", "1011\n\n12")

	env, err := ReadDir(td.path)
	require.Nil(t, err)
	require.Equal(t, EnvironmentList{"A": "123", "B_B": "456", "CcC": "789", "Dd12": "1011\n\n12"}, env)
}

func TestReadDir_PassNotExistingDir(t *testing.T) {
	env, err := ReadDir("/not/existing/path")
	require.EqualError(t, err, "open /not/existing/path: no such file or directory")
	require.Equal(t, EnvironmentList{}, env)
}

func TestReadDir_DirIsEmpty(t *testing.T) {
	td, err := newTempDir(t.Name())
	defer td.clean()
	require.Nil(t, err)

	env, err := ReadDir(td.path)
	require.Nil(t, err)
	require.Equal(t, EnvironmentList{}, env)
}

func TestReadDir_WorkingWithEmptyFile(t *testing.T) {
	td, err := newTempDir(t.Name())
	defer td.clean()
	require.Nil(t, err)

	td.addFile("A", "")
	td.addFile("B_B", "\n\n")

	env, err := ReadDir(td.path)
	require.Nil(t, err)
	require.Equal(t, EnvironmentList{}, env)
}

func TestReadDir_IgnoreNewLines(t *testing.T) {
	td, err := newTempDir(t.Name())
	defer td.clean()
	require.Nil(t, err)

	td.addFile("A", "abc\n")
	td.addFile("B_B", "def\n\n")

	env, err := ReadDir(td.path)
	require.Nil(t, err)
	require.Equal(t, EnvironmentList{"A": "abc", "B_B": "def"}, env)
}

func TestReadDir_DirHasSubDir(t *testing.T) {
	td, err := newTempDir(t.Name())
	defer td.clean()
	require.Nil(t, err)

	td.addSubDir("A")
	td.addFile("B", "123")

	env, err := ReadDir(td.path)
	require.Nil(t, err)
	require.Equal(t, EnvironmentList{"B": "123"}, env)
}

func TestReadDir_FileNameIsInvalid(t *testing.T) {
	td, err := newTempDir(t.Name())
	defer td.clean()
	require.Nil(t, err)

	td.addFile("B.txt", "123")
	td.addFile("B-?*A", "456")
	td.addFile("12BC", "789")

	env, err := ReadDir(td.path)
	require.Nil(t, err)
	require.Equal(t, EnvironmentList{}, env)
}

func TestRunCmd(t *testing.T) {
	stdout := capturer.CaptureStdout(func() {
		err := RunCmd([]string{"man", "man"}, EnvironmentList{"A": "322"})
		require.Nil(t, err)
	})
	require.Contains(t, stdout, "man - format and display the on-line manual pages")
}

func TestRunCmd_PassInvalidCommand(t *testing.T) {
	err := RunCmd([]string{"unknown_command", "-flag 123"}, EnvironmentList{"A": "322"})
	require.EqualError(t, err, "exec: \"unknown_command\": executable file not found in $PATH")
}

func TestRunCmd_PassInvalidFlags(t *testing.T) {
	err := RunCmd([]string{"env", "-T"}, EnvironmentList{"A": "322"})
	require.EqualError(t, err, "exit status 1")
}

func TestRunCmd_CheckUsingPassedEnv(t *testing.T) {
	stdout := capturer.CaptureStdout(func() {
		err := RunCmd([]string{"env"}, EnvironmentList{"A": "322"})
		require.Nil(t, err)
	})
	require.Contains(t, stdout, "A=322")
}

func TestEnvironmentList(t *testing.T) {
	env := EnvironmentList{"A": "123", "B": "456"}
	require.Equal(t, env.stringify(), []string{"A=123", "B=456"})
}

func TestEnvironmentList_Empty(t *testing.T) {
	env := EnvironmentList{}
	require.Equal(t, env.stringify(), []string{})
}
