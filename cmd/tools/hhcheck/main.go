package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"mcpPrep/internal/domain"
	hhclient "mcpPrep/internal/infrastructure/hh"
)

func main() {
	ctx := context.Background()
	client := hhclient.NewClient(os.Getenv("HH_TOKEN"))

	// ── Запрос 1: поиск вакансий ──────────────────────────────────────────────
	fmt.Println("=== Запрос 1: поиск Go-вакансий в Москве ===")

	result, err := client.SearchVacancies(ctx, domain.SearchParams{
		Text:    "Go developer",
		Area:    "1",
		PerPage: 5,
	})
	if err != nil {
		log.Fatalf("search: %v", err)
	}

	fmt.Printf("Найдено всего: %d вакансий\n\n", result.Found)
	for i, v := range result.Items {
		fmt.Printf("%d. %s\n", i+1, v.Name)
		fmt.Printf("   Компания : %s\n", v.Employer.Name)
		fmt.Printf("   Город    : %s\n", v.Area.Name)
		fmt.Printf("   Зарплата : %s\n", formatSalary(v.Salary))
		fmt.Printf("   Опыт     : %s\n", v.Experience)
		fmt.Printf("   Форматы  : %s\n", strings.Join(v.WorkFormats, ", "))
		fmt.Printf("   Ссылка   : %s\n\n", v.AlternateURL)
	}

	if len(result.Items) == 0 {
		return
	}

	// ── Запрос 2: детали первой вакансии ─────────────────────────────────────
	first := result.Items[0]
	fmt.Printf("=== Запрос 2: детали вакансии #%s ===\n", first.ID)

	detail, err := client.GetVacancy(ctx, first.ID)
	if err != nil {
		log.Fatalf("get vacancy: %v", err)
	}

	fmt.Printf("Название         : %s\n", detail.Name)
	fmt.Printf("Компания         : %s (trusted: %v)\n", detail.Employer.Name, detail.Employer.Trusted)
	fmt.Printf("Навыки           : %s\n", strings.Join(detail.KeySkills, ", "))
	fmt.Printf("Профессиональные : %s\n", strings.Join(detail.ProfessionalRoles, ", "))
	fmt.Printf("Описание (200 символов):\n%s...\n", truncate(detail.Description, 200))
}

func formatSalary(s *domain.Salary) string {
	if s == nil {
		return "не указана"
	}
	var parts []string
	if s.From != nil {
		parts = append(parts, fmt.Sprintf("от %d", *s.From))
	}
	if s.To != nil {
		parts = append(parts, fmt.Sprintf("до %d", *s.To))
	}
	return strings.Join(parts, " ") + " " + s.Currency
}

func truncate(s string, n int) string {
	// убираем html теги
	s = strings.ReplaceAll(s, "<li>", "\n- ")
	for _, tag := range []string{"<p>", "</p>", "<ul>", "</ul>", "</li>", "<strong>", "</strong>"} {
		s = strings.ReplaceAll(s, tag, "")
	}
	if len(s) > n {
		return s[:n]
	}
	return s
}
