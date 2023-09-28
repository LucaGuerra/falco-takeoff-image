package main

import "log"

func LogInfof(fmt string, args ...any) {
	log.Printf(fmt, args...)
}

func LogProgressf(fmt string, args ...any) {
	log.Printf(fmt, args...)
}

func LogErrorf(fmt string, args ...any) {
	log.Printf(fmt, args...)
}

func LogFatalf(fmt string, args ...any) {
	log.Fatalf(fmt, args...)
}
