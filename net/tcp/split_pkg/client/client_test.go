package main

import "testing"

func BenchmarkClient(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			server := "127.0.0.1:9988"

			sender(server)
		}
	})
}
