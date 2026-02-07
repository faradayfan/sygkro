package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faradayfan/sygkro/internal/config"
)

// initGitRepo creates a git repo in dir with an initial commit.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
}

func commitAll(t *testing.T, dir, msg string) string {
	t.Helper()
	run(t, dir, "git", "add", "-A")
	run(t, dir, "git", "commit", "-m", msg, "--allow-empty")
	out := run(t, dir, "git", "rev-parse", "HEAD")
	return strings.TrimSpace(out)
}

func run(t *testing.T, dir, cmd string, args ...string) string {
	t.Helper()
	c := exec.Command(cmd, args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("command %q %v in %s failed: %v\n%s", cmd, args, dir, err, out)
	}
	return string(out)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// buildTemplateRepo creates a git repo that acts as a sygkro template with two commits.
//
// v1 (initial):
//   - sygkro.template.yaml
//   - {{ .slug }}/README.md          → "# {{ .name }}\nA project.\n"
//   - {{ .slug }}/config.yaml        → "app: {{ .name }}\nport: 8080\nlog_level: info\ndebug: false\n"
//   - {{ .slug }}/src/main.go        → "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"
//   - {{ .slug }}/docs/guide.md      → "# Guide\nSome docs.\n"
//
// v2 (update):
//   - config.yaml line "port" changed to 9090, added "timeout: 30s"
//   - src/main.go added an import line
//   - added {{ .slug }}/Makefile (new file)
//   - removed {{ .slug }}/docs/guide.md (deleted file)
//   - README.md unchanged
func buildTemplateRepo(t *testing.T) (repoDir, v1sha, v2sha string) {
	t.Helper()
	repoDir = t.TempDir()
	initGitRepo(t, repoDir)

	slugDir := filepath.Join(repoDir, "{{ .slug }}")

	// Template config
	cfg := config.TemplateConfig{
		Name:        "test-template",
		Description: "Integration test template",
		Templating: config.TemplatingConfig{
			Inputs: map[string]string{
				"name": "My App",
				"slug": "my-app",
			},
		},
	}
	if err := cfg.Write(filepath.Join(repoDir, config.TemplateConfigFileName)); err != nil {
		t.Fatal(err)
	}

	// v1 files
	writeFile(t, filepath.Join(slugDir, "README.md"),
		"# {{ .name }}\nA project.\n")
	writeFile(t, filepath.Join(slugDir, "config.yaml"),
		"app: {{ .name }}\nport: 8080\nhost: localhost\nworkers: 4\nlog_level: info\ndebug: false\n")
	writeFile(t, filepath.Join(slugDir, "src", "main.go"),
		"package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n")
	writeFile(t, filepath.Join(slugDir, "docs", "guide.md"),
		"# Guide\nSome docs.\n")

	v1sha = commitAll(t, repoDir, "v1: initial template")

	// v2 changes
	writeFile(t, filepath.Join(slugDir, "config.yaml"),
		"app: {{ .name }}\nport: 9090\nhost: localhost\nworkers: 4\nlog_level: info\ndebug: false\ntimeout: 30s\n")
	writeFile(t, filepath.Join(slugDir, "src", "main.go"),
		"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n")
	writeFile(t, filepath.Join(slugDir, "Makefile"),
		"build:\n\tgo build -o bin/{{ .name }} .\n")
	os.RemoveAll(filepath.Join(slugDir, "docs"))

	v2sha = commitAll(t, repoDir, "v2: update config, add makefile, remove docs")

	return repoDir, v1sha, v2sha
}

// TestSyncIntegration_FullFlow exercises the complete sync pipeline:
//
//  1. Build a local template repo with two commits (v1, v2)
//  2. Render the v1 template as the "project" (simulating project create)
//  3. Add user customizations to the project
//  4. Run the full sync: render old (v1), render new (v2), 3-way merge, apply
//  5. Verify every merge scenario
func TestSyncIntegration_FullFlow(t *testing.T) {
	templateRepo, v1sha, _ := buildTemplateRepo(t)

	inputs := map[string]string{
		"name": "My App",
		"slug": "my-app",
	}

	// --- Step 1: Simulate "project create" at v1 ---
	GitCheckout(templateRepo, v1sha)

	projectDir := t.TempDir()
	if err := RenderTemplateAtPath(templateRepo, projectDir, inputs); err != nil {
		t.Fatalf("failed to render v1 template: %v", err)
	}

	// Verify initial render
	assertFileContent(t, filepath.Join(projectDir, "README.md"), "# My App\nA project.\n")
	assertFileContent(t, filepath.Join(projectDir, "config.yaml"), "app: My App\nport: 8080\nhost: localhost\nworkers: 4\nlog_level: info\ndebug: false\n")
	assertFileContent(t, filepath.Join(projectDir, "src", "main.go"), "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n")
	assertFileContent(t, filepath.Join(projectDir, "docs", "guide.md"), "# Guide\nSome docs.\n")

	// --- Step 2: User customizes the project ---

	// User edits config.yaml: changes log_level (far enough from template's port change for clean merge)
	writeFile(t, filepath.Join(projectDir, "config.yaml"),
		"app: My App\nport: 8080\nhost: localhost\nworkers: 4\nlog_level: debug\ndebug: false\n")

	// User edits src/main.go: changes the println message (same area template changes → conflict)
	writeFile(t, filepath.Join(projectDir, "src", "main.go"),
		"package main\n\nfunc main() {\n\tprintln(\"hello world\")\n}\n")

	// User edits README.md (template won't change this → user change preserved)
	writeFile(t, filepath.Join(projectDir, "README.md"),
		"# My App\nA project.\n\n## Custom Section\nUser added this.\n")

	// User deletes docs/guide.md (template also deletes it → both deleted, no action)
	os.Remove(filepath.Join(projectDir, "docs", "guide.md"))

	// --- Step 3: Render OLD (v1) and NEW (v2) templates ---
	GitCheckout(templateRepo, "main")

	// Render NEW (v2) template
	theirsDir := t.TempDir()
	if err := RenderTemplateAtPath(templateRepo, theirsDir, inputs); err != nil {
		t.Fatalf("failed to render v2 template: %v", err)
	}

	// Render OLD (v1) template
	GitCheckout(templateRepo, v1sha)
	baseDir := t.TempDir()
	if err := RenderTemplateAtPath(templateRepo, baseDir, inputs); err != nil {
		t.Fatalf("failed to render v1 template: %v", err)
	}

	// --- Step 4: 3-way merge ---
	result, err := ThreeWayMerge(baseDir, projectDir, theirsDir)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	// --- Step 5: Verify merge results before applying ---
	statuses := make(map[string]MergeStatus)
	for _, f := range result.Files {
		statuses[f.RelPath] = f.Status
		t.Logf("  %s → %d", f.RelPath, f.Status)
	}

	// config.yaml: template changed port (line 2), user changed log_level (line 3) → clean merge
	if s, ok := statuses["config.yaml"]; !ok || s != MergeClean {
		t.Errorf("config.yaml: expected MergeClean (%d), got %d, present=%v", MergeClean, s, ok)
	}

	// src/main.go: template changed the body, user also changed the body → conflict
	mainPath := filepath.Join("src", "main.go")
	if s, ok := statuses[mainPath]; !ok || s != MergeConflict {
		t.Errorf("src/main.go: expected MergeConflict (%d), got %d, present=%v", MergeConflict, s, ok)
	}

	// README.md: template didn't change it → user changes preserved (MergeUnchanged, not in results)
	if s, ok := statuses["README.md"]; ok && s != MergeUnchanged {
		t.Errorf("README.md: expected MergeUnchanged or absent, got %d", s)
	}

	// Makefile: new file in template → MergeNewFile
	if s, ok := statuses["Makefile"]; !ok || s != MergeNewFile {
		t.Errorf("Makefile: expected MergeNewFile (%d), got %d, present=%v", MergeNewFile, s, ok)
	}

	// docs/guide.md: deleted in template AND deleted by user → both gone, MergeUnchanged
	guidePath := filepath.Join("docs", "guide.md")
	if s, ok := statuses[guidePath]; ok && s != MergeUnchanged {
		t.Errorf("docs/guide.md: expected absent or MergeUnchanged, got %d", s)
	}

	if !result.HasConflict {
		t.Error("expected HasConflict = true (src/main.go should conflict)")
	}

	// --- Step 6: Apply merge results ---
	if err := ApplyMerge(projectDir, baseDir, theirsDir, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}

	// --- Step 7: Verify final project state ---

	// README.md: user's custom section should be preserved exactly
	readmeContent := readFileContent(t, filepath.Join(projectDir, "README.md"))
	if !strings.Contains(readmeContent, "## Custom Section") {
		t.Error("README.md: user's custom section was lost")
	}
	if !strings.Contains(readmeContent, "User added this.") {
		t.Error("README.md: user's custom content was lost")
	}

	// config.yaml: should have BOTH user's log_level=debug AND template's port=9090 + timeout
	configContent := readFileContent(t, filepath.Join(projectDir, "config.yaml"))
	if !strings.Contains(configContent, "port: 9090") {
		t.Error("config.yaml: template's port change not applied")
	}
	if !strings.Contains(configContent, "log_level: debug") {
		t.Error("config.yaml: user's log_level change was lost")
	}
	if !strings.Contains(configContent, "timeout: 30s") {
		t.Error("config.yaml: template's new timeout line not applied")
	}

	// src/main.go: original should be untouched (user's version)
	mainContent := readFileContent(t, filepath.Join(projectDir, "src", "main.go"))
	if !strings.Contains(mainContent, "hello world") {
		t.Error("src/main.go: original file was modified (should be untouched on conflict)")
	}

	// src/main.go.sygkro-conflict should exist with conflict markers
	conflictPath := filepath.Join(projectDir, "src", "main.go.sygkro-conflict")
	if !fileExists(conflictPath) {
		t.Fatal("src/main.go.sygkro-conflict was not created")
	}
	conflictContent := readFileContent(t, conflictPath)
	if !strings.Contains(conflictContent, "<<<<<<<") {
		t.Error("conflict file missing <<<<<<< markers")
	}
	if !strings.Contains(conflictContent, ">>>>>>>") {
		t.Error("conflict file missing >>>>>>> markers")
	}
	// Both sides should be represented
	if !strings.Contains(conflictContent, "hello world") {
		t.Error("conflict file missing user's change (hello world)")
	}
	if !strings.Contains(conflictContent, "fmt.Println") {
		t.Error("conflict file missing template's change (fmt.Println)")
	}

	// Makefile: should be created with rendered content
	makeContent := readFileContent(t, filepath.Join(projectDir, "Makefile"))
	if !strings.Contains(makeContent, "go build") {
		t.Error("Makefile: not created or missing content")
	}
	if !strings.Contains(makeContent, "My App") {
		t.Error("Makefile: template variables not rendered")
	}

	// docs/guide.md: should NOT exist (both template and user deleted it)
	if fileExists(filepath.Join(projectDir, "docs", "guide.md")) {
		t.Error("docs/guide.md: should not exist (deleted by both)")
	}
}

// TestSyncIntegration_NoChanges verifies sync does nothing when template hasn't changed.
func TestSyncIntegration_NoChanges(t *testing.T) {
	templateRepo, v1sha, _ := buildTemplateRepo(t)

	inputs := map[string]string{"name": "My App", "slug": "my-app"}

	// Render project at v1
	GitCheckout(templateRepo, v1sha)
	projectDir := t.TempDir()
	if err := RenderTemplateAtPath(templateRepo, projectDir, inputs); err != nil {
		t.Fatal(err)
	}

	// User makes customizations
	writeFile(t, filepath.Join(projectDir, "README.md"),
		"# My App\nCustomized!\n")

	// Sync against the SAME version (v1 → v1) — should detect no template changes
	baseDir := t.TempDir()
	if err := RenderTemplateAtPath(templateRepo, baseDir, inputs); err != nil {
		t.Fatal(err)
	}

	theirsDir := t.TempDir()
	if err := RenderTemplateAtPath(templateRepo, theirsDir, inputs); err != nil {
		t.Fatal(err)
	}

	result, err := ThreeWayMerge(baseDir, projectDir, theirsDir)
	if err != nil {
		t.Fatal(err)
	}

	// All files should be unchanged since base == theirs
	for _, f := range result.Files {
		if f.Status != MergeUnchanged {
			t.Errorf("expected MergeUnchanged for %s, got %d", f.RelPath, f.Status)
		}
	}
	if result.HasConflict {
		t.Error("expected no conflicts when template hasn't changed")
	}

	// User's customization should be preserved
	content := readFileContent(t, filepath.Join(projectDir, "README.md"))
	if content != "# My App\nCustomized!\n" {
		t.Errorf("user customization was lost: %q", content)
	}
}

// TestSyncIntegration_FirstSync verifies behavior when there's no old version (first sync after link).
func TestSyncIntegration_FirstSync(t *testing.T) {
	templateRepo, _, _ := buildTemplateRepo(t)

	inputs := map[string]string{"name": "My App", "slug": "my-app"}

	// Project has some pre-existing files (simulating project link)
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "README.md"),
		"# My App\nExisting project readme.\n")

	// No old version — base is empty
	GitCheckout(templateRepo, "main")
	baseDir := t.TempDir() // empty

	theirsDir := t.TempDir()
	if err := RenderTemplateAtPath(templateRepo, theirsDir, inputs); err != nil {
		t.Fatal(err)
	}

	result, err := ThreeWayMerge(baseDir, projectDir, theirsDir)
	if err != nil {
		t.Fatal(err)
	}

	statuses := make(map[string]MergeStatus)
	for _, f := range result.Files {
		statuses[f.RelPath] = f.Status
	}

	// README.md exists in both project and template with different content, no base → conflict
	if s := statuses["README.md"]; s != MergeConflict {
		t.Errorf("README.md: expected MergeConflict on first sync, got %d", s)
	}

	// config.yaml only in template → new file
	if s := statuses["config.yaml"]; s != MergeNewFile {
		t.Errorf("config.yaml: expected MergeNewFile, got %d", s)
	}

	// Apply
	if err := ApplyMerge(projectDir, baseDir, theirsDir, result); err != nil {
		t.Fatal(err)
	}

	// README.md original preserved, conflict file created
	readmeContent := readFileContent(t, filepath.Join(projectDir, "README.md"))
	if !strings.Contains(readmeContent, "Existing project readme.") {
		t.Error("original README.md was modified")
	}
	if !fileExists(filepath.Join(projectDir, "README.md.sygkro-conflict")) {
		t.Error("README.md.sygkro-conflict not created")
	}

	// New files added
	if !fileExists(filepath.Join(projectDir, "config.yaml")) {
		t.Error("config.yaml should have been added")
	}
}

// TestSyncIntegration_TemplateDeletesFile verifies that files deleted in the template
// are reported but not auto-deleted from the project.
func TestSyncIntegration_TemplateDeletesFile(t *testing.T) {
	templateRepo, v1sha, _ := buildTemplateRepo(t)

	inputs := map[string]string{"name": "My App", "slug": "my-app"}

	// Render project at v1
	GitCheckout(templateRepo, v1sha)
	projectDir := t.TempDir()
	if err := RenderTemplateAtPath(templateRepo, projectDir, inputs); err != nil {
		t.Fatal(err)
	}

	// User customized the file that template will delete
	writeFile(t, filepath.Join(projectDir, "docs", "guide.md"),
		"# Guide\nUser added important notes here.\n")

	// Render base (v1) and theirs (v2)
	baseDir := t.TempDir()
	if err := RenderTemplateAtPath(templateRepo, baseDir, inputs); err != nil {
		t.Fatal(err)
	}

	GitCheckout(templateRepo, "main") // v2
	theirsDir := t.TempDir()
	if err := RenderTemplateAtPath(templateRepo, theirsDir, inputs); err != nil {
		t.Fatal(err)
	}

	result, err := ThreeWayMerge(baseDir, projectDir, theirsDir)
	if err != nil {
		t.Fatal(err)
	}

	// docs/guide.md should be reported as deleted in template
	guidePath := filepath.Join("docs", "guide.md")
	found := false
	for _, f := range result.Files {
		if f.RelPath == guidePath {
			found = true
			if f.Status != MergeDeletedFile {
				t.Errorf("docs/guide.md: expected MergeDeletedFile, got %d", f.Status)
			}
		}
	}
	if !found {
		t.Error("docs/guide.md not found in merge results")
	}

	// Apply and verify file is NOT deleted
	if err := ApplyMerge(projectDir, baseDir, theirsDir, result); err != nil {
		t.Fatal(err)
	}
	if !fileExists(filepath.Join(projectDir, "docs", "guide.md")) {
		t.Error("docs/guide.md should NOT be auto-deleted")
	}
	content := readFileContent(t, filepath.Join(projectDir, "docs", "guide.md"))
	if !strings.Contains(content, "User added important notes") {
		t.Error("user's customization to docs/guide.md was lost")
	}
}

// TestSyncIntegration_ComputeTemplateDiff verifies that ComputeTemplateDiff
// shows only template changes (old→new), not project drift.
func TestSyncIntegration_ComputeTemplateDiff(t *testing.T) {
	templateRepo, v1sha, _ := buildTemplateRepo(t)

	inputs := map[string]string{"name": "My App", "slug": "my-app"}
	syncConfig := &config.SyncConfig{Inputs: inputs}

	// Checkout v2 (HEAD) first, then diff against v1
	GitCheckout(templateRepo, "main")

	diff, err := ComputeTemplateDiff(templateRepo, v1sha, syncConfig)
	if err != nil {
		t.Fatalf("ComputeTemplateDiff failed: %v", err)
	}

	if diff == "" {
		t.Fatal("expected non-empty diff between v1 and v2")
	}

	// Diff should show template changes
	if !strings.Contains(diff, "port: 9090") {
		t.Error("diff should show port change to 9090")
	}
	if !strings.Contains(diff, "timeout: 30s") {
		t.Error("diff should show new timeout line")
	}
	if !strings.Contains(diff, "Makefile") {
		t.Error("diff should show new Makefile")
	}
	if !strings.Contains(diff, "fmt.Println") {
		t.Error("diff should show main.go import change")
	}

	// Diff should NOT contain any user customizations (there are none in the template diff)
	// This is the key difference from the old ComputeDiff behavior
}

func assertFileContent(t *testing.T, path, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	if string(content) != expected {
		t.Errorf("%s content = %q, want %q", filepath.Base(path), string(content), expected)
	}
}
