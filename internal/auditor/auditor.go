package auditor

import (
	"fmt"
)

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
