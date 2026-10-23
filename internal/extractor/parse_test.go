package extractor

import (
	"strings"
	"testing"

	"github.com/amirreza-azadmehr/dlyven/internal/platform"
)

const videoFixture = `{
 "id": "abc", "title": "  Test Video  ", "uploader": "Chan", "duration": 100, "upload_date": "20240115",
 "view_count": 1234, "webpage_url": "https://www.youtube.com/watch?v=abc", "extractor_key": "Youtube",
 "thumbnail": "https://i.ytimg.com/vi/abc/maxresdefault.jpg",
 "thumbnails": [{"url": "https://i.ytimg.com/vi/abc/default.jpg", "width": 120, "height": 90},
                {"url": "https://i.ytimg.com/vi/abc/maxresdefault.jpg", "width": 1280, "height": 720}],
 "formats": [
  {"format_id": "sb0", "ext": "mhtml", "vcodec": "none", "acodec": "none", "format_note": "storyboard"},
  {"format_id": "139", "ext": "m4a", "vcodec": "none", "acodec": "mp4a.40.5", "abr": 48, "filesize": 600000},
  {"format_id": "140", "ext": "m4a", "vcodec": "none", "acodec": "mp4a.40.2", "abr": 129.5, "filesize": 1600000},
  {"format_id": "251", "ext": "webm", "vcodec": "none", "acodec": "opus", "abr": 140},
  {"format_id": "18", "ext": "mp4", "width": 640, "height": 360, "fps": 30, "vcodec": "avc1.42001E", "acodec": "mp4a.40.2", "tbr": 500},
  {"format_id": "137", "ext": "mp4", "width": 1920, "height": 1080, "fps": 30, "vcodec": "avc1.640028", "acodec": "none", "tbr": 4000, "filesize": 50000000},
  {"format_id": "248", "ext": "webm", "width": 1920, "height": 1080, "fps": 30, "vcodec": "vp9", "acodec": "none", "tbr": 3000, "filesize_approx": 37000000},
  {"format_id": "299", "ext": "mp4", "width": 1920, "height": 1080, "fps": 60, "vcodec": "avc1.64002a", "acodec": "none", "tbr": 6000},
  {"format_id": "337", "ext": "webm", "width": 3840, "height": 2160, "fps": 60, "vcodec": "vp09.02.51.10", "acodec": "none", "dynamic_range": "HDR10", "tbr": 30000},
  {"format_id": "short", "ext": "mp4", "width": 1080, "height": 1920, "vcodec": "avc1", "acodec": "none", "tbr": 2000}
 ],
 "subtitles": {"fa": [{"ext": "vtt", "name": "Persian"}, {"ext": "srv3"}], "live_chat": [{"ext": "json"}]},
 "automatic_captions": {"en": [{"ext": "vtt", "name": "English (auto)"}]}
}`

func TestParseInfoJSONVideo(t *testing.T) {
	mi, err := ParseInfoJSON([]byte(videoFixture))
	if err != nil {
		t.Fatal(err)
	}
	if mi.Title != "Test Video" || mi.UploadDate != "2024-01-15" || mi.Uploader != "Chan" || mi.IsPlaylist {
		t.Errorf("meta wrong: %+v", mi)
	}
	// groups: 2160, 1080 (incl. vertical 1080x1920), 360
	if len(mi.VideoGroups) != 3 {
		t.Fatalf("groups = %d", len(mi.VideoGroups))
	}
	if mi.VideoGroups[0].Height != 2160 || mi.MaxHeight != 2160 {
		t.Errorf("first group = %d", mi.VideoGroups[0].Height)
	}
	g1080 := mi.VideoGroups[1]
	if g1080.Height != 1080 || len(g1080.Formats) != 4 {
		t.Fatalf("1080 group = %+v", g1080)
	}
	if g1080.Formats[0].ID != "299" { // 60fps sorts first
		t.Errorf("1080 best = %s", g1080.Formats[0].ID)
	}
	f337 := mi.VideoGroups[0].Formats[0]
	if !f337.HDR || f337.VCodec != "VP9" || f337.Label != "2160p60" || f337.HasAudio {
		t.Errorf("337 = %+v", f337)
	}
	// sizes: explicit, approx, estimated from tbr*duration
	for _, f := range g1080.Formats {
		switch f.ID {
		case "137":
			if f.Filesize != 50000000 || f.SizeApprox {
				t.Errorf("137 size %+v", f)
			}
		case "248":
			if f.Filesize != 37000000 || !f.SizeApprox {
				t.Errorf("248 size %+v", f)
			}
		case "299":
			if f.Filesize != 6000*1000/8*100 || !f.SizeApprox {
				t.Errorf("299 size %+v", f)
			}
		}
	}
	if len(mi.Audio) != 3 || mi.Audio[0].ID != "251" || mi.Audio[0].ACodec != "Opus" {
		t.Errorf("audio = %+v", mi.Audio)
	}
	h := mi.Highlights
	if h.Best != "337" || h.Recommended != "299" || h.Smallest != "18" {
		t.Errorf("highlights = %+v", h)
	}
	if len(mi.Subtitles) != 2 || mi.Subtitles[0].Lang != "fa" || mi.Subtitles[0].Auto || !mi.Subtitles[1].Auto {
		t.Errorf("subs = %+v", mi.Subtitles)
	}
	if len(mi.Thumbnails) != 2 || mi.Thumbnails[0].Width != 1280 {
		t.Errorf("thumbs = %+v", mi.Thumbnails)
	}
}

func TestParseInfoJSONPlaylist(t *testing.T) {
	js := `{"_type": "playlist", "id": "PL1", "title": "My List", "uploader": "U", "webpage_url": "https://www.youtube.com/playlist?list=PL1",
	 "entries": [
	  {"_type": "url", "ie_key": "Youtube", "id": "v1", "url": "https://www.youtube.com/watch?v=v1", "title": "One", "duration": 61,
	   "thumbnails": [{"url": "https://t/1s.jpg", "width": 100}, {"url": "https://t/1m.jpg", "width": 300}, {"url": "https://t/1l.jpg", "width": 600}]},
	  {"_type": "url", "ie_key": "Youtube", "id": "v2", "url": "v2", "title": ""}
	 ]}`
	mi, err := ParseInfoJSON([]byte(js))
	if err != nil {
		t.Fatal(err)
	}
	if !mi.IsPlaylist || len(mi.Entries) != 2 {
		t.Fatalf("%+v", mi)
	}
	if mi.Entries[1].URL != "https://www.youtube.com/watch?v=v2" || mi.Entries[1].Title != "v2" || mi.Entries[1].Index != 2 {
		t.Errorf("entry 2 = %+v", mi.Entries[1])
	}
	if mi.Entries[0].Thumbnail != "https://t/1m.jpg" || mi.Thumbnail != "https://t/1m.jpg" {
		t.Errorf("thumb = %q / %q", mi.Entries[0].Thumbnail, mi.Thumbnail)
	}
}

func TestParseInfoJSONUnknownCodecs(t *testing.T) {
	// Instagram-style: progressive formats without codec info, plus an audio-only m4a.
	js := `{"id": "x", "title": "Reel", "duration": 10, "formats": [
	  {"format_id": "dash-a", "ext": "m4a", "vcodec": "none", "acodec": "mp4a.40.2"},
	  {"format_id": "8", "ext": "mp4", "width": 720, "height": 1280},
	  {"format_id": "aud", "ext": "m4a"}
	]}`
	mi, err := ParseInfoJSON([]byte(js))
	if err != nil {
		t.Fatal(err)
	}
	if len(mi.VideoGroups) != 1 || !mi.VideoGroups[0].Formats[0].HasAudio || mi.VideoGroups[0].Height != 720 {
		t.Errorf("groups = %+v", mi.VideoGroups)
	}
	if len(mi.Audio) != 2 {
		t.Errorf("audio = %+v", mi.Audio)
	}
}

func TestNormalizeCodec(t *testing.T) {
	cases := map[string]string{"avc1.640028": "H.264", "vp09.00.40.08": "VP9", "av01.0.08M.08": "AV1", "hev1.1.6.L93": "H.265", "mp4a.40.2": "AAC", "opus": "Opus", "none": "", "ec-3": "E-AC-3"}
	for in, want := range cases {
		if got := NormalizeCodec(in); got != want {
			t.Errorf("NormalizeCodec(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolutionLabel(t *testing.T) {
	if l := ResolutionLabel(1920, 1080, 59.94); l != "1080p60" {
		t.Error(l)
	}
	if l := ResolutionLabel(1080, 1920, 30); l != "1080p" {
		t.Error(l)
	}
	if l := ResolutionLabel(0, 0, 0); l != "" {
		t.Error(l)
	}
}

func TestParseInfoJSONInstagramPhoto(t *testing.T) {
	js := `{"id":"p1","title":"Photo","formats":[
	  {"format_id":"0","ext":"jpg","width":1080,"height":1350,"vcodec":"none","acodec":"none"}
	]}`
	mi, err := ParseInfoJSON([]byte(js))
	if err != nil {
		t.Fatal(err)
	}
	if len(mi.VideoGroups) != 1 || !mi.VideoGroups[0].Formats[0].HasVideo || mi.VideoGroups[0].Formats[0].HasAudio {
		t.Errorf("photo format dropped: %+v", mi.VideoGroups)
	}
}

func TestAnalyzeFlags(t *testing.T) {
	ig := analyzeFlags(platform.Detect("https://www.instagram.com/p/DecsQnQP5pF/"))
	joined := strings.Join(ig, " ")
	if !strings.Contains(joined, "--yes-playlist") || strings.Contains(joined, "--no-playlist") || !strings.Contains(joined, "--socket-timeout") {
		t.Errorf("instagram post flags: %v", ig)
	}
	yt := analyzeFlags(platform.Detect("https://youtu.be/dQw4w9WgXcQ"))
	joined = strings.Join(yt, " ")
	if !strings.Contains(joined, "--no-playlist") {
		t.Errorf("youtube flags: %v", yt)
	}
}
