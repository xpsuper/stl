package gen

import (
	"log"
)

func c(err error) {
	if err != nil {
		panic(err)
		log.Fatal(err)
	}
}

func p(v ...interface{}) {
	log.Println(v...)
}
