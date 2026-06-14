package ir

import (
	"os"
	"testing"
)

func TestVarintBeatsP4Examples(t *testing.T) {
	ex := TopVarintBeatsP4(fillCorpus(200, 42, defaultWords()), "synthetic", 3)
	if len(ex) == 0 {
		t.Log("synthetic: no strong examples (short corpus)")
	}
	for _, e := range ex {
		t.Logf("synthetic term=%q df=%d varint=%d B p4=%d B ratio=%.1fx",
			e.Term, e.DF, e.VarintB, e.P4B, e.RatioP4V)
	}
	if os.Getenv("WIKI_CODEC_EXAMPLES") == "" {
		return
	}
	path := ResolveCorpusPath()
	if path == "" {
		t.Skip("no wiki")
	}
	ix, _, err := BuildIndexFromWikiXML(path, CorpusOpts{MaxDocs: 20000})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range TopVarintBeatsP4(ix, "ruwiki", 5) {
		t.Logf("ruwiki term=%q df=%d varint=%d B p4=%d B ratio=%.1fx docs=%v",
			e.Term, e.DF, e.VarintB, e.P4B, e.RatioP4V, e.DocIDs)
	}
}
