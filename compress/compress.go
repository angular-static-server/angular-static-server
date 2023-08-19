package compress

import (
	"bufio"
	"fmt"
	"ngstaticserver/constants"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/urfave/cli/v2"
)

var Flags = []cli.Flag{
	&cli.Int64Flag{
		EnvVars: []string{"_COMPRESSION_THRESHOLD"},
		Name:    "compression-threshold",
		Value:   constants.DefaultCompressionThreshold,
	},
}

type CompressParams struct {
	Threshold        int64
	WorkingDirectory string
}

func Action(c *cli.Context) error {
	params, err := parseParams(c)
	if err != nil {
		return err
	}

	fmt.Printf(`Parameters:
	Working Directory: %v
	Threshold:         %v

`, params.WorkingDirectory, params.Threshold)

	return compressFilesInDirectory(params)
}

func parseParams(c *cli.Context) (*CompressParams, error) {
	var workingDirectory string
	var err error
	if c.NArg() > 0 {
		workingDirectory, err = filepath.Abs(c.Args().Get(0))
		if err != nil {
			return nil, fmt.Errorf("unable to resolve the absolute path of %v\n%v", c.Args().Get(0), err)
		}
	} else {
		workingDirectory, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve current working directory: %v", err)
		}
	}

	return &CompressParams{
		Threshold:        c.Int64("threshold"),
		WorkingDirectory: workingDirectory,
	}, nil
}

func compressFilesInDirectory(params *CompressParams) error {
	fmt.Printf("starting compression walk in %v:\n", params.WorkingDirectory)
	err := filepath.Walk(params.WorkingDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if info.IsDir() || isCompressedFile(path) {
			return nil
		} else if !isUnicodeFile(path) {
			fmt.Printf("- skipping %v (not a text/unicode file)\n", path)
			return nil
		} else if info.Size() < params.Threshold {
			fmt.Printf("- skipping %v (%v is below threshold %v)\n", path, info.Size(), params.Threshold)
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		err = CompressWithBrotliToFile(content, path+".br")
		if err != nil {
			return err
		} else {
			fmt.Printf("+ creating %v.br\n", path)
		}
		err = CompressWithGzipToFile(content, path+".gz")
		if err != nil {
			return err
		} else {
			fmt.Printf("+ creating %v.gz\n", path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("compression failed: %w", err)
	}

	fmt.Println("\nfinished compression walk")

	return nil
}

func isCompressedFile(path string) bool {
	extension := filepath.Ext(path)
	return extension == ".gz" || extension == ".br"
}

func isUnicodeFile(path string) bool {
	readFile, _ := os.Open(path)
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	fileScanner.Scan()

	return utf8.ValidString(string(fileScanner.Text()))
}
