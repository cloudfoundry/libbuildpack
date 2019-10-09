package docker_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/onsi/gomega/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDocker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cutlass/docker")
}

func MatchLines(expected string) types.GomegaMatcher {
	return &matchLinesMatcher{
		expected: expected,
	}
}

type matchLinesMatcher struct {
	expected string
}

func (matcher *matchLinesMatcher) Match(a interface{}) (success bool, err error) {
	_, match, err := matcher.lines(a)
	if err != nil {
		return false, err
	}

	return match, nil
}

func (matcher *matchLinesMatcher) lines(a interface{}) ([][]string, bool, error) {
	actual, ok := a.(string)
	if !ok {
		return nil, false, fmt.Errorf("MatchLines matcher expects a string, got %T", a)
	}

	expectedLines := strings.Split(strings.TrimSpace(matcher.expected), "\n")
	actualLines := strings.Split(strings.TrimSpace(actual), "\n")

	lines := [][]string{}
	match := true

	for index, expectedLine := range expectedLines {
		if index >= len(actualLines) {
			lines = append(lines, []string{expectedLine, ""})
			match = false
			continue
		}

		actualLine := actualLines[index]
		lines = append(lines, []string{expectedLine, actualLine})
		if actualLine != expectedLine {
			match = false
		}
	}

	return lines, match, nil
}

func (matcher *matchLinesMatcher) lineDiff(lines [][]string) string {
	var diff strings.Builder
	for _, line := range lines {
		expected, actual := line[0], line[1]
		if actual == expected {
			diff.WriteString(fmt.Sprintf("   %s\n", actual))
		} else {
			diff.WriteString(fmt.Sprintf("  + %s\n", actual))
			diff.WriteString(fmt.Sprintf("  - %s\n", expected))
		}
	}

	return diff.String()
}

func (matcher *matchLinesMatcher) FailureMessage(a interface{}) string {
	lines, _, _ := matcher.lines(a)

	var message strings.Builder
	message.WriteString("Expected lines to match, but found:\n")
	message.WriteString(matcher.lineDiff(lines))

	return message.String()
}

func (matcher *matchLinesMatcher) NegatedFailureMessage(a interface{}) string {
	lines, _, _ := matcher.lines(a)

	var message strings.Builder
	message.WriteString("Expected lines not to match, but found:\n")
	message.WriteString(matcher.lineDiff(lines))

	return message.String()
}
