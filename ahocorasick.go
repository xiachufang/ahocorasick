package goahocorasick

import (
	"fmt"

	godarts "github.com/anknown/darts"
)

const FAIL_STATE = -1
const ROOT_STATE = 1

type Machine struct {
	trie    *godarts.DoubleArrayTrie
	failure []int
	output  map[int]([][]rune)
}

type Term struct {
	Pos  int
	Word []rune
}

type SearchState struct {
	State int
	Pos   int
	Chars []rune
}

func (m *Machine) Build(keywords [][]rune) (err error) {
	if len(keywords) == 0 {
		return fmt.Errorf("empty keywords")
	}

	d := new(godarts.Darts)

	trie := new(godarts.LinkedListTrie)
	m.trie, trie, err = d.Build(keywords)
	if err != nil {
		return err
	}

	m.output = make(map[int]([][]rune), 0)
	for idx, val := range d.Output {
		m.output[idx] = append(m.output[idx], val)
	}

	queue := make([](*godarts.LinkedListTrieNode), 0)
	m.failure = make([]int, len(m.trie.Base))
	for _, c := range trie.Root.Children {
		m.failure[c.Base] = godarts.ROOT_NODE_BASE
	}
	queue = append(queue, trie.Root.Children...)

	for {
		if len(queue) == 0 {
			break
		}

		node := queue[0]
		for _, n := range node.Children {
			if n.Base == godarts.END_NODE_BASE {
				continue
			}
			inState := m.f(node.Base)
		set_state:
			outState := m.g(inState, n.Code-godarts.ROOT_NODE_BASE)
			if outState == FAIL_STATE {
				inState = m.f(inState)
				goto set_state
			}
			if _, ok := m.output[outState]; ok != false {
				copyOutState := make([][]rune, 0)
				for _, o := range m.output[outState] {
					copyOutState = append(copyOutState, o)
				}
				m.output[n.Base] = append(copyOutState, m.output[n.Base]...)
			}
			m.setF(n.Base, outState)
		}
		queue = append(queue, node.Children...)
		queue = queue[1:]
	}

	return nil
}

func (m *Machine) PrintFailure() {
	fmt.Printf("+-----+-----+\n")
	fmt.Printf("|%5s|%5s|\n", "index", "value")
	fmt.Printf("+-----+-----+\n")
	for i, v := range m.failure {
		fmt.Printf("|%5d|%5d|\n", i, v)
	}
	fmt.Printf("+-----+-----+\n")
}

func (m *Machine) PrintOutput() {
	fmt.Printf("+-----+----------+\n")
	fmt.Printf("|%5s|%10s|\n", "index", "value")
	fmt.Printf("+-----+----------+\n")
	for i, v := range m.output {
		var val string
		for _, o := range v {
			val = val + " " + string(o)
		}
		fmt.Printf("|%5d|%10s|\n", i, val)
	}
	fmt.Printf("+-----+----------+\n")
}

func (m *Machine) g(inState int, input rune) (outState int) {
	if inState == FAIL_STATE {
		return ROOT_STATE
	}

	t := inState + int(input) + godarts.ROOT_NODE_BASE
	if t >= len(m.trie.Base) {
		if inState == ROOT_STATE {
			return ROOT_STATE
		}
		return FAIL_STATE
	}
	if inState == m.trie.Check[t] {
		return m.trie.Base[t]
	}

	if inState == ROOT_STATE {
		return ROOT_STATE
	}

	return FAIL_STATE
}

func (m *Machine) f(index int) (state int) {
	return m.failure[index]
}

func (m *Machine) setF(inState, outState int) {
	m.failure[inState] = outState
}

func (m *Machine) MultiPatternSearch(
	content []rune,
	returnImmediately bool,
	NoncontinueChars int,
) [](*Term) {
	terms := make([](*Term), 0)
	var latestStates [](*SearchState)
	for pos, c := range content {
		var searchStates [](*SearchState)
		var newLatestStates [](*SearchState)
		latestStates = append(latestStates, &SearchState{State: ROOT_STATE, Pos: pos})
		for _, State := range latestStates {
			state := State.State
		start:
			if m.g(state, c) == FAIL_STATE {
				state = m.f(state)
				goto start
			}
			state = m.g(state, c)
			State.Chars = append(State.Chars, c)
			if state != ROOT_STATE {
				searchStates = append(searchStates, &SearchState{State: state, Pos: pos, Chars: State.Chars})
				newLatestStates = append(newLatestStates, &SearchState{State: state, Pos: pos, Chars: State.Chars})
			}
			if State.Pos < pos && pos-State.Pos <= NoncontinueChars {
				newLatestStates = append(newLatestStates, State)
			}
		}
		latestStates = newLatestStates
		if len(latestStates) > 100 {
			returnImmediately = true
		}
		for _, searchState := range searchStates {
			if val, ok := m.output[searchState.State]; ok != false {
				for _, word := range val {
					term := new(Term)
					term.Pos = searchState.Pos - len(searchState.Chars) + 1
					term.Word = word
					terms = append(terms, term)
					if returnImmediately {
						return terms
					}
				}
			}
		}
	}
	return terms
}

func (m *Machine) ExactSearch(content []rune) [](*Term) {
	if m.trie.ExactMatchSearch(content, 0) {
		t := new(Term)
		t.Word = content
		t.Pos = 0
		return [](*Term){t}
	}

	return nil
}
