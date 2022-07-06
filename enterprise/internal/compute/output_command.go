package compute

import (
	"context"
	"fmt"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type Output struct {
	Filters       []ComputeFilter
	SearchPattern MatchPattern
	OutputPattern string
	Separator     string
	Selector      string
	TypeValue     string
	Kind          string
}

func (c *Output) ToSearchPattern() string {
	return c.SearchPattern.String()
}

func (c *Output) String() string {
	return fmt.Sprintf("Output with separator: (%s) -> (%s) separator: %s", c.SearchPattern.String(), c.OutputPattern, c.Separator)
}

func substituteRegexp(content string, match *regexp.Regexp, replacePattern, separator string) string {
	var b strings.Builder
	for _, submatches := range match.FindAllStringSubmatchIndex(content, -1) {
		b.Write(match.ExpandString([]byte{}, replacePattern, content, submatches))
		b.WriteString(separator)
	}
	return b.String()
}

func output(ctx context.Context, fragment string, matchPattern MatchPattern, replacePattern string, separator string) (*Text, error) {
	var newContent string
	var err error
	switch match := matchPattern.(type) {
	case *Regexp:
		newContent = substituteRegexp(fragment, match.Value, replacePattern, separator)
	case *Comby:
		newContent, err = comby.Outputs(ctx, comby.Args{
			Input:           comby.FileContent(fragment),
			MatchTemplate:   match.Value,
			RewriteTemplate: replacePattern,
			Matcher:         ".generic", // TODO(rvantoner): use language or file filter
			ResultKind:      comby.NewlineSeparatedOutput,
			NumWorkers:      0,
		})
		if err != nil {
			return nil, err
		}

	}
	return &Text{Value: newContent, Kind: "output"}, nil
}

func resultContent(ctx context.Context, db database.DB, r result.Match, onlyPath bool) (string, bool, error) {
	switch m := r.(type) {
	case *result.RepoMatch:
		return string(m.Name), true, nil
	case *result.FileMatch:
		if onlyPath {
			return m.Path, true, nil
		}
		contentBytes, err := gitserver.NewClient(db).ReadFile(ctx, m.Repo.Name, m.CommitID, m.Path, authz.DefaultSubRepoPermsChecker)
		if err != nil {
			return "", false, err
		}
		return string(contentBytes), true, nil
	case *result.CommitDiffMatch:
		var sb strings.Builder
		for _, h := range m.Hunks {
			for _, l := range h.Lines {
				sb.WriteString(l)
			}
		}
		return sb.String(), true, nil
	case *result.CommitMatch:
		var content string
		if m.DiffPreview != nil {
			content = m.DiffPreview.Content
		} else {
			content = string(m.Commit.Message)
		}
		return content, true, nil
	default:
		return "", false, nil
	}
}

func toTextResult(ctx context.Context, content string, matchPattern MatchPattern, outputPattern, separator, selector string) (Result, error) {
	if selector != "" {
		// Don't run the search pattern over the search result content
		// when there's an explicit `select:` value.
		return &Text{Value: outputPattern, Kind: "output"}, nil
	}

	return output(ctx, content, matchPattern, outputPattern, separator)
}

func toTextExtraResult(text *Text, r result.Match) *TextExtra {
	return &TextExtra{
		Text:         *text,
		RepositoryID: int32(r.RepoName().ID),
		Repository:   string(r.RepoName().Name),
	}
}

func (c *Output) Run(ctx context.Context, db database.DB, r result.Match) (Result, error) {
	onlyPath := c.TypeValue == "path" // don't read file contents for file matches when we only want type:path
	content, ok, err := resultContent(ctx, db, r, onlyPath)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	log15.Info("before filters")

	for _, filter := range c.Filters {
		pattern, err := regexp.Compile(filter.pattern)
		if err != nil {
			return nil, err
		}
		if !pattern.MatchString(content) && !filter.negated {
			// skip this
			log15.Info("filtering not negated not matching", "pattern", filter.pattern)
			return nil, nil
		} else if pattern.MatchString(content) && filter.negated {
			// skip this
			log15.Info("filtering negated and matching", "pattern", filter.pattern)
			return nil, nil
		}
	}

	log15.Info("gogogoogogog")

	env := NewMetaEnvironment(r, content)
	outputPattern, err := substituteMetaVariables(c.OutputPattern, env)
	if err != nil {
		return nil, err
	}

	result, err := toTextResult(ctx, content, c.SearchPattern, outputPattern, c.Separator, c.Selector)
	if err != nil {
		return nil, err
	}

	switch c.Kind {
	case "output.extra":
		return toTextExtraResult(result.(*Text), r), nil
	default:
		return result, nil
	}
}
