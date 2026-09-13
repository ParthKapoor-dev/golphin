// bench2json parses `go test -bench -benchmem` output from stdin and appends
// one run to the data.json history file read by the benchmark dashboard.
//
//	go test ./... -bench=. -benchmem -count=6 -run='^$' | go run ./.github/bench/bench2json -data bench/data.json
//
// Commit metadata comes from GITHUB_SHA, COMMIT_MESSAGE, COMMIT_AUTHOR and
// COMMIT_TIMESTAMP. When the history file doesn't exist yet, -legacy imports the
// data.js written by github-action-benchmark so no history is lost.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Samples struct {
	Ns     []float64 `json:"ns"`
	Bytes  []float64 `json:"bytes"`
	Allocs []float64 `json:"allocs"`
	Iters  []int64   `json:"iters"`
}

type Bench struct {
	Pkg     string  `json:"pkg"`
	Name    string  `json:"name"`
	Procs   int     `json:"procs"`
	Samples Samples `json:"samples"`
}

type Env struct {
	Goos   string `json:"goos"`
	Goarch string `json:"goarch"`
	CPU    string `json:"cpu"`
	Go     string `json:"go"`
}

type Run struct {
	SHA        string   `json:"sha"`
	Message    string   `json:"message"`
	Author     string   `json:"author"`
	Timestamp  string   `json:"timestamp"`
	Env        Env      `json:"env"`
	Failed     []string `json:"failed,omitempty"`
	Benchmarks []*Bench `json:"benchmarks"`
}

type Data struct {
	Repo    string `json:"repo"`
	Module  string `json:"module"`
	Updated string `json:"updated"`
	Runs    []Run  `json:"runs"`
}

var (
	benchLine = regexp.MustCompile(`^Benchmark(\S+?)(?:-(\d+))?\s+(\d+)\s+(.*)$`)
	failLine  = regexp.MustCompile(`^FAIL\s+(\S+)`)
)

func main() {
	dataPath := flag.String("data", "data.json", "history file to append to")
	legacyPath := flag.String("legacy", "", "github-action-benchmark data.js to import when -data doesn't exist")
	module := flag.String("module", "", "module path, stripped from package names")
	repo := flag.String("repo", "", "repository URL, used for commit links")
	maxRuns := flag.Int("max-runs", 500, "keep at most this many runs")
	flag.Parse()

	d, err := load(*dataPath, *legacyPath, *module)
	if err != nil {
		fatal("%v", err)
	}

	run := parse(bufio.NewScanner(os.Stdin), *module)
	if len(run.Benchmarks) == 0 {
		fatal("no benchmark results found on stdin")
	}
	run.SHA = os.Getenv("GITHUB_SHA")
	run.Message = os.Getenv("COMMIT_MESSAGE")
	run.Author = os.Getenv("COMMIT_AUTHOR")
	run.Timestamp = os.Getenv("COMMIT_TIMESTAMP")
	if run.Timestamp == "" {
		run.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	// Re-running the workflow for the same commit replaces its run.
	kept := d.Runs[:0]
	for _, r := range d.Runs {
		if run.SHA == "" || r.SHA != run.SHA {
			kept = append(kept, r)
		}
	}
	d.Runs = append(kept, run)
	if len(d.Runs) > *maxRuns {
		d.Runs = d.Runs[len(d.Runs)-*maxRuns:]
	}
	if *repo != "" {
		d.Repo = *repo
	}
	d.Module = *module
	d.Updated = time.Now().UTC().Format(time.RFC3339)

	out, err := json.MarshalIndent(d, "", " ")
	if err != nil {
		fatal("encode: %v", err)
	}
	if err := os.WriteFile(*dataPath, out, 0o644); err != nil {
		fatal("write: %v", err)
	}
	fmt.Fprintf(os.Stderr, "bench2json: recorded %d benchmarks for %.7s (%d runs total)\n",
		len(run.Benchmarks), run.SHA, len(d.Runs))
}

// parse reads go test output, echoing it so CI logs stay readable.
func parse(sc *bufio.Scanner, module string) Run {
	run := Run{Env: Env{Go: runtime.Version()}}
	index := map[string]*Bench{}
	pkg := ""

	for sc.Scan() {
		line := sc.Text()
		fmt.Println(line)

		switch {
		case strings.HasPrefix(line, "goos: "):
			run.Env.Goos = strings.TrimPrefix(line, "goos: ")
		case strings.HasPrefix(line, "goarch: "):
			run.Env.Goarch = strings.TrimPrefix(line, "goarch: ")
		case strings.HasPrefix(line, "cpu: "):
			run.Env.CPU = strings.TrimPrefix(line, "cpu: ")
		case strings.HasPrefix(line, "pkg: "):
			pkg = relPkg(strings.TrimPrefix(line, "pkg: "), module)
		}

		if m := failLine.FindStringSubmatch(line); m != nil {
			run.Failed = append(run.Failed, relPkg(m[1], module))
			continue
		}
		m := benchLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := pkg + "." + m[1]
		b := index[key]
		if b == nil {
			procs, _ := strconv.Atoi(m[2])
			b = &Bench{Pkg: pkg, Name: m[1], Procs: procs}
			index[key] = b
			run.Benchmarks = append(run.Benchmarks, b)
		}
		iters, _ := strconv.ParseInt(m[3], 10, 64)
		b.Samples.Iters = append(b.Samples.Iters, iters)

		fields := strings.Fields(m[4])
		for i := 0; i+1 < len(fields); i += 2 {
			v, err := strconv.ParseFloat(fields[i], 64)
			if err != nil {
				continue
			}
			switch fields[i+1] {
			case "ns/op":
				b.Samples.Ns = append(b.Samples.Ns, v)
			case "B/op":
				b.Samples.Bytes = append(b.Samples.Bytes, v)
			case "allocs/op":
				b.Samples.Allocs = append(b.Samples.Allocs, v)
			}
		}
	}
	return run
}

func relPkg(pkg, module string) string {
	if pkg == module {
		return "."
	}
	return strings.TrimPrefix(pkg, module+"/")
}

func load(dataPath, legacyPath, module string) (Data, error) {
	var d Data
	b, err := os.ReadFile(dataPath)
	if err == nil {
		if err := json.Unmarshal(b, &d); err != nil {
			return d, fmt.Errorf("parse %s: %w", dataPath, err)
		}
		return d, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return d, err
	}
	if legacyPath == "" {
		return d, nil
	}
	b, err = os.ReadFile(legacyPath)
	if errors.Is(err, fs.ErrNotExist) {
		return d, nil
	}
	if err != nil {
		return d, err
	}
	d, err = importLegacy(b, module)
	if err != nil {
		return d, fmt.Errorf("import %s: %w", legacyPath, err)
	}
	fmt.Fprintf(os.Stderr, "bench2json: imported %d runs from %s\n", len(d.Runs), legacyPath)
	return d, nil
}

var legacyName = regexp.MustCompile(`^Benchmark(\S+?)(?: \((\S+)\))? - (ns/op|B/op|allocs/op)$`)

// importLegacy converts github-action-benchmark's `window.BENCHMARK_DATA = {...}`.
func importLegacy(b []byte, module string) (Data, error) {
	start := bytes.IndexByte(b, '{')
	if start < 0 {
		return Data{}, errors.New("no JSON object found")
	}
	var old struct {
		RepoURL string `json:"repoUrl"`
		Entries map[string][]struct {
			Commit struct {
				ID        string `json:"id"`
				Message   string `json:"message"`
				Timestamp string `json:"timestamp"`
				Author    struct {
					Name string `json:"name"`
				} `json:"author"`
			} `json:"commit"`
			Benches []struct {
				Name  string  `json:"name"`
				Value float64 `json:"value"`
				Extra string  `json:"extra"`
			} `json:"benches"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(bytes.TrimRight(b[start:], "; \n"), &old); err != nil {
		return Data{}, err
	}

	d := Data{Repo: old.RepoURL, Module: module}
	extra := regexp.MustCompile(`(\d+) times\n(\d+) procs`)
	for _, entries := range old.Entries {
		for _, e := range entries {
			run := Run{
				SHA:       e.Commit.ID,
				Message:   e.Commit.Message,
				Author:    e.Commit.Author.Name,
				Timestamp: e.Commit.Timestamp,
			}
			index := map[string]*Bench{}
			for _, ob := range e.Benches {
				m := legacyName.FindStringSubmatch(ob.Name)
				if m == nil {
					continue
				}
				// Before benchmarks were package-suffixed, internal/storage was the only package.
				pkg := m[2]
				if pkg == "" {
					pkg = module + "/internal/storage"
				}
				pkg = relPkg(pkg, module)
				b := index[pkg+"."+m[1]]
				if b == nil {
					b = &Bench{Pkg: pkg, Name: m[1]}
					index[pkg+"."+m[1]] = b
					run.Benchmarks = append(run.Benchmarks, b)
				}
				x := extra.FindStringSubmatch(ob.Extra)
				switch m[3] {
				case "ns/op":
					b.Samples.Ns = append(b.Samples.Ns, ob.Value)
					if x != nil {
						iters, _ := strconv.ParseInt(x[1], 10, 64)
						b.Procs, _ = strconv.Atoi(x[2])
						b.Samples.Iters = append(b.Samples.Iters, iters)
					}
				case "B/op":
					b.Samples.Bytes = append(b.Samples.Bytes, ob.Value)
				case "allocs/op":
					b.Samples.Allocs = append(b.Samples.Allocs, ob.Value)
				}
			}
			d.Runs = append(d.Runs, run)
		}
	}
	return d, nil
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "bench2json: "+format+"\n", args...)
	os.Exit(1)
}
