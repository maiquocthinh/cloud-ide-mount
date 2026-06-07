package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	err := Init(Options{Level: LevelInfo, Path: logPath})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	// Verify file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Fatal("log file was not created")
	}
}

func TestLevelFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "filter.log")

	err := Init(Options{Level: LevelInfo, Path: logPath})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	Debug("this should be filtered out")
	Info("this should appear")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	content := string(data)

	if strings.Contains(content, "filtered out") {
		t.Error("DEBUG message should not appear at INFO level")
	}
	if !strings.Contains(content, "should appear") {
		t.Error("INFO message should appear in log file")
	}
}

func TestLevelLabels(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "levels.log")

	err := Init(Options{Level: LevelDebug, Path: logPath})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	Info("info message")
	Warn("warn message")
	Error("error message")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "INFO") {
		t.Error("log should contain INFO label")
	}
	if !strings.Contains(content, "WARN") {
		t.Error("log should contain WARN label")
	}
	if !strings.Contains(content, "ERROR") {
		t.Error("log should contain ERROR label")
	}
}

func TestSetLevel(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "setlevel.log")

	err := Init(Options{Level: LevelInfo, Path: logPath})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	SetLevel(LevelError)
	Info("should not appear after level change")
	Error("should appear after level change")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	content := string(data)

	if strings.Contains(content, "should not appear") {
		t.Error("INFO message should not appear after SetLevel(Error)")
	}
	if !strings.Contains(content, "should appear after level change") {
		t.Error("ERROR message should appear after SetLevel(Error)")
	}
}

func TestStructuredFieldsInFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "fields.log")

	err := Init(Options{Level: LevelInfo, Path: logPath})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	Info("test with fields", "codespace", "my-cs", "port", 2222, "pid", 12345)

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "codespace=my-cs") {
		t.Error("log should contain codespace field")
	}
	if !strings.Contains(content, "port=2222") {
		t.Error("log should contain port field")
	}
	if !strings.Contains(content, "pid=12345") {
		t.Error("log should contain pid field")
	}
}

func TestTimestampPresence(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "ts.log")

	err := Init(Options{Level: LevelInfo, Path: logPath})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	Info("check timestamp")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	content := string(data)

	// Verify first character is a digit (start of ISO timestamp)
	if len(content) > 0 && content[0] < '0' || content[0] > '9' {
		t.Errorf("log should start with a timestamp (digit), got: %q", content[:20])
	}
}

func TestInitCreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "sub", "dir", "nested.log")

	err := Init(Options{Level: LevelInfo, Path: logPath})
	if err != nil {
		t.Fatalf("Init with nested path failed: %v", err)
	}
	defer Close()

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Fatal("log file was not created in nested directory")
	}
}

func TestCloseIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "close.log")

	err := Init(Options{Level: LevelInfo, Path: logPath})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Close multiple times should not panic
	Close()
	Close()
}

func TestNoInitDoesNotPanic(t *testing.T) {
	// Calling log functions without Init should not panic
	// Std is nil at this point
	Info("no panic")
	Warn("no panic")
	Error("no panic")
	Debug("no panic")
}
