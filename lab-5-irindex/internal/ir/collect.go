package ir

// PositiveTerms собирает лексемы, участвующие в «позитивной» части запроса для BM25:
// поддеревья под NOT не учитываются.
func PositiveTerms(n Node) []string {
	var out []string
	var walk func(Node, bool)
	walk = func(x Node, neg bool) {
		switch t := x.(type) {
		case *Term:
			if !neg {
				out = append(out, t.Lex)
			}
		case *Not:
			walk(t.Child, !neg)
		case *And:
			for _, ch := range t.Children {
				walk(ch, neg)
			}
		case *Or:
			walk(t.Left, neg)
			walk(t.Right, neg)
		case *Near:
			if !neg {
				out = append(out, t.A, t.B)
			}
		case *Adj:
			if !neg {
				out = append(out, t.A, t.B)
			}
		case *MSM:
			if !neg {
				out = append(out, t.Terms...)
			}
		case *EdgeStart:
			if !neg {
				out = append(out, t.Lex)
			}
		case *EdgeEnd:
			if !neg {
				out = append(out, t.Lex)
			}
		}
	}
	walk(n, false)
	return uniqStrings(out)
}

// AllQueryTerms — все лексемы из запроса (включая под NOT), для подсветки в документе.
func AllQueryTerms(n Node) []string {
	var out []string
	var walk func(Node)
	walk = func(x Node) {
		switch t := x.(type) {
		case *Term:
			out = append(out, t.Lex)
		case *Not:
			walk(t.Child)
		case *And:
			for _, ch := range t.Children {
				walk(ch)
			}
		case *Or:
			walk(t.Left)
			walk(t.Right)
		case *Near:
			out = append(out, t.A, t.B)
		case *Adj:
			out = append(out, t.A, t.B)
		case *MSM:
			out = append(out, t.Terms...)
		case *EdgeStart:
			out = append(out, t.Lex)
		case *EdgeEnd:
			out = append(out, t.Lex)
		}
	}
	walk(n)
	return uniqStrings(out)
}

func uniqStrings(in []string) []string {
	m := make(map[string]struct{}, len(in))
	var out []string
	for _, s := range in {
		if _, ok := m[s]; ok {
			continue
		}
		m[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
