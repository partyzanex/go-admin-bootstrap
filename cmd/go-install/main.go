package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/partyzanex/go-admin-bootstrap/pkg/cmd"
)

var (
	errInvalidModuleFormat = errors.New("module name should be in module@version format")
	errNotDirectory        = errors.New("path is not a directory")
)

func main() {
	app := new(cli.App)
	app.Name = "go-install"
	app.Description = "utility for installing go programs (wrapper for go install)"
	app.Flags = []cli.Flag{
		localBinFlag(),
		verboseFlag(),
		skipIfExistsFlag(),
		goTagsFlag(),
	}
	app.Action = action

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func action(ctx *cli.Context) error {
	localBin := ctx.String(localBinFlagName)

	if strings.HasPrefix(localBin, "./") || !strings.HasPrefix(localBin, "/") {
		currDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("os.Getwd: %w", err)
		}

		localBin = filepath.Join(currDir, localBin)
	}

	dirInfo, err := os.Stat(localBin)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			err = os.MkdirAll(localBin, 0o750) //nolint:mnd // standard directory permission
			if err != nil {
				return fmt.Errorf("os.MkdirAll: %w", err)
			}
		default:
			return fmt.Errorf("cannot open dir %q: %w", localBin, err)
		}
	}

	if dirInfo == nil || !dirInfo.IsDir() {
		return fmt.Errorf("%w: %q", errNotDirectory, localBin)
	}

	pkg := ctx.Args().First()
	target := ctx.Args().Get(1)
	version := ""

	if parts := strings.Split(pkg, "@"); len(parts) == 2 { //nolint:mnd
		pkg = parts[0]
		version = parts[1]

		if target == "" {
			target = filepath.Join(localBin, filepath.Base(pkg))
		}
	} else {
		return fmt.Errorf("%w: got %q", errInvalidModuleFormat, pkg)
	}

	if ctx.Bool(skipIfExistsFlagName) {
		targetInfo, err := os.Stat(target + "@" + version)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("os.Stat(%q): %w", target+"@"+version, err)
		}

		if targetInfo != nil {
			return nil
		}
	}

	return install(ctx.Context, ctx.String(goTagsFlagName), pkg, target, version, ctx.Bool(verboseFlagName))
}

func install(ctx context.Context, tags, pkg, target, version string, verbose bool) error {
	workDir, err := mkTempDir()
	if err != nil {
		return fmt.Errorf("mkTempDir: %w", err)
	}

	err = goModInit(ctx, workDir)
	if err != nil {
		return fmt.Errorf("goModInit: %w", err)
	}

	err = goGet(ctx, workDir, pkg)
	if err != nil {
		return fmt.Errorf("goGet: %w", err)
	}

	output, err := goBuild(ctx, workDir, tags, target+"@"+version, pkg)
	if err != nil {
		return fmt.Errorf("goBuild: %w", err)
	}

	if verbose {
		fmt.Println(output)
	}

	err = os.Symlink(target+"@"+version, target)
	if err != nil {
		return fmt.Errorf("os.Symlink: %w", err)
	}

	return nil
}

func goBuild(ctx context.Context, workDir, tags, target, pkg string) (output string, _ error) {
	args := []string{"build"}

	if tags != "" {
		args = append(args, fmt.Sprintf("-tags='%s'", tags))
	}

	args = append(args, "-v", "-o", target, pkg)

	buf, err := cmd.Execute(ctx, workDir, "go", args...)
	if err != nil {
		return buf.String(), fmt.Errorf("cmd.Execute %q: %w", buf.String(), err)
	}

	return buf.String(), nil
}

func goGet(ctx context.Context, workDir, pkg string) error {
	buf, err := cmd.Execute(ctx, workDir, "go", "get", "-v", "-d", pkg)
	if err != nil {
		return fmt.Errorf("cmd.Execute %q: %w", buf.String(), err)
	}

	return nil
}

func goModInit(ctx context.Context, workDir string) error {
	_, err := cmd.Execute(ctx, workDir, "go", "mod", "init", "fake")
	if err != nil {
		return fmt.Errorf("cmd.Execute: %w", err)
	}

	return nil
}

func mkTempDir() (string, error) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "go-install*")
	if err != nil {
		return "", fmt.Errorf("os.MkdirTemp: %w", err)
	}

	return tempDir, nil
}

const (
	localBinFlagName     = "local-bin"
	verboseFlagName      = "verbose"
	skipIfExistsFlagName = "skip-if-exists"
	goTagsFlagName       = "go-tags"

	defaultLocalBinValue = "./bin"
)

func localBinFlag() *cli.PathFlag {
	f := new(cli.PathFlag)
	f.Name = localBinFlagName
	f.Aliases = []string{"l", "bin"}
	f.Usage = "to local directory path for binaries"
	f.EnvVars = []string{"LOCAL_BIN", "GO_INSTALL_LOCAL_BIN"}
	f.FilePath = defaultLocalBinValue
	f.Required = true
	f.Value = defaultLocalBinValue
	f.HasBeenSet = true

	return f
}

func verboseFlag() *cli.BoolFlag {
	f := new(cli.BoolFlag)
	f.Name = verboseFlagName
	f.Aliases = []string{"v"}

	return f
}

func skipIfExistsFlag() *cli.BoolFlag {
	f := new(cli.BoolFlag)
	f.Name = skipIfExistsFlagName
	f.Aliases = []string{"e", "skip"}
	f.Value = true
	f.HasBeenSet = true

	return f
}

func goTagsFlag() *cli.StringFlag {
	f := new(cli.StringFlag)
	f.Name = goTagsFlagName
	f.Aliases = []string{"t", "tags"}

	return f
}
