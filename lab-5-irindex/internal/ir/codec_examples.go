package ir

import (
	"fmt"
	"os"
	"sort"
)

// CodecExample — сравнение doc Δ: varint vs PForDelta на одном posting list.
type CodecExample struct {
	Corpus   string
	Term     string
	DF       int
	DocIDs   []uint32
	VarintB  int
	P4B      int
	RatioP4V float64 // P4 / varint (>1 ⇒ varint лучше)
}

func docStreamBytesVarint(ps []posting) int {
	vals := docDeltas(ps)
	if len(vals) == 0 {
		return 0
	}
	return len(encodeVarintStream(vals))
}

// TopVarintBeatsP4 находит posting lists, где varint на doc Δ сильно компактнее PForDelta.
func TopVarintBeatsP4(ix *InvIndex, corpus string, limit int) []CodecExample {
	type pair struct {
		r  float64
		ex CodecExample
	}
	var all []pair
	for term, ps := range ix.postings {
		if len(ps) == 0 {
			continue
		}
		vb := docStreamBytesVarint(ps)
		p4b := docStreamBytesP4(ps)
		if vb == 0 {
			continue
		}
		r := float64(p4b) / float64(vb)
		if r < 1.5 {
			continue
		}
		ids := make([]uint32, len(ps))
		for i, p := range ps {
			ids[i] = p.DocID
		}
		if len(ids) > 4 {
			ids = ids[:4]
		}
		all = append(all, pair{r: r, ex: CodecExample{
			Corpus: corpus, Term: term, DF: len(ps), DocIDs: ids,
			VarintB: vb, P4B: p4b, RatioP4V: r,
		}})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].r > all[j].r })
	if limit > len(all) {
		limit = len(all)
	}
	out := make([]CodecExample, limit)
	for i := 0; i < limit; i++ {
		out[i] = all[i].ex
	}
	return out
}

// WriteCodecExamplesTSV — ruwiki/synthetic примеры для отчёта.
func WriteCodecExamplesTSV(outPath string) error {
	var rows []CodecExample
	// Синтетика: в fillCorpus все термы повторяются → добавляем df=1 вручную.
	ixSyn := fillCorpus(200, 4242, defaultWords())
	ixSyn.Add(Tokenize("демо2 уникальный терм для сравнения кодеков"))
	rows = append(rows, TopVarintBeatsP4(ixSyn, "synthetic", 3)...)
	path := ResolveCorpusPath()
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			ix, _, err := BuildIndexFromWikiXML(path, CorpusOpts{MaxDocs: 20000})
			if err != nil {
				return err
			}
			rows = append(rows, TopVarintBeatsP4(ix, "ruwiki", 10)...)
		}
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintln(f, "corpus\tterm\tdf\tvarint_doc_B\tp4_doc_B\tratio_p4_over_varint\tdoc_ids_sample")
	for _, ex := range rows {
		fmt.Fprintf(f, "%s\t%s\t%d\t%d\t%d\t%.1f\t%v\n",
			ex.Corpus, ex.Term, ex.DF, ex.VarintB, ex.P4B, ex.RatioP4V, ex.DocIDs)
	}
	return nil
}
