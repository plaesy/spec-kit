// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
	"strings"
)

// YouTubeAdapter handles YouTube video transcripts and search.
type YouTubeAdapter struct {
	BaseAdapter
}

func NewYouTubeAdapter() *YouTubeAdapter {
	return &YouTubeAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformYouTube,
			backends:     []Backend{"yt-dlp", "youtube-transcript-api"},
			requiresAuth: false,
		},
	}
}

// Execute fetches YouTube data.
func (y *YouTubeAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "yt-dlp":
		return y.executeYtDlp(ctx, query, opts)
	case "youtube-transcript-api":
		return y.executeTranscriptAPI(ctx, query)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for youtube", backend)
	}
}

func (y *YouTubeAdapter) executeYtDlp(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	// Check if query is a URL or search term
	args := []string{"yt-dlp"}

	// Extract options
	format := opts["format"]
	if format == "" {
		format = "best"
	}

	subtitles := opts["subtitles"]
	if subtitles == "true" || subtitles == "1" {
		args = append(args, "--write-subs", "--sub-lang", "en,zh,auto", "--skip-download")
	}

	// If it's a search query, use ytsearch
	if !strings.Contains(query, "youtube.com") && !strings.Contains(query, "youtu.be") {
		args = append(args, "ytsearch:"+query)
	} else {
		args = append(args, query)
	}

	args = append(args, "-o", "-") // Output to stdout
	return y.runCommand(ctx, "yt-dlp", args...)
}

func (y *YouTubeAdapter) executeTranscriptAPI(ctx context.Context, query string) (BackendResult, error) {
	// This would use a Python script with youtube-transcript-api
	// For now, fallback to yt-dlp
	return y.executeYtDlp(ctx, query, map[string]string{"subtitles": "true"})
}

// Health checks if yt-dlp is available.
func (y *YouTubeAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "yt-dlp":
		return y.checkCommand(ctx, backend, "yt-dlp", "--version"), nil
	case "youtube-transcript-api":
		return y.checkCommand(ctx, backend, "python3", "-c", "import youtube_transcript_api"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}
