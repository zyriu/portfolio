package misc

import "log"

func Die(err error) {
	if err != nil {
		log.Fatalf("%v", err)
	}
}
