package analyze

import (
	"encoding/json"
	"strings"
	"testing"
)

const universalSmallProjectRec = "Expand project with additional features and modules"

// TestGenerateAIInsightsProjectTypes walks every branch of the project-type
// switch and asserts the three invariants that hold for all of them: a
// non-generic overview, a non-empty recommendation list, and no empty strings.
func TestGenerateAIInsightsProjectTypes(t *testing.T) {
	cases := []struct {
		projectType string
		wantRecs    int // recommendations excluding the small-project universal one
		wantSnippet string
	}{
		{"nextjs", 3, "Next.js application"},
		{"react", 3, "React application"},
		{"vue", 3, "Vue.js application"},
		{"nuxt", 3, "Nuxt.js application"},
		{"svelte", 3, "Svelte application"},
		{"astro", 3, "Astro application"},
		{"remix", 3, "Remix web framework"},
		{"gatsby", 3, "Gatsby static site generator"},
		{"express", 3, "Express.js web application"},
		{"nestjs", 3, "NestJS application"},
		{"vite", 3, "Vite-powered development environment"},
		{"django", 3, "Django web application"},
		{"fastapi", 3, "FastAPI application"},
		{"flask", 3, "Flask web application"},
		{"poetry", 3, "Python project managed with Poetry"},
		{"pipenv", 2, "Python project with Pipenv"},
		{"flutter", 3, "Flutter application"},
		{"gin", 3, "Gin web framework for Go"},
		{"echo", 3, "Echo web framework for Go"},
		{"fiber", 3, "Fiber web framework for Go"},
		{"actix", 3, "Actix Web framework for Rust"},
		{"rocket", 3, "Rocket web framework for Rust"},
		{"axum", 3, "Axum web framework for Rust"},
		{"rust", 3, "Rust application"},
		{"springboot", 3, "Spring Boot application"},
		{"spring", 3, "Spring framework application"},
		{"maven", 3, "Java project managed with Maven"},
		{"gradle", 3, "Java project managed with Gradle"},
		{"ktor", 3, "Ktor framework for Kotlin"},
		{"rails", 3, "Ruby on Rails application"},
		{"sinatra", 3, "Sinatra DSL for Ruby"},
		{"laravel", 3, "Laravel PHP framework"},
		{"symfony", 3, "Symfony PHP framework"},
		{"wordpress", 3, "WordPress content management system"},
		{"aspnet", 3, "ASP.NET Core application"},
		{"dotnet", 3, ".NET application"},
		{"swift", 3, "Swift application"},
		{"docker", 3, "Docker containerized application"},
		{"docker-compose", 3, "Docker Compose application"},
		{"terraform", 3, "Terraform infrastructure as code"},
		{"vagrant", 3, "Vagrant development environment"},
		{"kustomize", 3, "Kustomize Kubernetes configuration"},
		{"spec-kit", 1, "Plaesy Spec-Kit framework for AI-assisted development"},
		{"generic", 3, "Generic project structure"},
		{"unknown-framework", 3, "Generic project structure"},
		{"", 3, "Generic project structure"},
	}
	for _, tc := range cases {
		t.Run(tc.projectType, func(t *testing.T) {
			got := generateAIInsights(tc.projectType, 100)
			if !strings.HasPrefix(got.Overview, tc.wantSnippet) {
				t.Errorf("overview = %q, want prefix %q", got.Overview, tc.wantSnippet)
			}
			if len(got.Recommendations) != tc.wantRecs {
				t.Errorf("recommendations = %v, want %d entries", got.Recommendations, tc.wantRecs)
			}
			for i, r := range got.Recommendations {
				if strings.TrimSpace(r) == "" {
					t.Errorf("recommendation %d is empty", i)
				}
			}
			if hasString(got.Recommendations, universalSmallProjectRec) {
				t.Errorf("a project with 100 files must not get the small-project recommendation: %v", got.Recommendations)
			}
		})
	}
}

func TestGenerateAIInsightsSmallProjectAddsUniversalRecommendation(t *testing.T) {
	cases := []struct {
		name        string
		projectType string
		totalFiles  int
		wantPresent bool
	}{
		{"zero files", "go", 0, true},
		{"nine files still counts as small", "go", 9, true},
		{"ten files is the boundary: not small", "go", 10, false},
		{"eleven files is not small", "go", 11, false},
		{"small django project", "django", 1, true},
		{"small generic project", "generic", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := generateAIInsights(tc.projectType, tc.totalFiles)
			if present := hasString(got.Recommendations, universalSmallProjectRec); present != tc.wantPresent {
				t.Errorf("universal small-project rec present = %v, want %v (recs: %v)",
					present, tc.wantPresent, got.Recommendations)
			}
			if tc.wantPresent {
				// It is always appended last.
				if last := got.Recommendations[len(got.Recommendations)-1]; last != universalSmallProjectRec {
					t.Errorf("last recommendation = %q, want %q", last, universalSmallProjectRec)
				}
			}
		})
	}
}

func TestGenerateAIInsightsDefaultRecommendations(t *testing.T) {
	got := generateAIInsights("something-unheard-of", 50)
	want := []string{
		"Add README.md with project documentation",
		"Consider adding automated testing",
		"Implement proper version control practices",
	}
	if len(got.Recommendations) != len(want) {
		t.Fatalf("recommendations = %v, want %v", got.Recommendations, want)
	}
	for i := range want {
		if got.Recommendations[i] != want[i] {
			t.Errorf("recommendation %d = %q, want %q", i, got.Recommendations[i], want[i])
		}
	}
}

func TestGenerateAIInsightsSpecKitHasSingleRecommendation(t *testing.T) {
	got := generateAIInsights("spec-kit", 1000)
	if len(got.Recommendations) != 1 || got.Recommendations[0] != "Add more AI platform integrations" {
		t.Errorf("spec-kit recommendations = %v", got.Recommendations)
	}
}

func TestAIInsightsJSONShape(t *testing.T) {
	data, err := json.Marshal(generateAIInsights("gin", 42))
	if err != nil {
		t.Fatalf("marshal aiInsights: %v", err)
	}
	got := string(data)
	mustContain(t, got, `"overview":"Gin web framework for Go`)
	mustContain(t, got, `"recommendations":[`)
	// The small-project recommendation must not appear for a 42-file project.
	mustNotContain(t, got, universalSmallProjectRec)
}
