package mediate

import "log/slog"

type Options func(*Mediate)

func WithLogger(l *slog.Logger) func(*Mediate) {
	return func(p *Mediate) {
		p.logger = l
	}
}
