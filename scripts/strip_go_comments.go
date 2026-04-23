package main

import (
	"encoding/json"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type FileResult struct {
	InputPath  string `json:"input_path"`
	OutputPath string `json:"output_path"`
	Status     string `json:"status"`
	InputSize  int64  `json:"input_size"`
	OutputSize int64  `json:"output_size"`
	Error      string `json:"error,omitempty"`
}

type Summary struct {
	TotalFiles  int          `json:"total_files"`
	Success     int          `json:"success"`
	Failed      int          `json:"failed"`
	Skipped     int          `json:"skipped"`
	Files       []FileResult `json:"files"`
}

func stripComments(inputPath, outputPath string) FileResult {
	result := FileResult{InputPath: inputPath, OutputPath: outputPath}

	info, err := os.Stat(inputPath)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("stat error: %v", err)
		return result
	}
	result.InputSize = info.Size()

	srcBytes, err := os.ReadFile(inputPath)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("read error: %v", err)
		return result
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Base(inputPath), srcBytes, parser.AllErrors|parser.SkipObjectResolution)
	if err != nil && f == nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("parse error (fatal): %v", err)
		return result
	}

	var buf strings.Builder
	if err := format.Node(&buf, fset, f); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("format error: %v", err)
		return result
	}

	stripped := buf.String()

	verifyFset := token.NewFileSet()
	_, verifyErr := parser.ParseFile(verifyFset, "", []byte(stripped), parser.AllErrors|parser.SkipObjectResolution)
	if verifyErr != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("re-parse verification failed: %v", verifyErr)
		return result
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("mkdir error: %v", err)
		return result
	}

	if err := os.WriteFile(outputPath, []byte(stripped), 0644); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("write error: %v", err)
		return result
	}

	result.OutputSize = int64(len(stripped))
	result.Status = "success"
	return result
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input_dir> <output_dir> [--dry-run] [--json-report <path>]\n", os.Args[0])
		os.Exit(1)
	}

	inputDir := os.Args[1]
	outputDir := os.Args[2]
	dryRun := false
	jsonReportPath := ""

	for i := 3; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--dry-run":
			dryRun = true
		case "--json-report":
			if i+1 < len(os.Args) {
				jsonReportPath = os.Args[i+1]
				i++
			}
		}
	}

	var goFiles []string
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Walk error: %v\n", err)
		os.Exit(1)
	}

	summary := Summary{TotalFiles: len(goFiles)}

	for _, goFile := range goFiles {
		relPath, _ := filepath.Rel(inputDir, goFile)
		outPath := filepath.Join(outputDir, relPath)

		if dryRun {
			fset := token.NewFileSet()
			srcBytes, err := os.ReadFile(goFile)
			if err != nil {
				summary.Files = append(summary.Files, FileResult{
					InputPath: goFile, Status: "failed",
					Error: fmt.Sprintf("read error: %v", err),
				})
				summary.Failed++
				continue
			}

			info, _ := os.Stat(goFile)
			f, parseErr := parser.ParseFile(fset, filepath.Base(goFile), srcBytes, parser.AllErrors|parser.SkipObjectResolution)
			if parseErr != nil && f == nil {
				summary.Files = append(summary.Files, FileResult{
					InputPath: goFile, Status: "failed",
					InputSize: info.Size(),
					Error:     fmt.Sprintf("parse error: %v", parseErr),
				})
				summary.Failed++
			} else {
				summary.Files = append(summary.Files, FileResult{
					InputPath: goFile, Status: "success",
					InputSize: info.Size(),
				})
				summary.Success++
			}
		} else {
			result := stripComments(goFile, outPath)
			summary.Files = append(summary.Files, result)
			if result.Status == "success" {
				summary.Success++
			} else {
				summary.Failed++
			}
		}
	}

	fmt.Printf("Total: %d | Success: %d | Failed: %d\n", summary.TotalFiles, summary.Success, summary.Failed)

	if summary.Failed > 0 {
		fmt.Println("\nFailed files:")
		for _, f := range summary.Files {
			if f.Status == "failed" {
				fmt.Printf("  %s: %s\n", f.InputPath, f.Error)
			}
		}
	}

	if jsonReportPath != "" {
		reportBytes, _ := json.MarshalIndent(summary, "", "  ")
		os.WriteFile(jsonReportPath, reportBytes, 0644)
		fmt.Printf("\nJSON report saved to: %s\n", jsonReportPath)
	}
}
