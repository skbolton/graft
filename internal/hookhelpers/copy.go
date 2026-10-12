package hookhelpers

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Copy copies each path from the repo's main worktree into the caller's
// current directory, preserving relative paths. Missing source entries are
// skipped quietly; a failed copy is reported on stderr and makes Copy
// return an error without aborting the remaining paths.
func Copy(paths []string, stderr io.Writer) error {
	if len(paths) == 0 {
		return fmt.Errorf("usage: graft hook copy <path>...")
	}
	for _, p := range paths {
		if err := validatePath(p); err != nil {
			return err
		}
	}

	src, err := mainWorktree("")
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if filepath.Clean(src) == filepath.Clean(cwd) {
		fmt.Fprintln(stderr, "already in the main worktree; nothing to copy")
		return nil
	}

	var failed bool
	for _, p := range paths {
		rel := filepath.Clean(p)
		if err := copyOne(src, cwd, rel); err != nil {
			fmt.Fprintf(stderr, "failed to copy %s: %v\n", rel, err)
			failed = true
		}
	}
	if failed {
		return fmt.Errorf("one or more paths failed to copy")
	}
	return nil
}

func validatePath(p string) error {
	if strings.TrimSpace(p) == "" {
		return fmt.Errorf("empty path")
	}
	if filepath.IsAbs(p) {
		return fmt.Errorf("absolute paths are not allowed: %s", p)
	}
	if p == ".." || strings.HasPrefix(p, "../") || strings.Contains(p, "/../") {
		return fmt.Errorf("paths must stay inside the worktree: %s", p)
	}
	return nil
}

func copyOne(srcRoot, dstRoot, rel string) error {
	src := filepath.Join(srcRoot, rel)
	dst := filepath.Join(dstRoot, rel)
	info, err := os.Lstat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	switch {
	case info.IsDir():
		return copyDir(src, dst)
	case info.Mode()&os.ModeSymlink != 0:
		return copySymlink(src, dst)
	default:
		return copyFile(src, dst, info)
	}
}

func copyFile(src, dst string, info os.FileInfo) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func copySymlink(src, dst string) error {
	target, err := os.Readlink(src)
	if err != nil {
		return err
	}
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Symlink(target, dst)
}

func copyDir(src, dst string) error {
	if _, err := os.Lstat(dst); err == nil {
		if err := os.RemoveAll(dst); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := cpDir(src, dst); err == nil {
		return nil
	}
	return copyDirFallback(src, dst)
}

// cpDir shells out to the platform cp with copy-on-write where supported,
// letting the kernel avoid re-copying large artifact trees like node_modules.
var cpDir = cpDirImpl

func cpDirImpl(src, dst string) error {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"-cRP"}
	case "linux":
		args = []string{"--reflink=auto", "-RP"}
	default:
		args = []string{"-RP"}
	}
	args = append(args, src, dst)
	if _, err := exec.Command("cp", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("cp %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

// copyDirFallback copies a directory tree with the stdlib, preserving
// symlinks as symlinks.
func copyDirFallback(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		switch {
		case d.Type()&fs.ModeSymlink != 0:
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(linkTarget, target)
		case d.IsDir():
			mode := fs.FileMode(0o755)
			if info, err := d.Info(); err == nil {
				mode = info.Mode().Perm()
			}
			return os.MkdirAll(target, mode)
		default:
			info, err := d.Info()
			if err != nil {
				return err
			}
			return copyFile(path, target, info)
		}
	})
}
