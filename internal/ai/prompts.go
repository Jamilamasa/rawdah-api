package ai

import (
	"fmt"

	"github.com/rawdah/rawdah-api/internal/models"
)

// BuildHadithPrompt creates a quiz generation prompt for a hadith.
func BuildHadithPrompt(hadith models.Hadith, childAge int) string {
	ageContext := ""
	if childAge > 0 {
		ageContext = fmt.Sprintf(" The child is %d years old, so calibrate language and difficulty appropriately.", childAge)
	}

	textAr := ""
	if hadith.TextAr != nil {
		textAr = fmt.Sprintf("\nArabic: %s", *hadith.TextAr)
	}

	topic := ""
	if hadith.Topic != nil {
		topic = fmt.Sprintf("\nTopic: %s", *hadith.Topic)
	}

	return fmt.Sprintf(`You are an Islamic education assistant creating a quiz for a Muslim child learning about Hadith.%s

Create exactly 3 multiple-choice questions based on this hadith:

English: %s%s
Source: %s%s
Difficulty: %s

Requirements:
- Each question must test understanding of the hadith's meaning, context, or application in daily life
- Questions should be age-appropriate and encouraging
- Each question must have exactly 4 options labeled A, B, C, D
- Provide a brief, educational explanation for the correct answer

Respond ONLY with a valid JSON array (no markdown, no code blocks) in this exact format:
[
  {
    "id": "q1",
    "question": "Question text here?",
    "options": {"A": "Option A", "B": "Option B", "C": "Option C", "D": "Option D"},
    "correct_answer": "A",
    "explanation": "Brief explanation why A is correct."
  }
]`,
		ageContext,
		hadith.TextEn,
		textAr,
		hadith.Source,
		topic,
		hadith.Difficulty,
	)
}

// BuildProphetPrompt creates a quiz generation prompt for a prophet's story.
func BuildProphetPrompt(prophet models.Prophet, childAge int) string {
	ageContext := ""
	if childAge > 0 {
		ageContext = fmt.Sprintf(" The child is %d years old, so calibrate language and difficulty appropriately.", childAge)
	}

	nameAr := ""
	if prophet.NameAr != nil {
		nameAr = fmt.Sprintf(" (%s)", *prophet.NameAr)
	}

	miracles := ""
	if prophet.KeyMiracles != nil {
		miracles = fmt.Sprintf("\nKey Miracles: %s", *prophet.KeyMiracles)
	}

	nation := ""
	if prophet.Nation != nil {
		nation = fmt.Sprintf("\nNation: %s", *prophet.Nation)
	}

	quranRefs := ""
	if prophet.QuranRefs != nil {
		quranRefs = fmt.Sprintf("\nQuran References: %s", *prophet.QuranRefs)
	}

	return fmt.Sprintf(`You are an Islamic education assistant creating a quiz for a Muslim child learning about the Prophets.%s

Create exactly 3 multiple-choice questions based on Prophet %s%s:

Story Summary: %s%s%s%s
Difficulty: %s

Requirements:
- Questions should cover the prophet's story, character, miracles, and lessons
- Questions should inspire love for the prophet and Islamic history
- Each question must have exactly 4 options labeled A, B, C, D
- Provide educational and encouraging explanations

Respond ONLY with a valid JSON array (no markdown, no code blocks) in this exact format:
[
  {
    "id": "q1",
    "question": "Question text here?",
    "options": {"A": "Option A", "B": "Option B", "C": "Option C", "D": "Option D"},
    "correct_answer": "A",
    "explanation": "Brief explanation why A is correct."
  }
]`,
		ageContext,
		prophet.NameEn,
		nameAr,
		prophet.StorySummary,
		miracles,
		nation,
		quranRefs,
		prophet.Difficulty,
	)
}

// BuildQuranPrompt creates a quiz generation prompt for a Quran verse.
func BuildQuranPrompt(verse models.QuranVerse, childAge int) string {
	ageContext := ""
	if childAge > 0 {
		ageContext = fmt.Sprintf(" The child is %d years old, so calibrate language and difficulty appropriately.", childAge)
	}

	transliteration := ""
	if verse.Transliteration != nil {
		transliteration = fmt.Sprintf("\nTransliteration: %s", *verse.Transliteration)
	}

	lifeApp := ""
	if verse.LifeApplication != nil {
		lifeApp = fmt.Sprintf("\nLife Application: %s", *verse.LifeApplication)
	}

	topic := ""
	if verse.Topic != nil {
		topic = fmt.Sprintf("\nTopic: %s", *verse.Topic)
	}

	return fmt.Sprintf(`You are an Islamic education assistant creating a quiz for a Muslim child learning about the Quran.%s

Create exactly 3 multiple-choice questions based on this Quran verse:

Surah: %s (Chapter %d, Verse %d)
Arabic: %s%s
English Translation: %s
Simple Tafsir: %s%s%s
Difficulty: %s

Requirements:
- Questions should test understanding of the verse's meaning, the tafsir, and its application in life
- One question should be about the Arabic word meaning or the translation
- Questions should nurture love for the Quran
- Each question must have exactly 4 options labeled A, B, C, D
- Provide clear, educational explanations

Respond ONLY with a valid JSON array (no markdown, no code blocks) in this exact format:
[
  {
    "id": "q1",
    "question": "Question text here?",
    "options": {"A": "Option A", "B": "Option B", "C": "Option C", "D": "Option D"},
    "correct_answer": "A",
    "explanation": "Brief explanation why A is correct."
  }
]`,
		ageContext,
		verse.SurahNameEn,
		verse.SurahNumber,
		verse.AyahNumber,
		verse.TextAr,
		transliteration,
		verse.TextEn,
		verse.TafsirSimple,
		lifeApp,
		topic,
		verse.Difficulty,
	)
}
