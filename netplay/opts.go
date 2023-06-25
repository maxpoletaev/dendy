package netplay

type Options struct {
	BatchSize  int
	Ping       int
	PingJitter int
}

func withOptions(np *Netplay, opts Options) {
	if opts.BatchSize > 0 {
		np.batchSize = opts.BatchSize
	}

	if opts.Ping > 0 {
		np.ping = opts.Ping
	}

	if opts.PingJitter > 0 {
		np.pingJitter = opts.PingJitter
	}
}
