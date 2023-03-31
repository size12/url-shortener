package storage

import (
	"fmt"
	"log"
	"math/rand"
	"testing"

	"github.com/size12/url-shortener/internal/config"
)

func BenchmarkDBStorage(b *testing.B) {
	var cfg = config.GetOldConfig()
	cfg.DBMigrationPath = "file://../../migrations"

	s, err := NewDBStorage(cfg)
	if err != nil {
		log.Fatalln("Failed get storage: ", err)
	}

	b.ResetTimer()
	b.Run("Add new links", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			url := fmt.Sprintf("https://random%v/random%v", rand.Intn(50000), rand.Intn(20000))
			userID := fmt.Sprint(rand.Intn(200))
			b.StartTimer()
			s.CreateShort(userID, url)
		}
	})

	b.Run("Get long urls", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			id := fmt.Sprint(rand.Intn(s.LastID))
			b.StartTimer()

			s.GetLong(id)
		}
	})

	b.Run("Delete urls", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			id := fmt.Sprint(rand.Intn(s.LastID))
			userID := fmt.Sprint(rand.Intn(200))
			b.StartTimer()

			s.Delete(userID, id)
		}
	})

	b.Run("Get history", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			userID := fmt.Sprint(rand.Intn(200))
			b.StartTimer()

			s.GetHistory(userID)
		}
	})

}

func BenchmarkMapStorage(b *testing.B) {
	var cfg = config.GetBenchConfig()

	s, err := NewMapStorage(cfg)
	if err != nil {
		log.Fatalln("Failed get storage: ", err)
	}

	b.ResetTimer()
	b.Run("Add new links", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			url := fmt.Sprintf("https://random%v/random%v", rand.Intn(5000), rand.Intn(2000))
			userID := fmt.Sprint(rand.Intn(200))
			b.StartTimer()
			s.CreateShort(userID, url)
		}
	})

	b.Run("Get long urls", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			id := fmt.Sprint(rand.Intn(len(s.Locations)))
			b.StartTimer()

			s.GetLong(id)
		}
	})

	b.Run("Delete urls", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			id := fmt.Sprint(len(s.Locations))
			userID := fmt.Sprint(rand.Intn(200))
			b.StartTimer()

			s.Delete(userID, id)
		}
	})

	b.Run("Get history", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			userID := fmt.Sprint(rand.Intn(200))
			b.StartTimer()

			s.GetHistory(userID)
		}
	})

}
