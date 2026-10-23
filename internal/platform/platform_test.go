package platform

import (
	"reflect"
	"testing"
)

func TestDetect(t *testing.T) {
	cases := []struct {
		in, id string
		valid  bool
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", "youtube", true},
		{"https://youtu.be/dQw4w9WgXcQ", "youtube", true},
		{"https://music.youtube.com/watch?v=abc", "youtube", true},
		{"youtube.com/shorts/abc", "youtube", true},
		{"  https://www.instagram.com/reel/Cxyz/  ", "instagram", true},
		{"https://open.spotify.com/track/4uLU6hMCjMI75M1A2tKUQC", "spotify", true},
		{"spotify:album:1DFixLWuPkv3KT3TnV35m3", "spotify", true},
		{"https://spotify.link/abcdef", "spotify", true},
		{"https://vm.tiktok.com/ZMabc/", "tiktok", true},
		{"https://x.com/user/status/1", "twitter", true},
		{"https://twitter.com/user/status/1", "twitter", true},
		{"https://fb.watch/abc/", "facebook", true},
		{"https://vimeo.com/123", "vimeo", true},
		{"https://soundcloud.com/a/b", "soundcloud", true},
		{"https://v.redd.it/abc", "reddit", true},
		{"https://clips.twitch.tv/Abc", "twitch", true},
		{"https://example.org/video/1", "generic", true},
		{"https://notyoutube.com/watch", "generic", true}, // suffix match must respect dot boundary
		{"not a url", "", false},
		{"ftp://example.com/file", "", false},
		{"", "", false},
		{"https://localhost/", "", false},
	}
	for _, c := range cases {
		got := Detect(c.in)
		if got.ID != c.id || got.Valid != c.valid {
			t.Errorf("Detect(%q) = {%s %v}, want {%s %v}", c.in, got.ID, got.Valid, c.id, c.valid)
		}
	}
}

func TestExpandCollection(t *testing.T) {
	yes := []string{
		"https://www.instagram.com/p/DecsQnQP5pF/",
		"https://www.instagram.com/reel/Cxyz/",
		"https://www.instagram.com/share/p/AbC/",
		"https://www.instagram.com/stories/user/123",
		"https://www.tiktok.com/@u/photo/123",
	}
	no := []string{
		"https://www.instagram.com/hirohunt_com/",
		"https://youtu.be/dQw4w9WgXcQ",
		"https://www.youtube.com/watch?v=abc&list=PLxyz",
		"https://www.tiktok.com/@u/video/123",
		"https://open.spotify.com/track/4uLU6hMCjMI75M1A2tKUQC",
	}
	for _, u := range yes {
		if !ExpandCollection(Detect(u)) {
			t.Errorf("expected expand for %s", u)
		}
	}
	for _, u := range no {
		if ExpandCollection(Detect(u)) {
			t.Errorf("did not expect expand for %s", u)
		}
	}
}

func TestDetectFlags(t *testing.T) {
	sp := Detect("https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M")
	if !sp.Matched || !sp.AudioOnly {
		t.Errorf("spotify should be matched+audioOnly: %+v", sp)
	}
	if yt := Detect("https://youtu.be/x"); yt.Matched || yt.AudioOnly {
		t.Errorf("youtube flags wrong: %+v", yt)
	}
}

func TestNormalizeStripsBidiAndWrappers(t *testing.T) {
	got, ok := Normalize("‏<https://youtu.be/abc>‎")
	if !ok || got != "https://youtu.be/abc" {
		t.Fatalf("got %q %v", got, ok)
	}
}

func TestExtractURLs(t *testing.T) {
	text := `My list:
https://youtu.be/a, https://youtu.be/a
- https://www.instagram.com/p/xyz/.
(spotify:track:4uLU6hMCjMI75M1A2tKUQC) لینک: https://vimeo.com/1،
junk text`
	want := []string{"https://youtu.be/a", "https://www.instagram.com/p/xyz/", "spotify:track:4uLU6hMCjMI75M1A2tKUQC", "https://vimeo.com/1"}
	if got := ExtractURLs(text); !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractURLs = %v, want %v", got, want)
	}
}

func TestParseSpotifyURL(t *testing.T) {
	cases := []struct{ in, kind, id string }{
		{"https://open.spotify.com/track/4uLU6hMCjMI75M1A2tKUQC?si=1", "track", "4uLU6hMCjMI75M1A2tKUQC"},
		{"https://open.spotify.com/intl-de/album/1DFixLWuPkv3KT3TnV35m3", "album", "1DFixLWuPkv3KT3TnV35m3"},
		{"spotify:playlist:37i9dQZF1DXcBWIGoYBM5M", "playlist", "37i9dQZF1DXcBWIGoYBM5M"},
	}
	for _, c := range cases {
		k, id, ok := ParseSpotifyURL(c.in)
		if !ok || k != c.kind || id != c.id {
			t.Errorf("ParseSpotifyURL(%q) = %s %s %v", c.in, k, id, ok)
		}
	}
	if _, _, ok := ParseSpotifyURL("https://open.spotify.com/"); ok {
		t.Error("expected failure for bare host")
	}
}

func TestParseSpotifyEmbedTrack(t *testing.T) {
	html := `<html><script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"state":{"data":{"entity":{
"type":"track","id":"abc","name":"Song Name","duration":215000,"isExplicit":false,
"artists":[{"name":"Artist One"},{"name":"Artist Two"}],
"releaseDate":{"isoString":"2021-05-01T00:00:00Z"},
"coverArt":{"sources":[{"url":"https://i.scdn.co/small","width":64},{"url":"https://i.scdn.co/big","width":640}]}
}}}}}}</script></html>`
	col, err := ParseSpotifyEmbed([]byte(html))
	if err != nil {
		t.Fatal(err)
	}
	if col.Kind != "track" || len(col.Tracks) != 1 {
		t.Fatalf("unexpected: %+v", col)
	}
	tr := col.Tracks[0]
	if tr.Title != "Song Name" || tr.DurationMs != 215000 || tr.CoverURL != "https://i.scdn.co/big" || !reflect.DeepEqual(tr.Artists, []string{"Artist One", "Artist Two"}) {
		t.Errorf("track = %+v", tr)
	}
	if q := MatchQuery(tr); q != "Artist One - Song Name" {
		t.Errorf("MatchQuery = %q", q)
	}
}

func TestParseSpotifyEmbedAlbum(t *testing.T) {
	html := `<script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"state":{"data":{"entity":{
"type":"album","id":"alb","name":"The Album","subtitle":"Band",
"coverArt":{"sources":[{"url":"https://i.scdn.co/c","width":300}]},
"trackList":[
 {"uri":"spotify:track:t1","title":"First","subtitle":"Band, Guest","duration":1000},
 {"uri":"spotify:track:t2","title":"Second","subtitle":"Band","duration":2000}
]}}}}}}</script>`
	col, err := ParseSpotifyEmbed([]byte(html))
	if err != nil {
		t.Fatal(err)
	}
	if col.Kind != "album" || len(col.Tracks) != 2 || col.Owner != "Band" {
		t.Fatalf("unexpected: %+v", col)
	}
	if col.Tracks[0].ID != "t1" || col.Tracks[0].Album != "The Album" || col.Tracks[0].TrackNumber != 1 ||
		!reflect.DeepEqual(col.Tracks[0].Artists, []string{"Band", "Guest"}) {
		t.Errorf("track[0] = %+v", col.Tracks[0])
	}
}
