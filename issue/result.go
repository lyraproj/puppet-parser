package issue

type (
	Result interface {
		// Error returns true if errors where found or false if this result
		// contains only warnings.
		Error() bool

		// Issues returns all errors and warnings. An empty slice is returned
		// when no errors or warnings were generated
		Issues() []*Reported
	}

	parseResult struct {
		issues []*Reported
	}
)

func NewResult(issues []*Reported) Result {
	return &parseResult{issues}
}

func (pr *parseResult) Error() bool {
	for _, i := range pr.issues {
		if i.Severity() == SEVERITY_ERROR {
			return true
		}
	}
	return false
}

func (pr *parseResult) Issues() []*Reported {
	return pr.issues
}
