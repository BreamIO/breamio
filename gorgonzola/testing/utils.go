package testing

import . "github.com/maxnordlund/breamio/gorgonzola"

func CheckError(realCh <-chan struct{}, errCh <-chan Error) error {
	select {
		case <-realCh: return nil
		case err := <-errCh: return err
	}
}