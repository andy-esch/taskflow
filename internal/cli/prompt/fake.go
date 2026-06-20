package prompt

import "errors"

// Fake is a scripted Prompter for tests: it pops queued answers and records the
// prompts it was asked, so prompt FLOW (which prompts fire, in what order) is
// testable in CI without a TTY — the whole reason Prompter is an interface.
type Fake struct {
	SelectAnswers []string // queued SelectOne results, popped in order
	TextAnswers   []string // queued Text results, popped in order
	Asked         []string // titles asked, in order
}

func (f *Fake) SelectOne(title string, _ []Option) (string, error) {
	f.Asked = append(f.Asked, title)
	if len(f.SelectAnswers) == 0 {
		return "", errors.New("fake prompter: no scripted SelectOne answer")
	}
	v := f.SelectAnswers[0]
	f.SelectAnswers = f.SelectAnswers[1:]
	return v, nil
}

func (f *Fake) Text(title, _ string) (string, error) {
	f.Asked = append(f.Asked, title)
	if len(f.TextAnswers) == 0 {
		return "", errors.New("fake prompter: no scripted Text answer")
	}
	v := f.TextAnswers[0]
	f.TextAnswers = f.TextAnswers[1:]
	return v, nil
}
