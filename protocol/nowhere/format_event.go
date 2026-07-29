package nowhere

import "github.com/ohmycggk/nowhere-go/diagnostic"

func formatDiagnosticEvent(event diagnostic.Event) string {
	return diagnostic.FormatEvent(event)
}
