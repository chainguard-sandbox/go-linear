package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RemovedUpstreamIn names the Linear SDK release whose API schema dropped the
// features handled by this file.
const RemovedUpstreamIn = "@linear/sdk 103.0.0"

// ErrRemovedUpstream reports that Linear removed what from its API. hint, if
// not empty, tells the user what to do instead.
func ErrRemovedUpstream(what, hint string) error {
	msg := fmt.Sprintf("%s was removed from the Linear API (%s)", what, RemovedUpstreamIn)
	if hint != "" {
		msg += ": " + hint
	}
	return fmt.Errorf("%s", msg)
}

// HideRemovedFlags marks flags whose API Linear removed as hidden. They still
// parse, so RejectRemovedFlags can explain the removal instead of cobra
// failing with "unknown flag", and the MCP server leaves hidden flags out.
func HideRemovedFlags(cmd *cobra.Command, names ...string) {
	for _, name := range names {
		if f := cmd.Flags().Lookup(name); f != nil {
			f.Hidden = true
			f.Usage = "Removed: Linear dropped this from its API (" + RemovedUpstreamIn + ")"
		}
	}
}

// RejectRemovedFlags returns ErrRemovedUpstream for the first of names that
// was passed. Call it before any API request.
func RejectRemovedFlags(cmd *cobra.Command, hint string, names ...string) error {
	for _, name := range names {
		if cmd.Flags().Changed(name) {
			return ErrRemovedUpstream("--"+name, hint)
		}
	}
	return nil
}
