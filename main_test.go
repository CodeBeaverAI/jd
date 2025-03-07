package main

import (
    "fmt"
    "os"
    "os/exec"
    "strings"
    "testing"
    jd "github.com/josephburnett/jd/lib"
    serve "github.com/josephburnett/jd/web/serve"
)

const (
    jdFlags = "JD_FLAGS"
)

func TestMain(t *testing.T) {

    if flags := os.Getenv(jdFlags); flags != "" {
    args := strings.Split(flags, " ")
    os.Args = append([]string{os.Args[0]}, args...)
    main()
    }

    testCases := []struct {
    name     string
    files    map[string]string
    args     []string
    exitCode int
    out      string
    outFile  string
    }{{
    name: "no diff",
    files: map[string]string{
    "a.json": `{"foo":"bar"}`,
    "b.json": `{"foo":"bar"}`,
    },
    args:     []string{"a.json", "b.json"},
    out:      "",
    exitCode: 0,
    }, {
    name: "diff",
    files: map[string]string{
    "a.json": `{"foo":"bar"}`,
    "b.json": `{"foo":"baz"}`,
    },
    args: []string{"a.json", "b.json"},
    out: s(
    `@ ["foo"]`,
    `- "bar"`,
    `+ "baz"`,
    ),
    exitCode: 1,
    }, {
    name: "no diff in patch mode",
    files: map[string]string{
    "a.json": `{}`,
    "b.json": `{}`,
    },
    args:     []string{"-f", "patch", "a.json", "b.json"},
    out:      `[]`,
    exitCode: 0,
    }, {
    name: "no diff in merge mode",
    files: map[string]string{
    "a.json": `{}`,
    "b.json": `{}`,
    },
    args:     []string{"-f", "merge", "a.json", "b.json"},
    out:      `{}`,
    exitCode: 0,
    }, {
    name: "diff in patch mode",
    files: map[string]string{
    "a.json": `{"foo":"bar"}`,
    "b.json": `{"foo":"baz"}`,
    },
    args:     []string{"-f", "patch", "a.json", "b.json"},
    out:      `[{"op":"test","path":"/foo","value":"bar"},{"op":"remove","path":"/foo","value":"bar"},{"op":"add","path":"/foo","value":"baz"}]`,
    exitCode: 1,
    }, {
    name: "diff in merge mode",
    files: map[string]string{
    "a.json": `{"foo":"bar"}`,
    "b.json": `{"foo":"baz"}`,
    },
    args:     []string{"-f", "merge", "a.json", "b.json"},
    out:      `{"foo":"baz"}`,
    exitCode: 1,
    }}

    testName := t.Name()
    for _, tc := range testCases {
    t.Run(tc.name, func(t *testing.T) {
    temp, err := os.MkdirTemp("", testName)
    if err != nil {
    t.Fatalf("error creating temp directory: %v", err)
    }
    defer func() {
    os.RemoveAll(temp)
    }()
    files := map[string]struct{}{}
    for filename, content := range tc.files {
    files[filename] = struct{}{}
    err := os.WriteFile(fmt.Sprintf("%v%v%v", temp, os.PathSeparator, filename), []byte(content), 0666)
    if err != nil {
        t.Fatalf("error writing temp file: %v", err)
    }
    }
    args := make([]string, len(tc.args))
    for i, arg := range tc.args {
    if _, isFile := files[arg]; isFile {
        args[i] = fmt.Sprintf("%v%v%v", temp, os.PathSeparator, arg)
    } else {
        args[i] = arg
    }
    }
    cmd := exec.Command(os.Args[0], "-test.run", testName)
    cmd.Env = append(os.Environ(), jdFlags+"="+strings.Join(args, " "))
    out, _ := cmd.CombinedOutput()
    if string(out) != tc.out {
    t.Errorf("wanted out %q. got %q", tc.out, string(out))
    }
    if exitCode := cmd.ProcessState.ExitCode(); exitCode != tc.exitCode {
    t.Errorf("wanted exit code %v. got %v", tc.exitCode, exitCode)
    }
    })
    }
}

func s(s ...string) string {
    return strings.Join(s, "\n") + "\n"
}

func TestParseMetadata(t *testing.T) {
    // Save original flag values to restore later
    origSet := *set
    origMset := *mset
    origSetkeys := *setkeys
    origFormat := *format
    origPrecision := *precision
    origYaml := *yaml
    defer func() {
    *set = origSet
    *mset = origMset
    *setkeys = origSetkeys
    *format = origFormat
    *precision = origPrecision
    *yaml = origYaml
    }()

    // Test valid metadata with merge format and set option.
    *set = true
    *setkeys = "id"
    *format = "merge"
    *precision = 0.0
    metadata, err := parseMetadata()
    if err != nil {
    t.Errorf("Expected no error, got %v", err)
    }
    if len(metadata) < 3 {
    t.Errorf("Expected at least 3 metadata entries, got %d", len(metadata))
    }

    // Test invalid metadata: precision non zero with set option.
    *precision = 0.001
    _, err = parseMetadata()
    if err == nil {
    t.Errorf("Expected error due to precision with set flags, but got none")
    }
}

func TestParseMetadataV2(t *testing.T) {
    // Save original flag values.
    origSet := *set
    origMset := *mset
    origSetkeys := *setkeys
    origFormat := *format
    origPrecision := *precision
    defer func() {
    *set = origSet
    *mset = origMset
    *setkeys = origSetkeys
    *format = origFormat
    *precision = origPrecision
    }()

    // Test valid options with mset option and patch format.
    *mset = true
    *setkeys = "key1,key2"
    *format = "patch"
    *precision = 0.0
    options, err := parseMetadataV2()
    if err != nil {
    t.Errorf("Expected no error, got %v", err)
    }
    if len(options) < 3 {
    t.Errorf("Expected at least 3 options entries, got %d", len(options))
    }

    // Test invalid: precision non zero with mset option.
    *precision = 0.01
    _, err = parseMetadataV2()
    if err == nil {
    t.Errorf("Expected error due to precision with mset flags, but got none")
    }
}

func TestDiffInvalidJSON(t *testing.T) {
    // Document: This test calls diff with invalid JSON to confirm it returns an error.
    invalidJSON := "invalid"
    validJSON := `{"foo": "bar"}`
    _, _, err := diff(invalidJSON, validJSON, []jd.Metadata{})
    if err == nil {
    t.Errorf("Expected error for invalid JSON input, got nil")
    }
}

func TestServeWebNoHandler(t *testing.T) {
    // Document: This test sets serve.Handle to nil so that serveWeb returns the proper error message.
    origHandle := serve.Handle
    defer func() {
    serve.Handle = origHandle
    }()

    serve.Handle = nil
    err := serveWeb("8080")
    if err == nil {
    t.Errorf("Expected error when serve.Handle is nil, got nil")
    }
    expected := "the web UI wasn't include in this build: use `make build` to include it"
    if err.Error() != expected {
    t.Errorf("Expected error message %q, got %q", expected, err.Error())
    }
}