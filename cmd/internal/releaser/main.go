package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/davidmdm/x/xerr"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"golang.org/x/mod/semver"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	dry := flag.Bool("dry", false, "dry-run")
	flag.Parse()

	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("failed to open git repo: %w", err)
	}

	iter, err := repo.Tags()
	if err != nil {
		return fmt.Errorf("failed to read tags: %w", err)
	}

	versions := map[string]string{}

	iter.ForEach(func(r *plumbing.Reference) error {
		fmt.Println("detected tag:", r.Name().String())
		release, version := path.Split(r.Name()[len("refs/tags/"):].String())
		if !semver.IsValid(version) {
			return nil
		}
		release = path.Clean(release)
		if semver.Compare(version, versions[release]) > 0 {
			versions[release] = version
		}
		return nil
	})

	releaser := Releaser{
		Versions: versions,
		Repo:     repo,
		DryRun:   *dry,
	}

	entries, err := os.ReadDir("cmd")
	if err != nil {
		return fmt.Errorf("failed to read cmd directory: %w", err)
	}

	var errs []error
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "internal" {
			continue
		}

		if err := releaser.handlePath(entry.Name()); err != nil {
			errs = append(errs, fmt.Errorf("failed to process %s: %w", entry.Name(), err))
		}
	}

	return xerr.MultiErrOrderedFrom("failed to release", errs...)
}

type Releaser struct {
	Versions map[string]string
	Repo     *git.Repository
	DryRun   bool
}

func (releaser Releaser) handlePath(name string) error {
	version := releaser.Versions[name]

	if version != "" {
		tag := path.Join(name, version)

		hash, err := releaser.Repo.ResolveRevision(plumbing.Revision(plumbing.NewTagReferenceName(tag)))
		if err != nil {
			return fmt.Errorf("failed to resolve: %s: %w", tag, err)
		}

		commit, err := releaser.Repo.CommitObject(*hash)
		if err != nil {
			return fmt.Errorf("failed to get commit for tag %s: %w", tag, err)
		}

		h, err := releaser.Repo.Head()
		if err != nil {
			return fmt.Errorf("failed to resolve head: %w", err)
		}

		head, err := releaser.Repo.CommitObject(h.Hash())
		if err != nil {
			return fmt.Errorf("failed to get head commit: %w", err)
		}

		headTree, err := head.Tree()
		if err != nil {
			return fmt.Errorf("failed to get tree for head commit: %w", err)
		}

		commitTree, err := commit.Tree()
		if err != nil {
			return fmt.Errorf("failed to get tree for tag commit: %w", err)
		}

		changes, err := headTree.Diff(commitTree)
		if err != nil {
			return fmt.Errorf("failed to diff trees: %w", err)
		}

		var changed bool
		for _, change := range changes {
			var (
				from = change.From.Name
				to   = change.To.Name
			)
			if strings.HasPrefix(from, "cmd/"+name+"/") || strings.HasPrefix(to, "cmd/"+name+"/") {
				fmt.Printf("detected change from %s to %s\n", from, to)
				changed = true
				break
			}
		}

		if !changed {
			return nil
		}
	} else {
		fmt.Println("No version found for", name)
	}

	outputPath, err := build(filepath.Join("cmd", name))
	if err != nil {
		return fmt.Errorf("failed to build wasm: %w", err)
	}

	v, _ := strconv.Atoi(semver.Major(version))
	tag := fmt.Sprintf("%s/v%d", name, v+1)

	if releaser.DryRun {
		fmt.Println("dry-run: create realease", tag)
		return nil
	}
	if err := release(tag, outputPath); err != nil {
		return fmt.Errorf("failed to release: %w", err)
	}

	return nil
}

func build(path string) (string, error) {
	_, name := filepath.Split(path)

	out := name + ".wasm"

	cmd := exec.Command("go", "build", "-o", out, "./"+path)
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, output)
	}
	return out, nil
}

func release(tag, path string) error {
	out, err := exec.Command("gh", "release", "create", tag, path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}
