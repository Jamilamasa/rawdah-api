//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type hadith struct {
	TextEn     string
	TextAr     *string
	Source     string
	Topic      string
	Difficulty string
}

func ptr(s string) *string { return &s }

var hadiths = []hadith{
	// --- Prayer & Worship ---
	{
		TextEn:     "The first matter that the slave will be brought to account for on the Day of Judgment is the prayer. If it is sound, then the rest of his deeds will be sound. And if it is corrupt, then the rest of his deeds will be corrupt.",
		Source:     "Al-Tabarani",
		Topic:      "prayer",
		Difficulty: "medium",
	},
	{
		TextEn:     "Pray as you have seen me praying.",
		Source:     "Bukhari",
		Topic:      "prayer",
		Difficulty: "easy",
	},
	{
		TextEn:     "Between a man and shirk and kufr there stands his neglect of the prayer.",
		Source:     "Muslim",
		Topic:      "prayer",
		Difficulty: "hard",
	},
	// --- Intentions ---
	{
		TextEn:     "Actions are judged by intentions, and every person will get the reward according to what he has intended.",
		TextAr:     ptr("إنما الأعمال بالنيات وإنما لكل امرئ ما نوى"),
		Source:     "Bukhari & Muslim",
		Topic:      "intentions",
		Difficulty: "easy",
	},
	// --- Honesty & Character ---
	{
		TextEn:     "Truthfulness leads to righteousness, and righteousness leads to Paradise. A man continues to tell the truth until he becomes truthful. Lying leads to wickedness, and wickedness leads to Hellfire.",
		Source:     "Bukhari & Muslim",
		Topic:      "honesty",
		Difficulty: "medium",
	},
	{
		TextEn:     "The most beloved of deeds to Allah are those that are most consistent, even if they are small.",
		Source:     "Bukhari & Muslim",
		Topic:      "worship",
		Difficulty: "easy",
	},
	{
		TextEn:     "He who does not thank people, does not thank Allah.",
		Source:     "Abu Dawud & Tirmidhi",
		Topic:      "gratitude",
		Difficulty: "easy",
	},
	{
		TextEn:     "None of you truly believes until he loves for his brother what he loves for himself.",
		TextAr:     ptr("لا يؤمن أحدكم حتى يحب لأخيه ما يحب لنفسه"),
		Source:     "Bukhari & Muslim",
		Topic:      "brotherhood",
		Difficulty: "easy",
	},
	// --- Knowledge ---
	{
		TextEn:     "Seeking knowledge is an obligation upon every Muslim.",
		TextAr:     ptr("طلب العلم فريضة على كل مسلم"),
		Source:     "Ibn Majah",
		Topic:      "knowledge",
		Difficulty: "easy",
	},
	{
		TextEn:     "Whoever follows a path in pursuit of knowledge, Allah will make a path to Paradise easy for him.",
		Source:     "Muslim",
		Topic:      "knowledge",
		Difficulty: "medium",
	},
	// --- Charity & Generosity ---
	{
		TextEn:     "The upper hand is better than the lower hand. The upper hand is the one that gives, and the lower hand is the one that receives.",
		Source:     "Bukhari & Muslim",
		Topic:      "charity",
		Difficulty: "easy",
	},
	{
		TextEn:     "Save yourself from hellfire by giving even half a date-fruit in charity.",
		Source:     "Bukhari",
		Topic:      "charity",
		Difficulty: "easy",
	},
	{
		TextEn:     "Every act of kindness is charity.",
		Source:     "Bukhari & Muslim",
		Topic:      "charity",
		Difficulty: "easy",
	},
	// --- Patience & Gratitude ---
	{
		TextEn:     "Amazing is the affair of the believer — all of it is good. If good times come his way, he is thankful, and that is good for him. If hardship comes his way, he is patient, and that is good for him.",
		Source:     "Muslim",
		Topic:      "patience",
		Difficulty: "medium",
	},
	{
		TextEn:     "There is no disease that Allah has created except that He also has created its treatment.",
		Source:     "Bukhari",
		Topic:      "patience",
		Difficulty: "medium",
	},
	// --- Family & Parents ---
	{
		TextEn:     "Paradise lies at the feet of your mother.",
		TextAr:     ptr("الجنة تحت أقدام الأمهات"),
		Source:     "Ahmad & Nasai",
		Topic:      "family",
		Difficulty: "easy",
	},
	{
		TextEn:     "The best of you are those who are best to their families, and I am the best of you to my family.",
		Source:     "Tirmidhi & Ibn Majah",
		Topic:      "family",
		Difficulty: "easy",
	},
	// --- Cleanliness ---
	{
		TextEn:     "Cleanliness is half of faith.",
		TextAr:     ptr("الطهور شطر الإيمان"),
		Source:     "Muslim",
		Topic:      "cleanliness",
		Difficulty: "easy",
	},
	// --- Manners & Speech ---
	{
		TextEn:     "Whoever believes in Allah and the Last Day should say something good or keep quiet.",
		Source:     "Bukhari & Muslim",
		Topic:      "manners",
		Difficulty: "easy",
	},
	{
		TextEn:     "Do not be angry, and Paradise will be yours.",
		Source:     "Al-Tabarani",
		Topic:      "manners",
		Difficulty: "easy",
	},
	{
		TextEn:     "Smiling at your brother is an act of charity.",
		Source:     "Tirmidhi",
		Topic:      "manners",
		Difficulty: "easy",
	},
	// --- Animals & Nature ---
	{
		TextEn:     "There is a reward for serving any living being.",
		Source:     "Bukhari & Muslim",
		Topic:      "kindness",
		Difficulty: "easy",
	},
	// --- Trust ---
	{
		TextEn:     "Return the trust to the one who entrusted you, and do not betray the one who betrayed you.",
		Source:     "Abu Dawud & Tirmidhi",
		Topic:      "trust",
		Difficulty: "medium",
	},
	// --- Remembrance of Allah ---
	{
		TextEn:     "The example of the one who remembers his Lord and the one who does not remember his Lord is like the example of the living and the dead.",
		Source:     "Bukhari",
		Topic:      "dhikr",
		Difficulty: "medium",
	},
	{
		TextEn:     "SubhanAllah, Alhamdulillah, La ilaha ill-Allah, and Allahu Akbar are more beloved to me than all that the sun rises upon.",
		Source:     "Muslim",
		Topic:      "dhikr",
		Difficulty: "easy",
	},
	// --- Quran ---
	{
		TextEn:     "The best of you are those who learn the Quran and teach it.",
		TextAr:     ptr("خيركم من تعلم القرآن وعلمه"),
		Source:     "Bukhari",
		Topic:      "quran",
		Difficulty: "easy",
	},
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM hadiths`).Scan(&count); err != nil {
		log.Fatalf("count hadiths: %v", err)
	}
	if count > 0 {
		fmt.Printf("hadiths table already has %d rows, skipping seed\n", count)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}

	inserted := 0
	for _, h := range hadiths {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO hadiths (text_en, text_ar, source, topic, difficulty)
			 VALUES ($1, $2, $3, $4, $5)`,
			h.TextEn, h.TextAr, h.Source, h.Topic, h.Difficulty,
		)
		if err != nil {
			tx.Rollback()
			log.Fatalf("insert hadith '%s': %v", h.Source, err)
		}
		inserted++
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit: %v", err)
	}

	fmt.Printf("seeded %d hadiths successfully\n", inserted)
}
