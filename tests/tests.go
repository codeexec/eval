package tests

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kjk/u"
)

var (
	panicIf = u.PanicIf
)

// TestFile represents a single file in a test case
type TestFile struct {
	fileName string
	src      string
}

// Test represents a single test case
type Test struct {
	files   []*TestFile
	run     []string
	cleanup []string
	stdin   string
	stdout  string
	stderr  string
	// unique n identifying a test, 0...N
	n   int
	raw string
}

func eatPrefix(s string, prefix string) (string, bool) {
	res := strings.TrimPrefix(s, prefix)
	return strings.TrimSpace(res), res != s
}

/*
:run swiftc $file -o main
:run ./main
:cleanup rm ./main
:file main.swift
print("hello from swift")
:stdout
hello from swift
*/
func parseTest(s string) *Test {
	var test Test
	test.raw = s
	lines := strings.Split(s, "\n")
	panicIf(len(lines) < 3, "len(lines)=%d, s:\n%s\n", len(lines), s)

	var file *TestFile
	inStdout := false
	inStderr := false
	inStdin := false
	var currLines []string
	collectLines := func() {
		s := strings.Join(currLines, "\n")
		//fmt.Printf("currLines: %#v\ns:\n%s\n", currLines, s)
		if file != nil {
			file.src = s
			test.files = append(test.files, file)
			file = nil
		} else if inStdout {
			test.stdout = s
			inStdout = false
		} else if inStderr {
			test.stderr = s
			inStderr = false
		} else if inStdin {
			test.stdin = s
			inStdin = false
		} else {
			panicIf(len(currLines) > 0)
		}
		currLines = nil
	}

	for len(lines) > 0 {
		line := lines[0]
		if len(line) == 0 || line[0] != ':' {
			panicIf(file == nil && !inStdout && !inStderr && !inStdin, "line:\n%s\n", line)
			currLines = append(currLines, line)
			lines = lines[1:]
			continue
		}
		collectLines()
		if s, ok := eatPrefix(line, ":run "); ok {
			test.run = append(test.run, s)
		} else if s, ok := eatPrefix(line, ":cleanup "); ok {
			test.cleanup = append(test.cleanup, s)
		} else if s, ok := eatPrefix(line, ":file "); ok {
			panicIf(file != nil)
			file = &TestFile{
				fileName: s,
			}
		} else if _, ok := eatPrefix(line, ":stdout"); ok {
			inStdout = true
		} else if _, ok := eatPrefix(line, ":stderr"); ok {
			inStderr = true
		} else if _, ok := eatPrefix(line, ":stdin"); ok {
			inStdin = true
		}
		lines = lines[1:]
	}
	collectLines()
	panicIf(len(lines) != 0)
	return &test
}

func validateTest(test *Test) {
	panicIf(test.raw == "")
	s := test.stdout + test.stderr
	panicIf(s == "", "test:\n%s\ns:\n%s\n", test.raw, s)
	//panicIf(len(test.run) == 0, "test:\n%s\n", test.raw)
	panicIf(len(test.files) == 0, "test:\n%s\n", test.raw)
	for _, f := range test.files {
		panicIf(f.fileName == "", "test:\n%s\n", test.raw)
		panicIf(f.src == "", "test:\n%s\n", test.raw)
	}
}

// c# is when at least one file is .cs and all other files
// are non-source
func isCSharp(files []*TestFile) bool {
	nMatching := 0
	for _, f := range files {
		name := strings.ToLower(f.fileName)
		if strings.HasSuffix(name, ".cs") {
			nMatching++
			continue
		}
		// those are extensions that
		for _, suff := range []string{".txt", ".md", ".text", ".xml", ".html", ".css"} {
			if !strings.HasSuffix(name, suff) {
				return false
			}
		}
	}
	return nMatching > 0
}

/*
func writeOutTest(test *Test, dir string) {
	u.CreateDirMust(dir)
	path := filepath.Join(dir, "exp_stdout.txt")
	u.WriteFileMust(path, []byte(test.stdout))
	path = filepath.Join(dir, "exp_stderr.txt")
	u.WriteFileMust(path, []byte(test.stderr))

	for _, f := range test.files {
		path = filepath.Join(dir, f.fileName)
		u.WriteFileMust(path, []byte(f.src))
	}

	if isCSharp(test.files) {
		// synthesize a .csproj file so that dotnet run . work
		s := `<Project Sdk="Microsoft.NET.Sdk">
<PropertyGroup>
	<OutputType>Exe</OutputType>
	<TargetFramework>netcoreapp3.1</TargetFramework>
</PropertyGroup>
</Project>
`
		path = filepath.Join(dir, "main.csproj")
		u.WriteFileMust(path, []byte(s))
	}
}

func writeOutTests(tests []*Test) {
	for _, test := range tests {
		testDir := dirForTest(test)
		writeOutTest(test, testDir)
	}
}

func deleteTests(dir string) {
	must(os.RemoveAll(dir))
}
*/

func loadTestsFromFile(path string) []*Test {
	d := u.ReadFileMust(path)
	d = u.NormalizeNewlines(d)
	s := string(d)
	tests := strings.Split(s, "\n---\n")
	panicIf(len(tests) < 2)
	var res []*Test
	for n, testStr := range tests {
		// skip empty string if at the end
		s := strings.TrimSpace(testStr)
		if len(s) == 0 {
			if n == len(tests)-1 {
				continue
			}
			panicIf(true, "empty test at post %d out of %d\n", n, len(tests))
		}
		test := parseTest(testStr)
		validateTest(test)
		test.n = len(res)
		res = append(res, test)
	}
	fmt.Printf("number of tests in %s: %d\n", path, len(res))
	return res
}

// var testFiles = []string{"tests.txt", "tests2.txt", "tests_node_js.js", "tests_cpp.txt"}

var testFiles = []string{"tests.txt"}

// LoadTests loads tests from a file
func LoadTests() []*Test {
	var res []*Test
	for _, fileName := range testFiles {
		path := filepath.Join("tests_data", fileName)
		tests := loadTestsFromFile(path)
		res = append(res, tests...)
	}
	return res
}

/*
func runTests() {
	path := filepath.Join("do", "tests.txt")
	tests := loadTests(path)
	fmt.Printf("%d tests\n", len(tests))

	deleteTests(testsDir)
	writeOutTests(tests)

	var err error
	for _, test := range tests {
		//fmt.Printf("test %d:\n%s\n---\n", test.n, test.files[0].src)
		if true {
			cmd := exec.Command("/bin/bash", "-c", "./"+testScriptName)
			cmd.Dir, err = filepath.Abs(dirForTest(test))
			must(err)
			fmt.Printf("Running test %din dir %s\n", test.n, cmd.Dir)
			u.RunCmdLoggedMust(cmd)
		} else {
			localTestsDir, err := filepath.Abs(testsDir)
			must(err)
			v := fmt.Sprintf("%s:/tests", localTestsDir)
			scriptPath := fmt.Sprintf("/tests/%02d/%s", test.n, testScriptName)
			cmd := exec.Command("docker", "run", "--rm", "-v", v, "-w=/tests", "eval-multi-base:latest", "/bin/bash", "-c", scriptPath)
			u.RunCmdLoggedMust(cmd)
		}
	}
}
*/
