package error_manage

import "log"

func ErrorWriter(err error) {
	if err == nil {
		return
	}

	log.Println(err)
}
