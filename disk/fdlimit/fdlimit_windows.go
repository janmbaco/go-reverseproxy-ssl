package fdlimit

// Current retrieves the number of file descriptors allowed to be opened by this
// process.
func Get() (int, error) {
	// Please see Raise for the reason why we use hard coded 16K as the limit
	return 16384, nil
}
