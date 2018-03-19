package xshell

type Console struct {
	keep []byte
	Src  chan []byte
}

func NewConsole() *Console {
	cs := &Console{
		keep: nil,
		Src:  make(chan []byte, 256),
	}
	return cs
}

func (o *Console) Read(p []byte) (n int, err error) {
	err = nil
	if o.keep != nil {
		n = len(o.keep)
		if n > 0 {
			if n > len(p) {
				n = len(p)
				copy(p, o.keep)
				o.keep = o.keep[n:]
			} else {
				copy(p, o.keep)
				o.keep = nil
			}
		}
		return
	} else {
		for {
			o.keep = <-o.Src
			if o.keep != nil {
				break
			}
		}
		n = len(o.keep)
		if n > 0 {
			if n > len(p) {
				n = len(p)
				copy(p, o.keep)
				o.keep = o.keep[n:]
			} else {
				copy(p, o.keep)
				o.keep = nil
			}
		}
	}
	return
}

func (o *Console) Write(p []byte) (n int, err error) {
	o.Src <- p
	return len(p), nil
}
