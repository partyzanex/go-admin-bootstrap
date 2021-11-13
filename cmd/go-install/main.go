package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/partyzanex/go-admin-bootstrap/pkg/cmd"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
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
			return errors.New("os.Getwd")
		}

		localBin = filepath.Join(currDir, localBin)
	}

	dirInfo, err := os.Stat(localBin)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			err = os.MkdirAll(localBin, 0755)
			if err != nil {
				return errors.Wrap(err, "os.MkdirAll")
			}
		default:
			return errors.Wrapf(err, "cannot open dir %q", localBin)
		}
	}

	if !dirInfo.IsDir() {
		return errors.Errorf("%q is not a directory", localBin)
	}

	pkg := ctx.Args().First()
	target := ctx.Args().Get(1)
	version := ""

	if parts := strings.Split(pkg, "@"); len(parts) == 2 {
		pkg = parts[0]
		version = parts[1]

		if target == "" {
			target = filepath.Join(localBin, filepath.Base(pkg))
		}
	} else {
		return errors.Errorf("module name should be a module@version format, got %q", pkg)
	}

	if ctx.Bool(skipIfExistsFlagName) {
		targetInfo, err := os.Stat(target + "@" + version)
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrapf(err, "os.Stat(%q)", target+"@"+version)
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
		return errors.Wrap(err, "mkTempDir")
	}

	err = goModInit(ctx, workDir)
	if err != nil {
		return errors.Wrap(err, "goModInit")
	}

	err = goGet(ctx, workDir, pkg)
	if err != nil {
		return errors.Wrap(err, "goGet")
	}

	output, err := goBuild(ctx, workDir, tags, target+"@"+version, pkg)
	err = goGet(ctx, workDir, pkg)
	if err != nil {
		return errors.Wrap(err, "goBuild")
	}

	if verbose {
		fmt.Println(output)
	}

	err = os.Symlink(target+"@"+version, target)
	if err != nil {
		return errors.Wrap(err, "os.Symlink")
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
		return buf.String(), errors.Wrapf(err, "cmd.Execute %q", buf.String())
	}

	return buf.String(), nil
}

func goGet(ctx context.Context, workDir, pkg string) error {
	buf, err := cmd.Execute(ctx, workDir, "go", "get", "-v", "-d", pkg)
	if err != nil {
		return errors.Wrapf(err, "cmd.Execute %q", buf.String())
	}

	return nil
}

func goModInit(ctx context.Context, workDir string) error {
	_, err := cmd.Execute(ctx, workDir, "go", "mod", "init", "fake")
	if err != nil {
		return errors.Wrap(err, "cmd.Execute")
	}

	return nil
}

func mkTempDir() (string, error) {
	tempDir, err := os.MkdirTemp("/tmp", "install*")
	if err != nil {
		return "", errors.Wrap(err, "os.MkdirTemp")
	}

	return tempDir, nil
}

const (
	localBinFlagName     = "local-bin"
	verboseFlagName      = "verbose"
	skipIfExistsFlagName = "skip-if-exists"
	goTagsFlagName       = "go-tags"
)

func localBinFlag() *cli.PathFlag {
	f := new(cli.PathFlag)
	f.Name = localBinFlagName
	f.Aliases = []string{"l", "bin"}
	f.Usage = "to local directory path for binaries"
	f.EnvVars = []string{"LOCAL_BIN", "GO_INSTALL_LOCAL_BIN"}
	f.FilePath = "./bin"
	f.Required = true
	f.Value = "./bin"
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
