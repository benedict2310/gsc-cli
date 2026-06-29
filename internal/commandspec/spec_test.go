package commandspec

import "testing"

func TestPublicManifestIncludesAllPublicCommandsAndSafetyMarkers(t *testing.T) {
	manifest := PublicManifest()

	if len(manifest.Commands) == 0 {
		t.Fatal("PublicManifest() returned no commands")
	}

	wantPaths := []string{
		"sites list",
		"analytics query",
		"analytics top-pages",
		"analytics top-queries",
		"doctor",
		"report overview",
		"report compare",
		"cache status",
		"sync analytics",
		"sync hourly",
		"diagnose drop",
		"monitor hourly",
		"audit indexability",
		"audit sitemap",
		"opportunities",
		"sitemaps list",
		"sitemaps get",
		"sitemaps submit",
		"sitemaps delete",
		"inspect url",
		"inspect batch",
		"version",
		"help",
		"commands",
		"command",
	}

	byPath := map[string]Command{}
	for _, command := range manifest.Commands {
		byPath[command.Path] = command
	}
	for _, path := range wantPaths {
		command, ok := byPath[path]
		if !ok {
			t.Fatalf("missing command %q", path)
		}
		if command.Summary == "" || command.Description == "" {
			t.Fatalf("command %q missing summary/description: %#v", path, command)
		}
		if len(command.OutputModes) == 0 {
			t.Fatalf("command %q missing output modes", path)
		}
	}

	for _, path := range []string{"sitemaps submit", "sitemaps delete"} {
		command := byPath[path]
		if command.Safety != SafetyWriteCapable || !command.WriteScopeRequired {
			t.Fatalf("command %q safety markers = %#v", path, command)
		}
	}

	for _, path := range wantPaths {
		if path == "sitemaps submit" || path == "sitemaps delete" {
			continue
		}
		command := byPath[path]
		if command.Safety != SafetyReadOnly || command.WriteScopeRequired {
			t.Fatalf("command %q should be read-only: %#v", path, command)
		}
	}
}

func TestOptionalSiteFlagIsRepresentedForSitesAndDoctor(t *testing.T) {
	for _, path := range []string{"sites list", "doctor"} {
		command, ok := Find(path)
		if !ok {
			t.Fatalf("%s command not found", path)
		}
		if !hasFlag(command, "site") {
			t.Fatalf("%s missing optional site flag: %#v", path, command.Flags)
		}
		if command.RequiresSite {
			t.Fatalf("%s should not require site", path)
		}
	}
}

func TestHelpPositionalsAreNotModeledAsFlags(t *testing.T) {
	command, ok := Find("help")
	if !ok {
		t.Fatal("help command not found")
	}
	if hasFlag(command, "topic") || hasFlag(command, "subcommand") {
		t.Fatalf("help positionals should not be modeled as flags: %#v", command.Flags)
	}
}

func TestVersionOutputDefaultMatchesCLI(t *testing.T) {
	command, ok := Find("version")
	if !ok {
		t.Fatal("version command not found")
	}
	for _, flag := range command.Flags {
		if flag.Name == "output" && flag.Default != "table" {
			t.Fatalf("version output default = %q, want table", flag.Default)
		}
	}
}

func hasFlag(command Command, name string) bool {
	for _, flag := range command.Flags {
		if flag.Name == name {
			return true
		}
	}
	return false
}
