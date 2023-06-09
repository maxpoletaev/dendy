package netplay

type Options struct {
	BatchSize int
}

func withOptions(np *Netplay, opts Options) {
	if opts.BatchSize > 0 {
		np.batchSize = opts.BatchSize
	}
}
