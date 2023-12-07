// Package simple is a simple test for cliche. It contains a single Command with
// no tags.
package simple

import (
	"context"
	"log/slog"
)

// Tester is a cliche command which exercises default inputs.
//
//go:generate cliche -type=Tester
type Tester struct {
	// String command input.
	String string
	// Int command input.
	Int int
	// Float command input.
	Float float64
	// Boolean command input.
	Boolean bool
	// MoreStrings for the command.
	MoreStrings []string
	// MoreInts for the command.
	MoreInts []int
	// MoreFloats for the command.
	MoreFloats []float64
	// MoreBoolans for the command.
	MoreBooleans []bool
}

// Run the Tester command.
func (cmd *Tester) Run(ctx context.Context) error {
	slog.InfoContext(ctx, "Tester.Run() called with values",
		slog.String("string", cmd.String),
		slog.Int("int", cmd.Int),
		slog.Float64("float", cmd.Float),
		slog.Bool("boolean", cmd.Boolean),
		slog.Any("more_strings", cmd.MoreStrings),
		slog.Any("more_ints", cmd.MoreInts),
		slog.Any("more_floats", cmd.MoreFloats),
		slog.Any("more_booleans", cmd.MoreBooleans),
	)
	return nil
}
