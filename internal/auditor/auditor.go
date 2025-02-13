package auditor

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketAuditScore  = []byte("AuditScore")
	bucketAuditReport = []byte("AuditReport")

	db *bolt.DB
)

func init() {
	var err error
	db, err = bolt.Open("auditor.db", 0600, nil)
	if err != nil {
		log.Fatal("Ошибка открытия bbolt DB:", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketAuditScore)
		if err != nil {
			return fmt.Errorf("не удалось создать бакет AuditScore: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists(bucketAuditReport)
		if err != nil {
			return fmt.Errorf("не удалось создать бакет AuditReport: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Fatal("Ошибка при создании бакетов:", err)
	}
}

// AuditScope хранит счетчики аудита.
type AuditScope struct {
	BananaCount int
	LikeCount   int
}

func TierByScore(score AuditScope) string {
	if score.BananaCount > 4 {
		return "A"
	}
	if score.BananaCount > 2 || score.LikeCount > 3 {
		return "B"
	}
	if score.BananaCount > 0 || score.LikeCount > 0 {
		return "C"
	}
	return "D"
}

func BakeAuditReport(score AuditScope) string {
	return fmt.Sprintf(`Сообщение прошло аудит отдела службы безопасности.

По результатам проверки, сообщение получило %v 👍 и %v 🍌.
Сообщению был присвоен %s тир хохотливости.`, score.LikeCount, score.BananaCount, TierByScore(score))
}

func StoreAuditScore(key string, value AuditScope) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Println("Ошибка маршалинга AuditScope:", err)
		return
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAuditScore)
		return b.Put([]byte(key), data)
	})
	if err != nil {
		log.Println("Ошибка сохранения AuditScope в bbolt:", err)
	}
}

func LoadAuditScore(key string) AuditScope {
	var result AuditScope
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAuditScore)
		v := b.Get([]byte(key))
		if v == nil {
			fmt.Println("Missed audit info for key:", key)
			return nil
		}
		return json.Unmarshal(v, &result)
	})
	if err != nil {
		log.Println("Ошибка загрузки AuditScope из bbolt:", err)
	}
	return result
}

func StoreAuditReport(key string, value int) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Println("Ошибка маршалинга int:", err)
		return
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAuditReport)
		return b.Put([]byte(key), data)
	})
	if err != nil {
		log.Println("Ошибка сохранения audit report в bbolt:", err)
	}
}

func LoadAuditReport(key string) int {
	var value int
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAuditReport)
		v := b.Get([]byte(key))
		if v == nil {
			fmt.Println("Missed audit report info for key:", key)
			return nil
		}
		return json.Unmarshal(v, &value)
	})
	if err != nil {
		log.Println("Ошибка загрузки audit report из bbolt:", err)
	}
	return value
}

type ScoreEntry struct {
	Key   string
	Score int
}

func GetTopAuditKeys(limit int) ([]string, error) {
	var entries []ScoreEntry

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketAuditScore)
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", bucketAuditScore)
		}

		return bucket.ForEach(func(k, v []byte) error {
			var score AuditScope
			if err := json.Unmarshal(v, &score); err != nil {
				return nil
			}
			s := score.BananaCount*10 + score.LikeCount*5
			entries = append(entries, ScoreEntry{
				Key:   string(k),
				Score: s,
			})
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})

	if limit > 0 && limit < len(entries) {
		entries = entries[:limit]
	}

	var keys []string
	for _, e := range entries {
		keys = append(keys, e.Key)
	}
	return keys, nil
}
