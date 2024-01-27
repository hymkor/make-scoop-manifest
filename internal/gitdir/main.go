package gitdir

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func quote(args []string, f func(string) error) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	r, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer r.Close()
	cmd.Start()

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		//println(sc.Text())
		if err := f(sc.Text()); err != nil {
			return err
		}
	}
	return nil
}

func listUpRemoteBranch() ([]string, error) {
	branches := []string{}
	quote([]string{"git", "remote", "show"}, func(line string) error {
		branches = append(branches, strings.TrimSpace(line))
		return nil
	})
	return branches, nil
}

var (
	rxSshUrl   = regexp.MustCompile(`Push +URL: \w+@github.com:([\w\.-]+)/([\w\.-]+)\.git`)
	rxHttpsUrl = regexp.MustCompile(`Push +URL: https://github.com/([\w\.-]+)/([\w\.-]+?)(\.git)?$`)
)

func GetNameAndRepo() (string, string, error) {
	branch, err := listUpRemoteBranch()
	if err != nil {
		return "", "", err
	}
	if len(branch) < 1 {
		return "", "", errors.New("remote branch not found")
	}
	var user, repo string
	quote([]string{"git", "remote", "show", "-n", branch[0]}, func(line string) error {
		if m := rxSshUrl.FindStringSubmatch(line); m != nil {
			user = m[1]
			repo = m[2]
			return io.EOF
		}
		if m := rxHttpsUrl.FindStringSubmatch(line); m != nil {
			user = m[1]
			repo = m[2]
			return io.EOF
		}
		return nil
	})
	return user, repo, nil
}
