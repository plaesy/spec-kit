---
name: youtube-automation
category: productivity
description: Automate YouTube video creation from topic to publish.
---

# youtube-automation

End-to-end YouTube video automation pipeline (inspired by YAAS). Researches topic, fact-checks, writes script, generates visuals, creates voiceover, renders video, optimizes title/description/tags, publishes to YouTube.

## When to Use
- User wants to create YouTube videos from a topic/idea
- Need faceless video automation (long-form 16:9 or Shorts 9:16)
- Want fact-checked content with citations
- Need thumbnail generation with brand consistency
- Want to schedule or auto-publish to YouTube

## Pipeline Stages (7 Agents)

```
topic > Research > Fact-check > Script > Visuals > Voice > Render > Review > Publish
          (trends +        (Wiki +      (AI)       (Pexels/      (Edge     (MoviePy)   (title lab,  (YouTube
          niche)          arXiv)                  Pixabay/       TTS)                  SEO score)   API)
```

### 1. Research Agent
- Sources: Google Trends, Wikipedia, Hacker News, arXiv, Spaceflight News
- Trend-aware topic suggestions per niche
- Output: research brief + source URLs

### 2. Fact-Check Agent
- Pulls background from Wikipedia + arXiv
- Instructs script AI not to contradict sources
- Lists sources in video description

### 3. Script Agent
- Generates engaging script with hook, body, CTA
- Long-form: chapters, B-roll cues
- Shorts: punchy, caption-ready
- Uses fact-check sources as grounding

### 4. Visuals Agent
- Chain: Pexels > Pixabay > NASA (public domain) > Openverse (CC) > AI-generated
- Downloads with proper credits
- Ken Burns motion, B-roll cuts for long-form
- Big captions for Shorts

### 5. Voice Agent
- Microsoft Edge neural TTS (keyless)
- Word-synced captions (SRT)
- Voice selection per channel brand

### 6. Render Agent (MoviePy)
- Long-form: 16:9, chapters, Ken Burns, watermark
- Shorts: 9:16, big captions, watermark
- Background music from library (looped)
- Output: MP4 + thumbnail + SRT

### 7. Review & Publish Agent
- Title Lab: 6 AI-suggested titles from search data, scored
- Findability Score: 11 checks (length, keywords, chapters, tags, CTA)
- Keyword Ideas: YouTube autocomplete + Datamuse
- Thumbnail Designer: 4 styles (Impact, Cinematic, Split, Studio), 16:9 + 9:16
- Upload now or schedule via YouTube Data API + OAuth

## Free API Stack (Fail-Soft)

| Service | Purpose | Key Required |
|---|---|---|
| Google Gemini | Research, scripts, titles | Recommended |
| Groq | Fast backup AI | Optional |
| OpenRouter | Backup AI (free models) | Optional |
| Pollinations (text) | Last-resort AI | No |
| Edge TTS | Neural voiceover | No |
| Pexels/Pixabay | Stock photos | Optional |
| NASA Image Library | Space imagery | No |
| Openverse | CC images (credited) | No |
| Pollinations (image) | AI B-roll/thumbnail art | No |
| Wikipedia/arXiv | Fact-checking + trends | No |
| Google Trends/HN | Trending topics | No |
| YouTube autocomplete | Keyword research | No |
| YouTube Data API | Channel stats + upload | Optional |

**Fail-soft rule**: If any service fails, retry > fallback to next > only fail if ALL options exhausted.

## Configuration

```
config.py          # All settings, documented
.env.example       # Copy to .env for API keys
.env               # Git-ignored, stores keys only
```

Non-secret settings (voice, channel name, video style) in config.py.

## Quality Gates

1. Research: at least 3 sources with URLs
2. Fact-check: All claims grounded or marked assumption
3. Script: Hook + structure + CTA present
4. Visuals: at least 5 images/clips, credits in description
5. Voice: SRT captions synced
6. Render: MP4 valid, duration matches target
7. Findability: Score at least 80/100
8. Thumbnail: Under 2MB, correct aspect ratio

## Integration with Pipeline

```
/pipeline --niche "tech education" --idea "AI tools for developers"
  > ideator > buyer-discovery > economics-engine > ...
  > youtube-automation --topic "AI tools for developers" --format long
  > Generates video + thumbnail + metadata > ready for review/publish
```

## Output Structure

```
output/
└── library/
    └── <video-id>/
        ├── video.mp4
        ├── thumbnail.jpg
        ├── captions.srt
        ├── script.md
        ├── sources.json
        ├── metadata.json
        └── findability_report.md
```

## Tests

Offline pytest suite covers:
- JSON repair logic
- Findability score calculation
- Thumbnail style rendering (4 styles x 2 ratios)
- Description/credits formatting
- Watermark application
- .env handling
- AI model chain failover

```
pip install pytest
pytest
```